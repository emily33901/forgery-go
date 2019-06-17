package render

import (
	"github.com/emily33901/go-forgery/render/cache"
	"github.com/emily33901/gosigl"
	"github.com/emily33901/lambda-core/core/entity"
	"github.com/emily33901/lambda-core/core/material"
	"github.com/emily33901/lambda-core/core/mesh"
	"github.com/go-gl/gl/v3.2-core/gl"
)

var (
	ModeTextured  = 1
	ModeWireFrame = 0

	RenderModes = [...]string{
		"Wireframe",
		"Textured",
	}

	shaderNames = [...]string{
		"assets/shaders/UnlitWireframe",
		"assets/shaders/UnlitGeneric",
	}
)

type Renderer struct {
	adapter Adapter

	shaders      []*gosigl.Context
	activeShader int

	uniforms []map[string]int32

	// @TODO: dont export these
	WinSize     [2]float32
	LineColor   [4]float32
	LineWidth   float32
	BlendFactor float32
	Fov         float32

	AxisComposition *Composition
	AxisMesh        *gosigl.VertexObject
}

func (renderer *Renderer) Initialize() {

}

func (renderer *Renderer) StartFrame() {
	renderer.adapter.EnableBlend()
	renderer.adapter.EnableDepthTest()
	renderer.adapter.EnableCullFaceBack()
	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (renderer *Renderer) BindCamera(cam *entity.Camera) {
	model := cam.ModelMatrix()
	view := cam.ViewMatrix()
	proj := cam.ProjectionMatrix()

	renderer.Fov = cam.Fov()

	for i, s := range renderer.shaders {
		s.UseProgram()

		// @TODO We should only be sending mvp - this calculation only needs to be done once!
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["projection"], &proj[0])
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["view"], &view[0])
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["model"], &model[0])
	}
}

func (renderer *Renderer) beginRender(renderType int) {
	// gl.Enable(gl.MULTISAMPLE)
	switch renderType {
	case ModeWireFrame:
		renderer.shaders[ModeWireFrame].UseProgram()

		renderer.adapter.SendUniformFloat(renderer.uniforms[ModeWireFrame]["lineWidth"], &renderer.LineWidth)
		renderer.adapter.SendUniformVec4(renderer.uniforms[ModeWireFrame]["color"], &renderer.LineColor[0])
		renderer.adapter.SendUniformFloat(renderer.uniforms[ModeWireFrame]["blendFactor"], &renderer.BlendFactor)

		gl.Disable(gl.DEPTH_TEST)
		gl.Disable(gl.CULL_FACE)
	case ModeTextured:
		renderer.shaders[ModeTextured].UseProgram()
		renderer.adapter.EnableBlend()
	}
}

func (renderer *Renderer) endRender(renderType int) {
	// gl.Disable(gl.MULTISAMPLE)

	switch renderType {
	case ModeWireFrame:

	case ModeTextured:
	}
}

func (renderer *Renderer) DrawComposition(composition *Composition, mesh *gosigl.VertexObject, renderType int) {
	if mesh == nil {
		return
	}

	gosigl.BindMesh(mesh)
	renderer.adapter.Error()

	renderer.beginRender(renderType)
	{
		for _, matObj := range composition.MaterialMeshes() {
			if tex, ok := cache.LookupTextureNoLoad(matObj.Material()); ok {
				gosigl.BindTexture2D(gosigl.TextureSlot(0), tex)
			}

			//gosigl.BindTexture2D(gosigl.TextureSlot(0), 0)

			renderer.adapter.DrawTriangleArray(matObj.Offset(), matObj.Length())
			renderer.adapter.Error()
		}
	}
	renderer.endRender(renderType)

	// Draw the unit axis

	gosigl.BindMesh(renderer.AxisMesh)
	renderer.adapter.Error()

	oldLineColor := renderer.LineColor

	for i, matObj := range renderer.AxisComposition.MaterialMeshes() {
		// We dont need to lookup textures here. There wont be any

		renderer.LineColor = [4]float32{0, 0, 0, 1}
		renderer.LineColor[i] = 1.0

		renderer.beginRender(ModeWireFrame)

		renderer.adapter.SendUniformVec4(renderer.uniforms[ModeWireFrame]["color"], &renderer.LineColor[0])

		{
			renderer.adapter.DrawTriangleArray(matObj.Offset(), matObj.Length())
			renderer.adapter.Error()
		}
		renderer.endRender(ModeWireFrame)
	}

	renderer.LineColor = oldLineColor
}

func NewRenderer(adapter Adapter) *Renderer {
	renderer := &Renderer{
		adapter:  adapter,
		uniforms: []map[string]int32{},
	}

	// @TODO: We really should have to be doing this!
	tempCompositor := &Compositor{}

	// @TODO this is a huge hack becuase we dont have a way to get a mesh
	// directly to a vertex buffer without using this!

	tempMesh := mesh.NewMesh()
	tempMesh.SetMaterial(material.NewMaterial("editor/wireframe1"))
	tempMesh.AddVertex(64, 0, 0)
	tempMesh.AddVertex(0, 0, 4)
	tempMesh.AddVertex(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempCompositor.AddMesh(tempMesh)

	tempMesh = mesh.NewMesh()
	tempMesh.SetMaterial(material.NewMaterial("editor/wireframe2"))
	tempMesh.AddVertex(0, 64, 0)
	tempMesh.AddVertex(4, 0, 0)
	tempMesh.AddVertex(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempCompositor.AddMesh(tempMesh)

	tempMesh = mesh.NewMesh()
	tempMesh.SetMaterial(material.NewMaterial("editor/wireframe3"))
	tempMesh.AddVertex(0, 0, 64)
	tempMesh.AddVertex(0, 4, 0)
	tempMesh.AddVertex(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddNormal(0, 0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempMesh.AddUV(0, 0)
	tempCompositor.AddMesh(tempMesh)

	renderer.AxisComposition = tempCompositor.ComposeScene()
	sceneMesh := gosigl.NewMesh(renderer.AxisComposition.Vertices())
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, renderer.AxisComposition.Normals(), 3)
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, renderer.AxisComposition.UVs(), 2)
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, renderer.AxisComposition.Tangents(), 3)
	gosigl.FinishMesh()

	renderer.AxisMesh = sceneMesh

	for i, s := range shaderNames {
		s := renderer.adapter.LoadSimpleShader(s)
		renderer.shaders = append(renderer.shaders, s)

		// Get uniforms for the new shader
		s.UseProgram()

		uniforms := make(map[string]int32)

		uniforms["model"] = s.GetUniform("model")
		uniforms["view"] = s.GetUniform("view")
		uniforms["projection"] = s.GetUniform("projection")

		switch i {
		case ModeTextured:
			break
		case ModeWireFrame:
			uniforms["lineWidth"] = s.GetUniform("lineWidth")
			uniforms["color"] = s.GetUniform("color")
			uniforms["blendFactor"] = s.GetUniform("blendFactor")
		}

		renderer.uniforms = append(renderer.uniforms, uniforms)
	}

	return renderer
}
