package view

import (
	"fmt"

	"github.com/emily33901/go-forgery/formats"
	"github.com/emily33901/go-forgery/render"
	"github.com/emily33901/go-forgery/render/cache"
	"github.com/emily33901/go-forgery/render/convert"
	"github.com/emily33901/go-forgery/valve/world"
	"github.com/emily33901/gosigl"
	"github.com/emily33901/lambda-core/core/entity"
	"github.com/emily33901/lambda-core/core/filesystem"
	model "github.com/emily33901/lambda-core/core/model"
	"github.com/go-gl/mathgl/mgl32"
)

type Scene struct {
	Solids      map[int]*world.Solid
	SolidMeshes map[int]*model.Model

	cameras map[string]*entity.Camera
	// activeCamera *entity.Camera

	FrameCompositor *render.Compositor
	FrameComposed   *render.Composition
	FrameMesh       *gosigl.VertexObject
}

func (scene *Scene) Camera(name string) *entity.Camera {
	return scene.cameras[name]
}

func (scene *Scene) CameraNames() []string {
	results := make([]string, 0, len(scene.cameras))
	for k := range scene.cameras {
		results = append(results, k)
	}

	return results
}

func (scene *Scene) Composition() *gosigl.VertexObject {
	return scene.FrameMesh
}

func (scene *Scene) RecomposeScene(fs filesystem.IFileSystem) *gosigl.VertexObject {
	if scene.FrameMesh != nil {
		gosigl.DeleteMesh(scene.FrameMesh)
	}

	scene.FrameComposed = scene.FrameCompositor.ComposeScene()
	sceneMesh := gosigl.NewMesh(scene.FrameComposed.Vertices())
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, scene.FrameComposed.Normals(), 3)
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, scene.FrameComposed.UVs(), 2)
	gosigl.CreateVertexAttributeArrayBuffer(sceneMesh, scene.FrameComposed.Tangents(), 3)
	gosigl.FinishMesh()

	scene.FrameMesh = sceneMesh

	// Make sure that we have all the textures we need
	for _, m := range scene.FrameComposed.MaterialMeshes() {
		cache.LookupTexture(fs, m.Material())
	}
	return scene.FrameMesh
}

func (scene *Scene) AddSolid(solid *world.Solid) {
	scene.Solids[solid.Id] = solid

	model := convert.SolidToModel(solid)
	scene.SolidMeshes[solid.Id] = model

	for idx := range model.Meshes() {
		scene.FrameCompositor.AddMesh(model.Meshes()[idx])
	}
}

func (scene *Scene) AddCamera(camera *formats.Camera, name string) {
	c := entity.NewCamera(70, 1024/768)

	c.Transform().Position = mgl32.Vec3{float32(camera.Position.X()), float32(camera.Position.Y()), float32(camera.Position.Z())}
	c.Transform().Rotation = mgl32.Vec3{
		mgl32.DegToRad(float32(camera.Look.X())),
		mgl32.DegToRad(float32(camera.Look.Y())),
		mgl32.DegToRad(float32(camera.Look.Z()))}

	scene.cameras[name] = c
}

func (scene *Scene) Close() {
	gosigl.DeleteMesh(scene.FrameMesh)
}

func NewScene() *Scene {
	return &Scene{
		Solids:          map[int]*world.Solid{},
		SolidMeshes:     map[int]*model.Model{},
		cameras:         map[string]*entity.Camera{},
		FrameCompositor: &render.Compositor{},
	}
}

func NewSceneFromVmf(vmf *formats.Vmf) *Scene {
	s := NewScene()

	for i := 0; i < vmf.Entities().Length(); i++ {
		// s.AddSolid(solid *world.Solid)
		// widget.dispatcher.Dispatch(events.NewEntityCreated(project.Vmf.Entities().Get(i)))
	}

	for i := 0; i < len(vmf.Worldspawn().Solids); i++ {
		s.AddSolid(&vmf.Worldspawn().Solids[i])
	}

	for i := 0; i < len(vmf.Cameras().CameraList); i++ {
		s.AddCamera(&vmf.Cameras().CameraList[i], fmt.Sprintf("Default_%d", i))
	}

	return s
}
