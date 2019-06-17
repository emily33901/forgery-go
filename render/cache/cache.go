package cache

import (
	"sort"
	"strings"
	"sync"

	"github.com/emily33901/go-forgery/render/lazy"
	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/loader/material"
	"github.com/emily33901/lambda-core/core/logger"
	material2 "github.com/emily33901/lambda-core/core/material"
	"github.com/emily33901/lambda-core/core/resource"
	"github.com/emily33901/gosigl"
)

var textureLookup map[string]gosigl.TextureBindingId
var textureLookupMutex sync.RWMutex

func createEntryForTexture(fs filesystem.IFileSystem, name string) {
	name = strings.ToLower(name)

	if _, found := LookupTextureNoLoad(name); found {
		return
	}

	lazy.LoadSingleLazyMaterial(name+".vmt", fs)

	textureLookupMutex.Lock()
	textureLookup[name] = 0
	textureLookupMutex.Unlock()
}

func BindTexture(fs filesystem.IFileSystem, name string) gosigl.TextureBindingId {
	baseMat := resource.Manager().Material("materials/" + name + ".vmt")
	if baseMat == nil {
		// Really try to make sure this is loaded first
		baseMat = lazy.LoadSingleLazyMaterial(name+".vmt", fs)

		if baseMat == nil {
			// We have actually failed now so use the error texture
			baseMat = resource.Manager().Material(resource.Manager().ErrorTextureName())
		}
	}
	mat := baseMat.(*material2.Material)

	// Make sure this is loaded
	if mat.Textures.Albedo.Reload() != nil {
		// We actually failed to reload this textures data (amazing right?)
		// So use the error material for this one
		baseMat = resource.Manager().Material(resource.Manager().ErrorTextureName())
	}

	mat = baseMat.(*material2.Material)

	newTex := gosigl.CreateTexture2D(
		gosigl.TextureSlot(0),
		mat.Textures.Albedo.Width(),
		mat.Textures.Albedo.Height(),
		mat.Textures.Albedo.PixelDataForFrame(0),
		gosigl.PixelFormat(glTextureFormatFromVtfFormat(mat.Textures.Albedo.Format())),
		false)

	textureLookupMutex.Lock()
	textureLookup[name] = newTex
	textureLookupMutex.Unlock()

	mat.EvictTextures()

	logger.Notice("Bound texture %s", name)

	return newTex
}

// LookupTexture tries to get an individual texture or
// Loads it if necessary
func LookupTexture(fs filesystem.IFileSystem, name string) gosigl.TextureBindingId {
	name = strings.ToLower(name)

	if texId, found := LookupTextureNoLoad(name); found && texId != 0 {
		return texId
	} else if found {
		// texId was 0 so bind it now
		createEntryForTexture(fs, name)
	}

	return BindTexture(fs, name)
}

// LookupTextureNoLoad tries to get a texture but doesnt load one if it doesnt already exist
// Returns nil if the texture hasnt been loaded yet
func LookupTextureNoLoad(name string) (gosigl.TextureBindingId, bool) {
	name = strings.ToLower(name)

	textureLookupMutex.Lock()
	tex, ok := textureLookup[name]
	textureLookupMutex.Unlock()

	if ok == true && tex == 0 {
		logger.Warn("Texture %s is not bound yet...", name)
		return 0, false
	}

	return tex, ok
}

func GetTable() map[string]gosigl.TextureBindingId {
	return textureLookup
}

func GetMaterials() []string {
	materials := make([]string, 0, len(textureLookup))
	textureLookupMutex.Lock()
	for k := range textureLookup {
		materials = append(materials, k)
	}
	textureLookupMutex.Unlock()
	sort.Strings(materials)

	return materials
}

// getGLTextureFormat swap vtf format to openGL format
func glTextureFormatFromVtfFormat(vtfFormat uint32) gosigl.PixelFormat {
	switch vtfFormat {
	case 0:
		return gosigl.RGBA
	case 2:
		return gosigl.RGB
	case 3:
		return gosigl.BGR
	case 12:
		return gosigl.BGRA
	case 13:
		return gosigl.DXT1
	case 14:
		return gosigl.DXT3
	case 15:
		return gosigl.DXT5
	default:
		return gosigl.RGB
	}
}

func InitTextureLookup() {
	textureLookup = make(map[string]gosigl.TextureBindingId)
}

func LoadAllKnownMaterials(fs filesystem.IFileSystem, texDone chan struct{}) int {
	// Ensure error texture is loaded
	material.LoadErrorMaterial()

	// First get all the materials
	materials := make([]string, 0, 2048)
	for _, x := range fs.AllPaths() {
		if strings.Index(x, "materials/") != -1 {
			if strings.Index(x, ".vmt") != -1 {
				if x != "materials/models/player/custom_player/econ/head/ctm_fbi/ctm_fbi_v2_head_varianta.vmt" &&
					x != "materials/models/player/custom_player/econ/head/tm_leet/tm_leet_v2_head_variantc.vmt" {
					materials = append(materials, x[len("materials/"):len(x)-len(".vmt")])
				}
			}
		}
	}

	// Then chunk that array up
	const chunkCount = 8
	chunkSize := len(materials) / chunkCount
	chunks := make([][]string, 0, chunkCount)

	for i := 0; i < len(materials); i += chunkSize {
		end := i + chunkSize

		if end > len(materials) {
			end = len(materials)
		}

		chunks = append(chunks, materials[i:end])
	}

	// Now parallel lazy load all the textures
	for _, x := range chunks {
		go func(materials []string) {
			for _, m := range materials {
				createEntryForTexture(fs, m)
				texDone <- struct{}{}
			}
		}(x)
	}
	logger.Notice("Preloading %d textures...", len(materials))

	return len(materials)
}
