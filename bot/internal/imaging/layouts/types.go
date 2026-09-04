package layouts

import "image/color"

const (
	FontPrimary = 0
	FontMono    = 1
)

// Canvas defines the rendering primitives available to all layout files.
// This interface breaks the circular dependency between the imaging engine and layouts.
type Canvas interface {
	DrawText(x, y float64, text string, fontType int, size float64, clr color.Color)
	DrawTextAnchored(x, y float64, text string, fontType int, size float64, clr color.Color, anchorX, anchorY float64)
	DrawCurrencyValue(x, y float64, amount int64, clr color.Color)
	DrawProgressBar(x, y, width, height, percent float64, clr color.Color)
	DrawMiniChart(x, y float64, values []float64, clr color.Color, width, height float64)
	DrawAvatar(x, y, size float64, imageURL string)
}