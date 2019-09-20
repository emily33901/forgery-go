package windows

import (
	"github.com/emily33901/go-forgery/render/cache"
	"github.com/emily33901/imgui-go"
	"github.com/emily33901/lambda-core/core/filesystem"
)

func RenderMaterialsWindow(fs filesystem.IFileSystem, shouldOpen *bool, materialSelected func(string)) {
	if imgui.BeginV("Materials", shouldOpen, 0) {
		areaAvailable := imgui.ContentRegionAvail()
		if imgui.BeginChildV("Materials Scrollable", areaAvailable.Minus(imgui.Vec2{0, 100}), false, 0) {
			startCursorPos := imgui.ScrollY()
			contentSize := imgui.ContentRegionAvail()
			contentEnd := startCursorPos + contentSize.Y

			const thumbSize = 128
			const thumbPad = 4
			const totalSize = thumbSize + thumbPad

			// Since there are so many textures that want to be loaded here
			// We are going to clip it to the visible ones

			rowi := 0
			coli := 0

			// TODO we should be caching these results!!!!
			materials := cache.GetMaterials()

			for _, k := range materials {
				if float32((rowi+4)*totalSize) > startCursorPos &&
					float32((rowi-2)*totalSize) < contentEnd {

					imgui.PushID(k)
					imgui.PushTextWrapPosV(float32((coli + 1) * totalSize))
					{
						// Image
						pos := imgui.CursorPos()
						imgui.Image(cache.OglToImguiTextureId(uint32(cache.LookupTexture(fs, k))), imgui.Vec2{thumbSize, thumbSize})

						// Text
						imgui.SetCursorPos(pos)
						imgui.Text(k)

						// Selectable
						imgui.SetCursorPos(pos)
						if imgui.SelectableV("", false, 0, imgui.Vec2{thumbSize, thumbSize}) {
							materialSelected(k)
						}
					}
					imgui.PopTextWrapPos()
					imgui.PopID()
				} else {
					imgui.Dummy(imgui.Vec2{thumbSize, thumbSize})
				}
				oldCursor := imgui.CursorPos()

				if contentSize.X < float32((coli+2)*totalSize) {
					imgui.SetCursorPos(oldCursor)
					coli = 0
					rowi++
				} else {
					imgui.SameLine()
					coli++
				}
			}
		}
		imgui.EndChild()

		imgui.Separator()

		imgui.Text("wow nice meme")
		imgui.Text("wow nice meme")
		imgui.Text("wow nice meme")
	}
	imgui.End()
}
