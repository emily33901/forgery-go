package render

import (
	"github.com/emily33901/gosigl"
	"github.com/emily33901/lambda-core/core/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

// TODO a lot of this functionality should probably be merged with the actual mesh class

// MeshHelper is a struct that helps with converting meshes to opengl meshes via composition or otherwise
type MeshHelper struct {
	mesh        *mesh.Mesh
	composition *Composition

	// Generated from the result of composition
	vertexObject *gosigl.VertexObject

	dirty bool
}

// Mesh gets the mesh from the helper
// Calling this function sets the dirty flag!
func (m *MeshHelper) Mesh() *mesh.Mesh {
	m.dirty = true
	return m.mesh
}

func (m *MeshHelper) Dirty() bool {
	return m.dirty
}

func (m *MeshHelper) Valid() bool {
	if m.mesh != nil {
		return len(m.mesh.Vertices()) > 0
	}

	return false
}

func (m *MeshHelper) HasBeenBuilt() bool {
	return m.composition != nil && m.vertexObject != nil
}

func (m *MeshHelper) ResetMesh() {
	m.dirty = true
	if m.vertexObject != nil {
		gosigl.DeleteMesh(m.vertexObject)
	}
	m.mesh = mesh.NewMesh()
}

func (m *MeshHelper) SetMesh(newMesh mesh.IMesh) {
	m.dirty = true
	if m.vertexObject != nil {
		gosigl.DeleteMesh(m.vertexObject)
	}
	m.mesh = (newMesh).(*mesh.Mesh)
}

func (m *MeshHelper) AddMesh(newMesh ...mesh.IMesh) {
	m.dirty = true

	for _, x := range newMesh {
		m.mesh.AddVertex(x.Vertices()...)
		m.mesh.AddUV(x.UVs()...)
		m.mesh.AddNormal(x.Normals()...)
		m.mesh.AddColor(x.Colors()...)
		m.mesh.AddLightmapCoordinate(x.LightmapCoordinates()...)
	}

	// Regenerate tangents
	m.mesh.GenerateTangents()
}

// Rebuild rebuilds the opengl vertex object
func (m *MeshHelper) Rebuild() (*Composition, *gosigl.VertexObject) {
	if !m.dirty {
		return m.composition, m.vertexObject
	}

	if m.vertexObject != nil {
		gosigl.DeleteMesh(m.vertexObject)
	}

	compositor := Compositor{}
	compositor.AddMesh(m.mesh)

	m.composition = compositor.ComposeScene()

	m.vertexObject = gosigl.NewMesh(m.composition.Vertices)
	gosigl.CreateVertexAttributeArrayBuffer(m.vertexObject, m.composition.Normals, 3)
	gosigl.CreateVertexAttributeArrayBuffer(m.vertexObject, m.composition.UVs, 2)
	gosigl.CreateVertexAttributeArrayBuffer(m.vertexObject, m.composition.Tangents, 4)
	gosigl.CreateVertexAttributeArrayBuffer(m.vertexObject, m.composition.Colors, 4)
	gosigl.FinishMesh()

	m.dirty = false

	return m.composition, m.vertexObject
}

// TODO: meshes should be in terms of vectors instead of floats to begin with !
func VertsAsVectors(verts []float32) []mgl32.Vec3 {
	ret := make([]mgl32.Vec3, 0, len(verts)/3)

	for i := 0; i < len(verts); i += 3 {
		ret = append(ret, mgl32.Vec3{verts[i], verts[i+1], verts[i+2]})
	}

	return ret
}

func NewMeshHelper() *MeshHelper {
	return &MeshHelper{
		mesh:  mesh.NewMesh(),
		dirty: true,
	}
}
