package render

import (
	"unsafe"

	"github.com/emily33901/gosigl"
)

// Adapter is an interface for graphics api agnostic gpu function calls
type Adapter interface {
	Init()

	// Simplified
	CreateTexture2D(id *uint32, width, height int32, data []byte)
	CreateTextureStorage2D(id *uint32, width, height int32)
	BindTexture2DToFramebuffer(framebufferId uint32)
	BindFramebuffer(framebufferId uint32)
	BindColourBufferToFramebuffer(framebufferId uint32)
	BindDepthBufferToFramebuffer(framebufferId uint32)
	BindTexture2D(id uint32)
	DrawBuffers()
	LoadSimpleShader(path string) *gosigl.Context
	CreateRenderbufferStorageMultiSampledColour(width, height int32) uint32
	CreateRenderbufferStorageMultiSampledDepth(width, height int32) uint32
	CreateRenderbufferStorageColour(width, height int32) uint32
	CreateRenderbufferStorageDepth(width, height int32) uint32

	// General
	Viewport(x, y, width, height int32)
	ClearColor(r, g, b, a float32)
	Clear(mask uint32)
	ClearAll()

	// Framebuffer
	CreateFramebuffers(n int32, framebuffers *uint32)
	BindFramebufferInternal(target uint32, framebuffer uint32)
	DeleteFramebuffers(n int32, framebuffers *uint32)
	FramebufferTexture2D(target uint32, attachment uint32, textarget uint32, texture uint32, level int32)
	DeleteRenderBuffer(n int32, target *uint32)

	// Texture
	DeleteTextures(n int32, textures *uint32)
	GenTextures(n int32, textures *uint32)
	BindTexture(target uint32, texture uint32)
	TexImage2D(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels unsafe.Pointer)
	TexParameteri(target uint32, pname uint32, param int32)

	// Drawing
	DrawTriangleArray(offset int32, count int32)

	// Misc
	EnableBlend()
	EnableDepthTest()
	EnableCullFaceBack()

	// Uniforms
	SendUniformMat4(uniform int32, matrix *float32)
	SendUniformFloat(uniform int32, float *float32)
	SendUniformVec3(uniform int32, float *float32)
	SendUniformVec2(uniform int32, float *float32)
	SendUniformVec4(uniform int32, float *float32)
	SendUniformBool(uniform int32, b bool)

	Error() bool
}
