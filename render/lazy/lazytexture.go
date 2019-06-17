package lazy

import (
	"strings"

	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/resource"
	"github.com/emily33901/lambda-core/core/texture"
	"github.com/emily33901/vtf"
)

type TextureLazy2D struct {
	filePath   string
	fileSystem filesystem.IFileSystem
	width      int
	height     int
	vtf        *vtf.Vtf
}

// FilePath Get the filepath this data was loaded from
func (tex *TextureLazy2D) FilePath() string {
	return tex.filePath
}

// Width returns materials width
func (tex *TextureLazy2D) Width() int {
	return tex.width
}

// Height returns materials height
func (tex *TextureLazy2D) Height() int {
	return tex.height
}

// Format returns this materials colour format
func (tex *TextureLazy2D) Format() uint32 {
	if tex.vtf == nil {
		logger.Panic("Always Reload() a texture before attempting to access its fields")
	}

	return tex.vtf.Header().HighResImageFormat
}

// PixelDataForFrame get raw colour data for this frame
func (tex *TextureLazy2D) PixelDataForFrame(frame int) []byte {
	if tex.vtf == nil {
		logger.Panic("Always Reload() a texture before attempting to access its fields")
	}

	return tex.vtf.HighestResolutionImageForFrame(frame)
}

// Thumbnail returns a small thumbnail image of a material
func (tex *TextureLazy2D) Thumbnail() []byte {
	if tex.vtf == nil {
		logger.Panic("Always Reload() a texture before attempting to access its fields")
	}

	return tex.vtf.LowResImageData()
}

func (tex *TextureLazy2D) Reload() error {
	stream, err := tex.fileSystem.GetFile(tex.filePath)
	if err != nil {
		logger.Error("Unable to load %s from Disk: %s", tex.filePath, err)
		return err
	}

	// Attempt to parse the vtf into color data we can use,
	// if this fails (it shouldn't) we can treat it like it was missing
	read, err := vtf.ReadFromStream(stream)
	if err != nil {
		logger.Error("Unable to load %s from Disk: %s", tex.filePath, err)
		return err
	}

	tex.vtf = read
	return nil
}

func (tex *TextureLazy2D) EvictFromMainMemory() {
	// This will trigger gc to evict this memory
	tex.vtf = nil
}

// NewTexture2D returns a new texture from Vtf
func NewLazyTexture(filePath string, fs filesystem.IFileSystem, width int, height int) *TextureLazy2D {
	// @TODO: we should be able to load the vtf all by ourselves!
	return &TextureLazy2D{
		fileSystem: fs,
		filePath:   filePath,
		width:      width,
		height:     height,
	}
}

func LoadLazyTexture(filePath string, fs filesystem.IFileSystem) texture.ITexture {
	filePath = filesystem.NormalisePath(filePath)
	if !strings.HasSuffix(filePath, filesystem.ExtensionVtf) {
		filePath = filePath + filesystem.ExtensionVtf
	}
	if resource.Manager().Texture(filesystem.BasePathMaterial+filePath) != nil {
		return resource.Manager().Texture(filesystem.BasePathMaterial + filePath).(texture.ITexture)
	}
	if filePath == "" {
		return resource.Manager().Texture(resource.Manager().ErrorTextureName()).(texture.ITexture)
	}
	mat, err := readVtf(filesystem.BasePathMaterial+filePath, fs)
	if err != nil {
		logger.Warn("Failed to load texture: %s. Reason: %s", filesystem.BasePathMaterial+filePath, err)
		return resource.Manager().Texture(resource.Manager().ErrorTextureName()).(texture.ITexture)
	}
	return mat
}

// @TODO: these should be exposed from texture.go so I dont have to paste them here!

// readVtf
func readVtf(path string, fs filesystem.IFileSystem) (texture.ITexture, error) {
	ResourceManager := resource.Manager()
	stream, err := fs.GetFile(path)
	if err != nil {
		return nil, err
	}

	// @TODO this will load the WHOLE texture from the disk
	// when we only really want the header!

	header, err := vtf.ReadHeaderFromStream(stream)
	if err != nil {
		return nil, err
	}

	// Store filesystem containing raw data in memory
	ResourceManager.AddTexture(
		NewLazyTexture(
			path,
			fs,
			int(header.Width),
			int(header.Height)))

	// Finally generate the gpu buffer for the material
	return ResourceManager.Texture(path), nil
}
