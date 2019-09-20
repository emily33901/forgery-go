package render

import (
	"github.com/emily33901/go-forgery/render/cache"
	"github.com/emily33901/gosigl"
	"github.com/emily33901/lambda-core/core/entity"
	"github.com/go-gl/gl/v3.2-core/gl"
)

var (
	ModeWireFrame = 0
	ModeTextured  = 1
	ModeFlat      = 2

	RenderModes = [...]string{
		"Wireframe",
		"Textured",
		"Flat",
	}

	shaderNames = [...]string{
		"assets/shaders/UnlitWireframe",
		"assets/shaders/UnlitGeneric",
		"assets/shaders/UnlitFlat",
	}
)

type Renderer struct {
	adapter Adapter

	shaders      []*gosigl.Context
	activeShader int

	uniforms []map[string]int32

	// TODO: dont export these
	WinSize     [2]float32
	LineWidth   float32
	BlendFactor float32
	Fov         float32
}

func (renderer *Renderer) Initialize() {

}

func (renderer *Renderer) StartFrame() {
	renderer.adapter.EnableBlend()
	renderer.adapter.EnableDepthTest()
	renderer.adapter.EnableCullFaceBack()
	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (renderer *Renderer) BindCamera(cam *entity.Camera, aspectRatio float32) {
	model := cam.ModelMatrix()
	view := cam.ViewMatrix()
	proj := cam.ProjectionMatrix(aspectRatio)

	renderer.Fov = cam.Fov()

	for i, s := range renderer.shaders {
		s.UseProgram()

		// TODO We should only be sending mvp - this calculation only needs to be done once!
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["projection"], &proj[0])
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["view"], &view[0])
		renderer.adapter.SendUniformMat4(renderer.uniforms[i]["model"], &model[0])
	}
}

func (renderer *Renderer) beginRender(renderType int, discard bool) {
	// gl.Enable(gl.MULTISAMPLE)

	renderer.shaders[renderType].UseProgram()

	switch renderType {
	case ModeWireFrame:
		renderer.adapter.SendUniformFloat(renderer.uniforms[ModeWireFrame]["lineWidth"], &renderer.LineWidth)
		//renderer.adapter.SendUniformVec4(renderer.uniforms[ModeWireFrame]["color"], &renderer.LineColor[0])
		renderer.adapter.SendUniformFloat(renderer.uniforms[ModeWireFrame]["blendFactor"], &renderer.BlendFactor)

		// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		//gl.Disable(gl.DEPTH_TEST)
		gl.Disable(gl.CULL_FACE)
	case ModeTextured:
		if discard {
			gl.Enable(gl.CULL_FACE)
			gl.Enable(gl.DEPTH_TEST)
		}

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

func (renderer *Renderer) DrawMeshHelper(mesh *MeshHelper, renderType int) {
	if mesh == nil || !mesh.Valid() {
		// If there is nothing to draw then dont try
		return
	}

	comp, glMesh := mesh.Rebuild()

	gosigl.BindMesh(glMesh)
	renderer.adapter.Error()

	renderer.beginRender(renderType, false)
	{
		for _, matObj := range comp.MaterialMeshes() {
			if tex, ok := cache.LookupTextureNoLoad(matObj.Material()); ok {
				gosigl.BindTexture2D(gosigl.TextureSlot(0), tex)
			}

			//gosigl.BindTexture2D(gosigl.TextureSlot(0), 0)

			renderer.adapter.DrawTriangleArray(matObj.Offset(), matObj.Length())
			renderer.adapter.Error()
		}
	}
	renderer.endRender(renderType)
}

func (renderer *Renderer) DrawComposition(composition *Composition, mesh *gosigl.VertexObject, renderType int) {
	if mesh == nil {
		return
	}

	{
		gosigl.BindMesh(mesh)
		renderer.adapter.Error()

		renderer.beginRender(renderType, false)
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
	}
}

func NewRenderer(adapter Adapter) *Renderer {
	renderer := &Renderer{
		adapter:  adapter,
		uniforms: []map[string]int32{},
	}

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
			uniforms["blendFactor"] = s.GetUniform("blendFactor")
		}

		renderer.uniforms = append(renderer.uniforms, uniforms)
	}

	return renderer
}
