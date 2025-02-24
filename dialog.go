//go:build windows || linux || darwin

package main

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// DisplayQRCode shows a QR code in a native window
func DisplayQRCode(title string, qrImg image.Image) {
	w := new(app.Window)
	done := make(chan struct{})
	w.Option(app.Title(title))
	w.Option(app.Size(unit.Dp(300), unit.Dp(300)))

	qrOp := paint.NewImageOp(qrImg)

	go func() {
		var ops op.Ops

	loop:
		for {
			e := w.Event()
			switch e := e.(type) {
			case app.DestroyEvent:
				break loop
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widget.Image{
								Src:   qrOp,
								Scale: 1,
								Fit:   widget.Contain,
							}.Layout(gtx)
						})
					}),
				)

				e.Frame(gtx.Ops)
			}
		}
		close(done)
	}()

	<-done // Wait for window to close
}
