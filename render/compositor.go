package render

import (
	"github.com/emily33901/lambda-core/core/mesh"
)

type Composition struct {
	mesh.Mesh
	materialCompositions []*compositionMesh

	indices []uint32
}

// Compose constructs the indices information for the current state of the Composition
func (comp *Composition) Compose() {
	comp.indices = make([]uint32, 0)
	for _, materialComposition := range comp.materialCompositions {
		materialComposition.GenerateIndicesList()
		comp.indices = append(comp.indices, materialComposition.indices...)
	}
}

// MaterialMeshes returns composed material information
func (comp *Composition) MaterialMeshes() []*compositionMesh {
	return comp.materialCompositions
}

// Indices returns the indices of this compositions faces
func (comp *Composition) Indices() []uint32 {
	return comp.indices
}

// AddMesh
func (comp *Composition) AddMesh(mat *compositionMesh) {
	comp.materialCompositions = append(comp.materialCompositions, mat)
}

// NewComposition returns a new Composition.
func NewComposition() *Composition {
	return &Composition{}
}

type compositionMesh struct {
	texturePath string
	offset      int
	length      int

	indices []uint32
}

func (texMesh *compositionMesh) Material() string {
	return texMesh.texturePath
}

// Indices returns all indices for vertices that use this material
func (texMesh *compositionMesh) Indices() []uint32 {
	return texMesh.indices
}

// Indices returns the Offset for vertices that use this material
func (texMesh *compositionMesh) Offset() int32 {
	return int32(texMesh.offset)
}

// Indices returns the number for vertices that use this material
func (texMesh *compositionMesh) Length() int32 {
	return int32(texMesh.length)
}

// GenerateIndicesList generates the indices list from offset and length of Composition vertex data.
func (texMesh *compositionMesh) GenerateIndicesList() {
	indices := make([]uint32, 0)
	for i := texMesh.offset; i < texMesh.offset+texMesh.length; i++ {
		indices = append(indices, uint32(i))
	}

	texMesh.indices = indices
}

// NewCompositionMesh returns a new compositionMesh
func NewCompositionMesh(texName string, offset int, length int) *compositionMesh {
	return &compositionMesh{
		texturePath: texName,
		length:      length,
		offset:      offset,
	}
}

// Compositor is a struct that provides a mechanism to compose 1 or more models into a single renderable set of data,
// indexed by material.
// This is super handy for reducing draw calls down a bunch.
// A resultant Composition should result in a single set of vertex data + 1 pair of index offset+length info per material
// referenced by all models composed.
type Compositor struct {
	meshes []mesh.IMesh

	isOutdated bool
}

// AddModel adds a new model to be composed.
func (compositor *Compositor) AddMesh(m mesh.IMesh) {
	compositor.meshes = append(compositor.meshes, m)
	compositor.isOutdated = true
}

func (compositor *Compositor) IsOutdated() bool {
	return compositor.isOutdated
}

// ComposeScene builds a sceneComposition mesh for rendering
func (compositor *Compositor) ComposeScene() *Composition {
	compositor.isOutdated = false
	texMappings := map[string][]mesh.IMesh{}

	// Step 1. Map meshes into contiguous groups by texture
	for idx, m := range compositor.meshes {
		if _, ok := texMappings[m.Material().FilePath()]; !ok {
			texMappings[m.Material().FilePath()] = make([]mesh.IMesh, 0)
		}

		texMappings[m.Material().FilePath()] = append(texMappings[m.Material().FilePath()], compositor.meshes[idx])
	}

	// Step 2. Construct a single vertex object Composition ordered by material
	sceneComposition := NewComposition()
	vertCount := 0
	for key, texMesh := range texMappings {
		// @TODO verify if this is the vertex offset of the actual array offset (vertexOffset * 3)
		matVertOffset := vertCount
		matVertCount := 0
		for _, sMesh := range texMesh {
			sceneComposition.AddVertex(sMesh.Vertices()...)
			sceneComposition.AddNormal(sMesh.Normals()...)
			sceneComposition.AddUV(sMesh.UVs()...)

			matVertCount += len(sMesh.Vertices()) / 3
		}

		sceneComposition.AddMesh(NewCompositionMesh(key, matVertOffset, matVertCount))
		vertCount += matVertCount
	}
	sceneComposition.GenerateTangents()

	// Step 3. Generate indices from composed materials
	sceneComposition.Compose()

	return sceneComposition
}
