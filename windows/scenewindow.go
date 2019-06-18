package windows

import (
	"fmt"
	"sort"

	"github.com/emily33901/go-forgery/formats"
	"github.com/emily33901/go-forgery/native"
	"github.com/emily33901/go-forgery/render"
	"github.com/emily33901/go-forgery/render/view"
	"github.com/emily33901/imgui-go"
	"github.com/emily33901/lambda-core/core/entity"
	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/go-gl/mathgl/mgl32"
)

// @TODO this is a MASSIVE hack
var cameraControlFrame int

type SceneWindow struct {
	filesystem      filesystem.IFileSystem
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
}

func NewSceneWindow(fs filesystem.IFileSystem,
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
		filesystem:      fs,
		graphicsAdapter: adapter,
		renderer:        renderer,
		scene:           scene,
		camera:          camera,
		width:           width,
		height:          height,
		cameraSens:      cameraSens,
		cameraMoveSens:  cameraMoveSens,
		windowId:        fmt.Sprintf("Scene view %d", windowId),
		platform:        platform,
		renderType:      0,
		open:            true,
	}

	renderer.LineWidth = 3
	renderer.LineColor = [4]float32{1, 1, 1, 1}

	return r
}

func (window *SceneWindow) Initialize() {
	window.window = view.NewRenderWindow(window.graphicsAdapter, window.width, window.height)
	window.renderer.Initialize()

	// window.renderer.BindShader(window.graphicsAdapter.LoadSimpleShader("assets/shaders/UnlitWireframe"))
}

func (window *SceneWindow) RenderScene() {
	dirtyComposition := window.scene.FrameCompositor.IsOutdated()
	if dirtyComposition {
		window.scene.RecomposeScene(window.filesystem)
	}

	window.renderer.StartFrame()

	if c := window.Camera(); c == nil {
		logger.Error("Error trying to get camera %s", window.camera)
		window.renderer.BindCamera(window.scene.Camera("Default_0"))
	} else {
		window.renderer.BindCamera(c)
	}

	// @TODO: this isnt used...
	window.renderer.BlendFactor = 2.0

	// window.renderer.BindCamera(window.scene.ActiveCamera())
	// if window.window.Width() != int(window.wSize.X) {
	// 	window.window.SetSize(int(window.wSize.X), int(window.wSize.Y))
	// }
	window.window.Bind(window.wSize.X, window.wSize.Y)
	window.renderer.DrawComposition(window.scene.FrameComposed, window.scene.Composition(), window.renderType)
	window.graphicsAdapter.Error()
	window.window.Unbind()
}

func (window *SceneWindow) Render(deltaTime float64) {
	// Dont render if we have been closed
	if !window.open {
		logger.Notice("Scene window closed!")
		window.window.Close()
		return
	}

	if imgui.BeginV(window.windowId, &window.open, imgui.WindowFlagsNoBringToFrontOnFocus|imgui.WindowFlagsNoScrollbar|imgui.WindowFlagsMenuBar) {
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

				// @TODO: The renderer really should not know about any of these params...
				imgui.DragFloatV("Line Width", &window.renderer.LineWidth, 0.01, 0.001, 5, "%f", 1)
				imgui.ColorEdit4("Line Color", &window.renderer.LineColor)

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
		window.Camera().SetAspect(wSize.X / wSize.Y)

		// @TODO change 4000 to the framebuffer size

		imgui.ImageButtonV(imgui.TextureID(window.window.BufferId()),
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

				if window.platform.IsKeyPressed('W') {
					window.Camera().Forwards(float64(deltaTime))
				}
				if window.platform.IsKeyPressed('S') {
					window.Camera().Backwards(float64(deltaTime))
				}
				if window.platform.IsKeyPressed('A') {
					window.Camera().Left(float64(deltaTime))
				}
				if window.platform.IsKeyPressed('D') {
					window.Camera().Right(float64(deltaTime))
				}

				// @TODO fov is in radians (It probably shouldnt be!)
				window.Camera().Zoom(-0.01 * scrollDelta)
				window.Camera().Rotate(-delta.X/180, 0, -delta.Y/180)
			}
		} else {
			// 2D view
			realDragDelta := imgui.MouseDragDeltaV(2, 0)
			dragDelta := realDragDelta.Minus(window.lastMouseDrag)
			window.lastMouseDrag = realDragDelta
			if dragDelta.X != 0 || dragDelta.Y != 0 {
				var drag4 mgl32.Vec4

				// @TODO this probably shouldnt be done here
				switch window.orthoMode {
				case entity.OrthoX:
					// Z
					drag4[2] = dragDelta.X
					// Y
					drag4[1] = dragDelta.Y
				case entity.OrthoY:
					// X
					drag4[0] = dragDelta.X
					// Z
					drag4[2] = dragDelta.Y
				case entity.OrthoZ:
					// X
					drag4[0] = dragDelta.X
					// Y
					drag4[1] = dragDelta.Y
				}

				realDelta := window.Camera().ScreenToWorld(mgl32.Vec2{wSize.X, wSize.Y}, mgl32.Vec2{dragDelta.X, dragDelta.Y})

				if realDelta.Len() > 0 {
					window.Camera().Move(realDelta)
				}
			} else {
				window.lastMouseDrag = imgui.Vec2{0, 0}
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
