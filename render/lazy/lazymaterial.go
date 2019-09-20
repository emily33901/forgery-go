package lazy

import (
	"errors"
	"strings"

	"github.com/emily33901/lambda-core/core/filesystem"
	keyvalues2 "github.com/emily33901/lambda-core/core/loader/keyvalues"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/material"
	"github.com/emily33901/lambda-core/core/resource"
	"github.com/emily33901/lambda-core/core/texture"
	keyvalues "github.com/galaco/KeyValues"
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

		mat, err := readVmt(filesystem.BasePathMaterial+materialPath, fs)
		if err != nil {
			logger.Warn("Failed to load material: %s. Reason: %s", filesystem.BasePathMaterial+materialPath, err)
			missingList = append(missingList, materialPath)
			continue
		}
		vmt := mat.(*material.Material)

		if vmt.BaseTextureName == "" {
			vmt.Textures.Albedo = ResourceManager.Texture(ResourceManager.ErrorTextureName()).(texture.ITexture)
			missingList = append(missingList, materialPath)

			ResourceManager.AddMaterial(vmt)
			continue
		}

		// NOTE: in patch vmts include is not supported
		vtfTexturePath = vmt.BaseTextureName
		if !strings.HasSuffix(vtfTexturePath, filesystem.ExtensionVtf) {
			vtfTexturePath = vtfTexturePath + filesystem.ExtensionVtf
		}

		vmt.Textures.Albedo = LoadLazyTexture(vtfTexturePath, fs)

		if vmt.Textures.Albedo == nil {
			vmt.Textures.Albedo = ResourceManager.Texture(ResourceManager.ErrorTextureName()).(texture.ITexture)
			missingList = append(missingList, materialPath)
			ResourceManager.AddMaterial(vmt)
			continue
		}

		if vmt.BumpMapName != "" {
			vmt.Textures.Normal = LoadLazyTexture(vmt.BumpMapName, fs)
		}
		ResourceManager.AddMaterial(vmt)
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

// TODO vmt.go should expose these so that i dont have to copy them here

func readVmt(path string, fs filesystem.IFileSystem) (material.IMaterial, error) {
	kvs, err := keyvalues2.ReadKeyValues(path, fs)
	if err != nil {
		return nil, err
	}
	roots, err := kvs.Children()
	if err != nil {
		return nil, err
	}
	root := roots[0]

	include, err := root.Find("include")
	if include != nil && err == nil {
		includePath, _ := include.AsString()
		root, err = mergeIncludedVmtRecursive(root, includePath, fs)
		if err != nil {
			return nil, err
		}
	}

	// @NOTE this will be replaced with a proper kv->material builder
	mat, err := materialFromKeyValues(root, path)
	if err != nil {
		return nil, err
	}
	return mat, nil
}

func mergeIncludedVmtRecursive(base *keyvalues.KeyValue, includePath string, fs filesystem.IFileSystem) (*keyvalues.KeyValue, error) {
	parent, err := keyvalues2.ReadKeyValues(includePath, fs)
	if err != nil {
		return base, errors.New("failed to read included vmt")
	}
	result, err := base.Patch(parent)
	if err != nil {
		return base, errors.New("failed to merge included vmt")
	}
	include, err := result.Find("include")
	if err == nil {
		newIncludePath, _ := include.AsString()
		if newIncludePath == includePath {
			err = result.RemoveChild("include")
			return &result, err
		}
		return mergeIncludedVmtRecursive(&result, newIncludePath, fs)
	}
	return &result, nil
}

func materialFromKeyValues(kv *keyvalues.KeyValue, path string) (*material.Material, error) {
	shaderName := kv.Key()

	// $basetexture
	baseTexture := findKeyValueAsString(kv, "$basetexture")

	// $bumpmap
	bumpMapTexture := findKeyValueAsString(kv, "$bumpmap")

	mat := material.NewMaterial(path)
	mat.ShaderName = shaderName
	mat.BaseTextureName = baseTexture
	mat.BumpMapName = bumpMapTexture
	return mat, nil
}

func findKeyValueAsString(kv *keyvalues.KeyValue, keyName string) string {
	target, err := kv.Find(keyName)
	if err != nil {
		return ""
	}
	ret, err := target.AsString()
	if err != nil {
		return ""
	}

	return ret
}
