package lazy

import (
	"strings"

	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/material"
	"github.com/emily33901/lambda-core/core/resource"
	"github.com/emily33901/lambda-core/core/texture"
	"github.com/golang-source-engine/vmt"
)

// loadMaterialsLazy "private" function that actually does the loading
func loadMaterialsLazy(fs filesystem.IFileSystem, materialList ...string) (missingList []string) {
	ResourceManager := resource.Manager()

	for _, materialPath := range materialList {
		vtfTexturePath := ""

		if !strings.HasSuffix(materialPath, filesystem.ExtensionVmt) {
			materialPath += filesystem.ExtensionVmt
		}

		if ResourceManager.HasMaterial(filesystem.BasePathMaterial + materialPath) {
			continue
		}

		mat, err := vmt.FromFilesystem(materialPath, fs, vmt.NewProperties())

		if err != nil {
			logger.Warn("Failed to load material: %s. Reason: %s", filesystem.BasePathMaterial+materialPath, err)
			missingList = append(missingList, materialPath)
			continue
		}
		properties := mat.(*vmt.Properties)

		material := material.NewMaterial(materialPath, properties)

		if material.Props.BaseTexture == "" {
			material.Textures.Albedo = ResourceManager.Texture(ResourceManager.ErrorTextureName()).(texture.ITexture)
			missingList = append(missingList, materialPath)

			ResourceManager.AddMaterial(material)
			continue
		}

		// NOTE: in patch vmts include is not supported
		vtfTexturePath = material.Props.BaseTexture
		if !strings.HasSuffix(vtfTexturePath, filesystem.ExtensionVtf) {
			vtfTexturePath = vtfTexturePath + filesystem.ExtensionVtf
		}

		material.Textures.Albedo = LoadLazyTexture(vtfTexturePath, fs)

		if material.Textures.Albedo == nil {
			material.Textures.Albedo = ResourceManager.Texture(ResourceManager.ErrorTextureName()).(texture.ITexture)
			missingList = append(missingList, materialPath)
			ResourceManager.AddMaterial(material)
			continue
		}

		if material.Props.Bumpmap != "" {
			material.Textures.Normal = LoadLazyTexture(material.Props.Bumpmap, fs)
		}
		ResourceManager.AddMaterial(material)
	}
	return missingList
}

// LoadSingleMaterial loads a single material with known file path
func LoadSingleLazyMaterial(filePath string, fs filesystem.IFileSystem) material.IMaterial {
	if resource.Manager().HasMaterial(filesystem.BasePathMaterial + filePath) {
		return resource.Manager().Material(filesystem.BasePathMaterial + filePath)
	}

	result := loadMaterialsLazy(fs, filePath)
	if len(result) == 0 {
		return resource.Manager().Material(filesystem.BasePathMaterial + filePath)

	}
	return resource.Manager().Material(resource.Manager().ErrorTextureName())
}
