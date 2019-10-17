package windows

import (
	"fmt"
	"sort"

	"github.com/emily33901/lambda-core/core/model"

	"github.com/emily33901/lambda-core/core/math"

	"github.com/emily33901/go-forgery/formats"
	"github.com/emily33901/go-forgery/native"
	"github.com/emily33901/go-forgery/render"
	"github.com/emily33901/go-forgery/render/view"
	"github.com/emily33901/imgui-go"
	"github.com/emily33901/lambda-core/core/entity"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/material"
	"github.com/go-gl/mathgl/mgl32"
)

func igToMglVec2(v imgui.Vec2) mgl32.Vec2 {
	return mgl32.Vec2{v.X, v.Y}
}

type selectionResult struct {
	model *model.Model
	depth float32
}

func createAxesObject() *render.MeshHelper {
	helper := render.NewMeshHelper()
	mesh := helper.Mesh()

	mesh.SetMaterial(material.NewMaterial("editor/axes", material.NewProperties()))
	mesh.AddLine([]float32{1, 0, 0, 1}, mgl32.Vec3{64, 0, 0}, mgl32.Vec3{0, 0, 0})
	mesh.AddLine([]float32{0, 1, 0, 1}, mgl32.Vec3{0, 64, 0}, mgl32.Vec3{0, 0, 0})
	mesh.AddLine([]float32{0, 0, 1, 1}, mgl32.Vec3{0, 0, 64}, mgl32.Vec3{0, 0, 0})

	return helper
}

// TODO this is a MASSIVE hack
var cameraControlFrame int

type SceneWindow struct {
	graphicsAdapter render.Adapter

	window   *view.RenderWindow
	renderer *render.Renderer
	windowId string

	width, height int
	wSize         imgui.Vec2

	scene  *view.Scene
	camera string

	cameraSens     *float32
	cameraMoveSens *float32
	orthoSelected  bool
	orthoMode      int
	lastMouseDrag  imgui.Vec2

	platform native.Platform

	mouseCaptured bool
	open          bool
	inMove        bool

	renderType int

	// TODO: needs to be replaced with a smarter structure
	selectionInProgress bool
	selectionToMake     mgl32.Vec2
	selectionResult     selectionResult
	segmentRay          mgl32.Vec3
	segmentOrigin       mgl32.Vec3

	selectedMeshHelper *render.MeshHelper

	selectionMesh *render.MeshHelper
	axesMesh      *render.MeshHelper
}

func oglToImguiTextureId(id uint32) imgui.TextureID {
	return imgui.TextureID(uint64(id) | (1 << 32))
}

func NewSceneWindow(
	adapter render.Adapter,
	renderer *render.Renderer,
	scene *view.Scene,
	width, height int,
	cameraSens, cameraMoveSens *float32,
	windowId int,
	platform native.Platform) *SceneWindow {

	camera := fmt.Sprintf("Scene_%d", windowId)
	newCamera := formats.NewCamera(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, -1})
	scene.AddCamera(newCamera, camera)

	r := &SceneWindow{
		graphicsAdapter:    adapter,
		renderer:           renderer,
		scene:              scene,
		camera:             camera,
		width:              width,
		height:             height,
		cameraSens:         cameraSens,
		cameraMoveSens:     cameraMoveSens,
		windowId:           fmt.Sprintf("Scene view %d", windowId),
		platform:           platform,
		renderType:         0,
		open:               true,
		selectionMesh:      render.NewMeshHelper(),
		axesMesh:           createAxesObject(),
		selectedMeshHelper: render.NewMeshHelper(),
	}

	r.SelectionChanged()

	// TODO: this is only used by the old line renderer
	renderer.LineWidth = 3

	return r
}

func (window *SceneWindow) Initialize() {
	window.window = view.NewRenderWindow(window.graphicsAdapter, window.width, window.height)
	window.renderer.Initialize()
}

// SelectionChanged handles updating what meshes have been selected
func (window *SceneWindow) SelectionChanged() {
	selectionColor := []float32{1, 0, 0, 0.5}

	// Handle object selection

	aspect := window.wSize.X / window.wSize.Y

	view := window.Camera().ViewMatrix()
	proj := window.Camera().ProjectionMatrix(aspect)

	// Get the point clicked on both the near and far planes
	// and then work out the line between them
	segmentVec := window.Camera().ScreenToWorld(window.selectionToMake.Vec3(1.0), igToMglVec2(window.wSize), aspect)
	segmentOrigin := window.Camera().ScreenToWorld(window.selectionToMake.Vec3(0.1), igToMglVec2(window.wSize), aspect)

	segmentVec = segmentVec.Sub(segmentOrigin)

	window.segmentOrigin = segmentOrigin
	window.segmentRay = segmentVec

	selectionResults := make(map[mgl32.Vec3]selectionResult)

	for _, m := range window.scene.SolidMeshes {
		for _, mesh := range m.Meshes() {
			// Collide with meshes that are made of triangles
			// TODO: we could do with a more robust way of knowing this
			if len(mesh.Vertices())%3 == 0 {
				verts := mesh.Vertices()
				// Transform all verticies
				for i := 0; i < len(verts); i += 3 {
					point, didCollide := math.IntersectSegmentTriangle(segmentOrigin, segmentVec, verts[0], verts[1], verts[2])

					if !didCollide {
						continue
					}

					// We need to project the point to get the depth of the collision
					depth := mgl32.Project(point, view, proj, 0, 0, int(window.wSize.X), int(window.wSize.Y))

					selectionResults[point] = selectionResult{m, depth.Z()}
				}
			}
		}
	}

	// Sort the slection points
	// we are looking for the one that is closest to the camera (lowest Z)

	minResult := mgl32.Vec3{0, 0, 0}

	for k := range selectionResults {
		minResult = k
		break
	}

	for point, result := range selectionResults {
		if result.depth < selectionResults[minResult].depth {
			minResult = point
		}
	}

	resultModel := selectionResults[minResult].model

	// Select the mesh that is selected
	if resultModel != nil {
		window.selectedMeshHelper.ResetMesh()

		mesh := window.selectedMeshHelper.Mesh()

		for _, m := range resultModel.Meshes() {
			window.selectedMeshHelper.AddMesh(m)
		}

		newColors := make([]float32, 0, len(mesh.Vertices())*4)
		for range mesh.Vertices() {
			newColors = append(newColors, selectionColor...)
		}

		window.selectedMeshHelper.Mesh().ResetColors(newColors...)
	}

	window.selectionInProgress = false
}

func (window *SceneWindow) RenderScene() {
	dirtyComposition := window.scene.FrameCompositor.IsOutdated()
	if dirtyComposition {
		window.scene.RecomposeScene()
	}

	window.renderer.StartFrame()

	if c := window.Camera(); c == nil {
		logger.Error("Error trying to get camera %s", window.camera)
		window.renderer.BindCamera(window.scene.Camera("Default_0"), window.wSize.X/window.wSize.Y)
	} else {
		window.renderer.BindCamera(c, window.wSize.X/window.wSize.Y)
	}

	window.window.Bind(window.wSize.X, window.wSize.Y)
	window.renderer.DrawComposition(window.scene.FrameComposed, window.scene.Composition(), window.renderType)
	window.graphicsAdapter.Error()

	// TODO: this no longer has to be done here!
	if window.selectionInProgress {
		logger.Notice("Processing selection!")

		window.SelectionChanged()
	}

	// Render other misc items (axes, selectionpoint)...

	window.renderer.DrawMeshHelper(window.axesMesh, render.ModeWireFrame)
	window.graphicsAdapter.Error()

	if window.selectionMesh.Valid() {
		window.renderer.DrawMeshHelper(window.selectionMesh, render.ModeWireFrame)
		window.graphicsAdapter.Error()
	}

	if window.selectedMeshHelper.Valid() {
		window.renderer.DrawMeshHelper(window.selectedMeshHelper, render.ModeFlat)
		window.graphicsAdapter.Error()
	}

	window.window.Unbind()
}

func (window *SceneWindow) Render(deltaTime float32) {
	// Dont render if we have been closed
	// We will get cleaned up by Forgery{} at the beginning of next frame
	if !window.open {
		logger.Notice("Scene window closed!")
		window.window.Close()
		return
	}

	if imgui.BeginV(window.windowId, &window.open, imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsMenuBar) {
		window.orthoSelected = window.Camera().Ortho()
		window.orthoMode = window.Camera().OrthoDirection()

		if imgui.BeginMenuBar() {
			if imgui.BeginMenu("Movement") {
				imgui.DragFloatV("Aim Sensitivity", window.cameraSens, 0.01, 0, 20, "%f", 5)
				imgui.DragFloatV("Move Sensitivity", window.cameraMoveSens, 0.01, 0, 20, "%f", 5)
				imgui.EndMenu()
			}
			if imgui.BeginMenu("Camera") {
				if imgui.BeginCombo("Cameras", window.camera) {
					names := window.scene.CameraNames()

					// Make sure that names are stable
					sort.Strings(names)

					for _, x := range names {
						if imgui.Selectable(x) {
							window.camera = x
						}
					}
					imgui.EndCombo()
				}

				if imgui.Checkbox("2D View", &window.orthoSelected) {
					window.Camera().SetOrtho(window.orthoSelected)
				}

				if imgui.BeginCombo("2D direction", entity.OrthoDirections[window.orthoMode]) {
					for i, x := range entity.OrthoDirections {
						if imgui.Selectable(x) {
							window.Camera().SetOrthoDirection(i)
						}
					}
					imgui.EndCombo()
				}

				imgui.EndMenu()
			}

			if imgui.BeginMenu("View") {
				if imgui.BeginCombo("Mode", render.RenderModes[window.renderType]) {
					for i, v := range render.RenderModes {
						if imgui.Selectable(v) {
							window.renderType = i
						}
					}
					imgui.EndCombo()
				}

				if imgui.MenuItem("Reset position") {
					window.Camera().SetPos(mgl32.Vec3{0, 0, 0})
				}

				// TODO: The renderer really should not know about any of these params...
				imgui.DragFloatV("Line Width", &window.renderer.LineWidth, 0.01, 0.001, 5, "%f", 1)

				imgui.EndMenu()
			}

			imgui.EndMenuBar()
		}

		imgui.PushStyleColor(imgui.StyleColorChildBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1})
		imgui.PushStyleColor(imgui.StyleColorWindowBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1})

		if window.mouseCaptured {
			imgui.Text("Mouse captured")
		} else {
			imgui.Text("Mouse NOT captured")
		}

		wSize := imgui.ContentRegionAvail()

		// Make sure this is a factor of 2 otherwise we get bad things
		// wSize = imgui.Vec2{wSize.X + float32(int(wSize.X)%2), wSize.Y + float32(int(wSize.Y)%2)}
		window.wSize = wSize
		wPos := imgui.CursorScreenPos()
		aspect := wSize.X / wSize.Y
		// window.Camera().SetAspect(aspect)

		// TODO change 4000 to the framebuffer size

		imgui.ImageButtonV(oglToImguiTextureId(window.window.BufferId()),
			wSize, //imgui.Vec2{X: wSize.X, Y: wSize.Y},
			imgui.Vec2{X: 0, Y: wSize.Y / 4000},
			imgui.Vec2{X: wSize.X / 4000, Y: 0},
			0,
			imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})

		if window.orthoSelected != true {
			// 3D view
			if cameraControlFrame != window.platform.FrameCount() {
				if !window.mouseCaptured {
					if imgui.IsItemHovered() && window.platform.KeyWentDown('Z') {
						window.mouseCaptured = true
						window.platform.SetCursorDisabled(true)

						cameraControlFrame = window.platform.FrameCount()
					}
				} else if window.platform.KeyWentDown('Z') {
					window.mouseCaptured = false
					window.platform.SetCursorDisabled(false)

					cameraControlFrame = window.platform.FrameCount()
				}
			}

			if imgui.IsItemHovered() && !window.selectionInProgress && imgui.IsMouseClicked(0) {
				window.selectionInProgress = true
				curCursorPos := imgui.CurrentIO().MousePos()
				windowPos := curCursorPos.Minus(wPos)
				window.selectionToMake = mgl32.Vec2{windowPos.X, wSize.Y - windowPos.Y}
			}

			if window.mouseCaptured {
				// Draw a little cursor on the screen
				oldCursorPos := imgui.CursorScreenPos()
				imgui.SetCursorScreenPos(wPos.Plus(wSize.Times(0.5)))
				imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{0, 1, 0, 1})
				imgui.Text("+")
				imgui.PopStyleColor()
				imgui.SetCursorScreenPos(oldCursorPos)

				// imgui.CurrentIO().SetMousePosition(imgui.Vec2{0, 0})
				delta := imgui.CurrentIO().MouseDelta()
				scrollDelta := imgui.CurrentIO().MouseWheel()
				delta = delta.Times(*window.cameraSens)

				moveDelta := deltaTime * *window.cameraMoveSens

				if window.platform.IsKeyPressed('W') {
					window.Camera().Forwards(float32(moveDelta))
				}
				if window.platform.IsKeyPressed('S') {
					window.Camera().Backwards(float32(moveDelta))
				}
				if window.platform.IsKeyPressed('A') {
					window.Camera().Left(float32(moveDelta))
				}
				if window.platform.IsKeyPressed('D') {
					window.Camera().Right(float32(moveDelta))
				}

				window.Camera().Zoom(scrollDelta)
				window.Camera().Rotate(-delta.X/180, 0, -delta.Y/180)
			}
		} else if imgui.IsItemHovered() {
			// 2D view
			realDragDelta := imgui.MouseDragDeltaV(2, 0)

			dragDelta := realDragDelta.Minus(window.lastMouseDrag)
			window.lastMouseDrag = realDragDelta
			if realDragDelta.X != 0 || realDragDelta.Y != 0 {
				// TODO this probably shouldnt be done here

				// every 1 pixel has to be scaled by the window size and the camera zoom
				xScale := (window.Camera().OrthoZoom() / wSize.X) * aspect
				yScale := (window.Camera().OrthoZoom() / wSize.Y)

				realDelta := imgui.Vec2{dragDelta.X * xScale, dragDelta.Y * yScale}

				// realDelta := window.Camera().ScreenToWorld(mgl32.Vec2{wSize.X, wSize.Y}, mgl32.Vec2{dragDelta.X, dragDelta.Y})

				var drag3 mgl32.Vec3
				switch window.orthoMode {
				case entity.OrthoX:
					// Z
					drag3[2] = realDelta.X
					// Y
					drag3[1] = realDelta.Y
				case entity.OrthoY:
					// X
					drag3[0] = realDelta.X
					// Z
					drag3[2] = realDelta.Y
				case entity.OrthoZ:
					// X
					drag3[0] = realDelta.X
					// Y
					drag3[1] = realDelta.Y
				}

				window.Camera().Forwards((-realDelta.Y))
				window.Camera().Right((-realDelta.X))
			} else {
				window.lastMouseDrag = imgui.Vec2{0, 0}
			}

			if imgui.IsItemHovered() {
				scrollDelta := imgui.CurrentIO().MouseWheel()

				// Zoom needs to be an exp
				window.Camera().Zoom(scrollDelta)
			}
		}

		imgui.PopStyleColor()
		imgui.PopStyleColor()
	}

	// sorted := window.scene.CompositionMaterials()
	// for k, v := range sorted {
	// 	imgui.Text(k)
	// 	imgui.Image(imgui.TextureID(v), imgui.Vec2{100, 100})
	// }

	imgui.End()
}

// func (window *SceneWindow) newSolidCreated(received event.Dispatchable) {
// 	window.scene.AddSolid(received.(*events.NewSolidCreated).Target())
// }

// func (window *SceneWindow) newCameraCreated(received event.Dispatchable) {
// 	window.scene.AddCamera(received.(*events.NewCameraCreated).Target())
// }

// func (window *SceneWindow) cameraChanged(received event.Dispatchable) {
// 	window.scene.ChangeCamera(received.(*events.CameraChanged).Target())
// }

// func (window *SceneWindow) sceneClosed(received event.Dispatchable) {
// 	window.scene.Close()
// 	window.scene = NewScene()
// 	window.window.Close()
// 	window.window = renderer.NewRenderWindow(window.graphicsAdapter, window.width, window.height)
// }

func (window *SceneWindow) HasClosed() bool {
	return !window.open
}

func (window *SceneWindow) Camera() *entity.Camera {
	return window.scene.Camera(window.camera)
}

func (window *SceneWindow) Close() {
	window.scene.Close()
	window.window.Close()
}
