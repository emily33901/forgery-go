package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"time"

	"github.com/emily33901/go-forgery/valve"

	"github.com/emily33901/go-forgery/formats"
	"github.com/emily33901/go-forgery/native"
	"github.com/emily33901/go-forgery/render"
	"github.com/emily33901/go-forgery/render/adapters"
	"github.com/emily33901/go-forgery/render/cache"
	"github.com/emily33901/go-forgery/render/view"
	"github.com/emily33901/go-forgery/windows"
	imgui "github.com/emily33901/imgui-go"
	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/logger"
)

type stdOut struct{}

func (log *stdOut) Write(data []byte) (n int, err error) {
	line := string(data) + "\n"
	return os.Stdout.Write([]byte(line))
}

// Renderer covers rendering imgui draw data.
type ImguiRenderer interface {
	// PreRender causes the display buffer to be prepared for new output.
	PreRender(clearColor [4]float32)
	// Render draws the provided imgui draw data.
	Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData)

	Dispose()
}

type ForgeryContext struct {
	filesystem    filesystem.IFileSystem
	context       *imgui.Context
	platform      native.Platform
	imguiRenderer ImguiRenderer

	render  *render.Renderer
	adapter render.Adapter

	documentLoaded    bool
	activeMap         *formats.Vmf
	sceneWindows      []*windows.SceneWindow
	lastSceneWindowId int
	scene             *view.Scene
	activeCamera      *formats.Camera

	deltaTime time.Duration

	// UI stuff
	slider            float32
	counter           int
	showAnotherWindow bool
	lastTime          time.Time
	overlayCorner     int
	fpsHistory        []float32

	// Do we show these windows?
	showDemoWindow      bool
	showAboutWindow     bool
	showMaterialsWindow bool
	showInfoOverlay     bool

	// @TODO these really shouldnt be here!
	cameraSens     float32
	cameraMoveSens float32

	// Sent down when textures are finished loading
	texturesLoadedCount         int
	texturesLoadedExpected      int
	texturesLoadingCompleteChan chan struct{}
	texturesLoadingComplete     bool
}

func (f *ForgeryContext) RenderScene() {
	for _, s := range f.sceneWindows {
		s.RenderScene()
	}
}

func (f *ForgeryContext) RenderUI() {
	// 1. Show the big demo window (Most of the sample code is in ImGui::ShowDemoWindow()!
	// You can browse its code to learn more about Dear ImGui!).

	if !f.texturesLoadingComplete {
		done := false
		for i := 0; i < 100 && !done; i++ {
			select {
			case _ = <-f.texturesLoadingCompleteChan:
				f.texturesLoadedCount++
			default:
				done = true
			}
		}

		imgui.BeginV("Loading textures", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoScrollbar)
		imgui.ProgressBarV(float32(f.texturesLoadedCount)/float32(f.texturesLoadedExpected), imgui.Vec2{200, 30},
			fmt.Sprintf("%d/%d", f.texturesLoadedCount, f.texturesLoadedExpected))
		imgui.End()

		if f.texturesLoadedCount >= f.texturesLoadedExpected {
			f.texturesLoadingComplete = true
			close(f.texturesLoadingCompleteChan)
		}
	}

	if f.showDemoWindow {
		imgui.ShowDemoWindow(&f.showDemoWindow)
	}

	f.deltaTime = time.Now().Sub(f.lastTime)
	f.lastTime = time.Now()
	f.fpsHistory = append(f.fpsHistory, float32(1.0/f.deltaTime.Seconds()))
	if len(f.fpsHistory) > 120 {
		f.fpsHistory = f.fpsHistory[1:]
	}

	if f.documentLoaded {
		for _, window := range f.sceneWindows {
			// We do this here because we dont have a deltaTime in sceneWindow.Render()
			window.Camera().Update(f.deltaTime.Seconds())

			window.Render(f.deltaTime.Seconds())
		}
	}

	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItem("New") {
			}
			if imgui.MenuItemV("Open", "Ctrl-O", false, true) {

			}
			if imgui.BeginMenu("Recent") {
				imgui.EndMenu()
			}
			if imgui.MenuItemV("Save", "Ctrl-S", false, true) {
			}
			if imgui.MenuItem("Save As...") {
			}
			if imgui.BeginMenu("Color Scheme") {
				if imgui.MenuItem("Light") {
					imgui.StyleColorsLight()
				}
				if imgui.MenuItem("Dark") {
					imgui.StyleColorsDark()
				}
				imgui.EndMenu()
			}
			if imgui.MenuItem("Quit") {
				// @TODO: NEVER DO THIS AAARRRGHHH
				os.Exit(0)
			}

			imgui.EndMenu()
		}

		if imgui.BeginMenu("View") {
			if imgui.MenuItem("New View") {
				f.NewSceneWindow()
			}
			if imgui.MenuItem("Material Viewer") {
				f.showMaterialsWindow = true
			}
			if imgui.Checkbox("Overlay", &f.showInfoOverlay) {
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("About") {
			if imgui.MenuItem("Show demo window") {
				f.showDemoWindow = true
			}
			imgui.Separator()

			if imgui.MenuItem("About") {
				f.showAboutWindow = true
			}

			imgui.EndMenu()
		}
		imgui.EndMainMenuBar()
	}

	if f.showMaterialsWindow {
		windows.RenderMaterialsWindow(f.filesystem, &f.showMaterialsWindow)
	}

	if f.showInfoOverlay {
		const overlayDistanceX = 10.0
		const overlayDistanceY = 30.0

		if f.overlayCorner != -1 {
			size := f.platform.DisplaySize()
			pos := imgui.Vec2{}
			pivot := imgui.Vec2{}

			if f.overlayCorner&1 != 1 {
				pos.X = overlayDistanceX

				pivot.X = 0
			} else {
				pos.X = size[0] - overlayDistanceX
				pivot.X = 1
			}

			if f.overlayCorner&2 != 2 {
				pos.Y = overlayDistanceY
				pivot.Y = 0
			} else {
				pos.Y = size[1] - overlayDistanceY
				pivot.Y = 1
			}

			imgui.SetNextWindowPosV(pos, imgui.ConditionAlways, pivot)
		}

		imgui.SetNextWindowBgAlpha(0.33)

		flags := imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsAlwaysAutoResize |
			imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoFocusOnAppearing |
			imgui.WindowFlagsNoNav

		if f.overlayCorner != -1 {
			flags |= imgui.WindowFlagsNoMove
		}

		if imgui.BeginV("Info Overlay", &f.showInfoOverlay, flags) {
			// Calcuate some fps statistics
			minFps := 1111111.0
			maxFps := 0.0
			totalFps := float64(0.0)
			for _, x := range f.fpsHistory {
				fx := float64(x)
				totalFps += float64(fx)
				if fx < minFps {
					minFps = fx
				}
				if fx > maxFps {
					maxFps = fx
				}
			}

			imgui.PlotLinesV("", f.fpsHistory, 0, "FPS", float32(minFps)-10, float32(maxFps)+10, imgui.Vec2{0, 80})

			imgui.Text(fmt.Sprintf("Average FPS: %f", totalFps/float64(len(f.fpsHistory))))
			imgui.Separator()

			if imgui.BeginPopupContextWindow() {
				if imgui.MenuItemV("Custom", "", f.overlayCorner == -1, true) {
					f.overlayCorner = -1
				}
				if imgui.MenuItemV("Top-left", "", f.overlayCorner == 0, true) {
					f.overlayCorner = 0
				}
				if imgui.MenuItemV("Top-right", "", f.overlayCorner == 1, true) {
					f.overlayCorner = 1
				}
				if imgui.MenuItemV("Bottom-left", "", f.overlayCorner == 2, true) {
					f.overlayCorner = 2
				}
				if imgui.MenuItemV("Bottom-right", "", f.overlayCorner == 3, true) {
					f.overlayCorner = 3
				}

				imgui.EndPopup()
			}
		}
		imgui.End()
	}

	if f.showAboutWindow {
		imgui.BeginV("About Forgery", &f.showAboutWindow, imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoResize)
		imgui.Text("Forgery - Open source Hammer (r) replacement.")
		imgui.Text("Written by Emily Hudson.")
		imgui.Text("Thanks to Galaco for providing a large amount of backend code.")
		imgui.End()
	}
}

func (f *ForgeryContext) Run() {
	clearColor := [4]float32{0.1, 0.1, 0.1, 1.0}

	for !f.platform.ShouldStop() {
		f.platform.ProcessEvents()

		// Render all our viewports
		f.RenderScene()

		// Signal start of a new frame
		f.platform.NewFrame()
		imgui.NewFrame()

		// Render our ui
		f.RenderUI()

		// Create the draw lists
		imgui.Render()

		f.imguiRenderer.PreRender(clearColor)
		// Actually draw the imgui data
		f.imguiRenderer.Render(f.platform.DisplaySize(), f.platform.FramebufferSize(), imgui.RenderedDrawData())
		f.platform.PostRender()

		// @TODO: we should get some method of capping fps
		// but this is not that!
		// <-time.After(time.Millisecond * 10)
	}
}

func (f *ForgeryContext) NewSceneWindow() {
	// 1024 is the framebuffer size
	f.lastSceneWindowId++
	newWindow := windows.NewSceneWindow(f.filesystem,
		f.adapter,
		f.render,
		f.scene,
		4000, 4000,
		&f.cameraSens,
		&f.cameraMoveSens,
		f.lastSceneWindowId,
		f.platform)

	newWindow.Initialize()

	f.sceneWindows = append(f.sceneWindows, newWindow)
}

func (f *ForgeryContext) NewApp() {
	f.context = imgui.CreateContext(nil)
	io := imgui.CurrentIO()
	io.SetConfigFlags(imgui.ConfigFlagNavEnableKeyboard)

	logger.EnablePretty()
	logger.SetWriter(&stdOut{})

	imgFile, err := os.Open("assets/icon.png")
	defer imgFile.Close()

	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(imgFile)

	platform, err := native.NewGLFW(io, "Forgery", img)
	if err != nil {
		panic(err)
	}

	imguiRenderer, err := render.NewOpenGL3(io)
	if err != nil {
		panic(err)
	}

	// @TODO We need to load our settings from a file somehow!

	f.cameraSens = 4
	f.cameraMoveSens = 4

	f.filesystem = valve.NewFileSystem("E:\\steam\\steamapps\\common\\Counter-Strike Global Offensive\\csgo")
	valve.DumpAllKnownMaterials(f.filesystem)

	f.adapter = &adapters.OpenGL{}
	f.adapter.Init()

	f.render = render.NewRenderer(f.adapter)

	cache.InitTextureLookup()

	// @TODO Dont do this on the same thread doh
	f.texturesLoadingCompleteChan = make(chan struct{}, 1000)
	f.texturesLoadedExpected = cache.LoadAllKnownMaterials(f.filesystem, f.texturesLoadingCompleteChan)

	f.platform = platform
	f.imguiRenderer = imguiRenderer

	// TODO: Dont just load this default
	{
		newMap, err := formats.LoadVmf("assets/default_cs.vmf")

		if err != nil {
			panic(err)
		}

		f.activeMap = newMap
		f.documentLoaded = true
		f.scene = view.NewSceneFromVmf(f.activeMap)
	}

}

func (f *ForgeryContext) DestroyApp() {
	f.imguiRenderer.Dispose()

	f.platform.Dispose()
	f.context.Destroy()
}
