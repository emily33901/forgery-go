package view

import (
	"github.com/emily33901/go-forgery/render"
)

type fbo struct {
	adapter       render.Adapter
	framebuffer   uint32
	colourTexture uint32
	depthTexture  uint32
	width         int
	height        int
}

// Resize resizes this framebuffer object
func (win *fbo) Resize(width int, height int) {
	win.width = width
	win.height = height

	win.Bind()

	if win.colourTexture != 0 {
		win.adapter.DeleteTextures(1, &win.colourTexture)
		win.adapter.DeleteRenderBuffer(1, &win.depthTexture)
	}

	win.depthTexture = win.adapter.CreateRenderbufferStorageDepth(int32(win.width), int32(win.height))

	win.adapter.CreateTextureStorage2D(&win.colourTexture, int32(win.width), int32(win.height))
	win.adapter.BindTexture2D(win.colourTexture)
	win.adapter.BindTexture2DToFramebuffer(win.colourTexture)
	win.adapter.BindDepthBufferToFramebuffer(win.depthTexture)
	win.adapter.DrawBuffers()
	win.adapter.ClearColor(0, 0, 0, 0)
	win.adapter.ClearAll()
	win.adapter.BindTexture2D(0)

	win.Unbind()
}

// Bind this framebuffer
func (win *fbo) Bind() {
	win.adapter.BindFramebuffer(win.framebuffer)
}

// Unbind unbind this framebuffer
func (win *fbo) Unbind() {
	win.adapter.BindFramebuffer(0)
}

// Destroy deletes and cleans up this framebuffer
func (win *fbo) Destroy() {
	win.adapter.DeleteFramebuffers(1, &win.framebuffer)
}

// NewFbo returns a new framebuffer
func NewFbo(adapter render.Adapter, width int, height int) *fbo {
	f := &fbo{
		adapter: adapter,
	}
	f.adapter.CreateFramebuffers(1, &f.framebuffer)
	f.Resize(width, height)
	return f
}
