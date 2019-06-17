package view

import (
	"github.com/emily33901/go-forgery/render"
)

// RenderWindow provides a broader wrapper for a framebuffer.
type RenderWindow struct {
	adapter     render.Adapter
	width       int
	height      int
	frameBuffer *fbo
}

// BufferId returns the RenderWindows bound FBO id.
func (win *RenderWindow) BufferId() uint32 {
	return win.frameBuffer.colourTexture
}

// Bind binds this RenderWindow(s fbo)
// @TODO shouldnt this bind to the size of the target texture and not the size of the framebuffer
func (win *RenderWindow) Bind(width, height float32) {
	win.adapter.Viewport(0, 0, int32(width), int32(height))
	win.frameBuffer.Bind()

	win.adapter.ClearAll()
}

// Unbind this RenderWindow(s fbo)
func (win *RenderWindow) Unbind() {
	win.frameBuffer.Unbind()
}

func (win *RenderWindow) Width() int {
	return win.width
}

// SetSize resizes the bound fbo
func (win *RenderWindow) SetSize(width int, height int) {
	win.width = width
	win.height = height
	win.frameBuffer.Destroy()
	win.frameBuffer = NewFbo(win.adapter, width, height)
}

// Close cleans up and destroys this RenderWindow.
func (win *RenderWindow) Close() {
	win.frameBuffer.Destroy()
}

// NewRenderWindow returns a new RenderWindow
func NewRenderWindow(adapter render.Adapter, width int, height int) *RenderWindow {
	r := &RenderWindow{
		adapter:     adapter,
		width:       width,
		height:      height,
		frameBuffer: nil,
	}
	r.frameBuffer = NewFbo(adapter, width, height)

	return r
}
