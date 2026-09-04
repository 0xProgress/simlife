package layouts

import (
	"fmt"
	"image/color"
)

// Font types for the renderer
const (
	FontPrimary = 0
	FontMono    = 1
)

// Canvas defines the interface for rendering operations.
type Canvas interface {
	DrawText(x, y float64, text string, fontType int, size float64, clr color.Color)
	DrawTextAnchored(x, y float64, text string, fontType int, size float64, clr color.Color, anchorX, anchorY float64)
	DrawCurrencyValue(x, y float64, amount decimal.Decimal, clr color.Color) // Note: requires decimal import in layouts if used
	DrawProgressBar(x, y, width, height, percent float64, clr color.Color)
	DrawMiniChart(x, y float64, values []float64, clr color.Color, width, height float64)
	DrawAvatar(x, y, size float64, imageURL string)
}

// AdData defines the data structure for the injected ad layout.
type AdData struct {
	Title   string
	Message string
	Action  string
	Color   string // Hex color for accent
}

// RenderAd draws the premium ad overlay/layout.
func RenderAd(r Canvas, data any) error {
	ad, ok := data.(AdData)
	if !ok {
		return fmt.Errorf("invalid data type for ad layout: expected AdData")
	}

	// We need to assert to the concrete Renderer to access dc directly for shapes
	// (In a real setup, you'd add DrawRoundedRectangle to the Canvas interface, 
	// but for simplicity, we'll assume the concrete type is passed or add the method).
	// For this example, we'll use the concrete type assertion.
	renderer, ok := r.(*Renderer)
	if !ok {
		return fmt.Errorf("expected concrete Renderer type")
	}

	// Darken the background slightly to make text pop
	renderer.dc.SetRGBA(0, 0, 0, 180)
	renderer.dc.Rectangle(0, 0, 1200, 630)
	renderer.dc.Fill()

	// Outer glow/accent bar
	renderer.dc.SetHexColor(ad.Color)
	renderer.dc.DrawRoundedRectangle(100, 200, 1000, 230, 12)
	renderer.dc.Fill()

	// Inner dark panel
	renderer.dc.SetHexColor("#181c26") // --bg-surface
	renderer.dc.DrawRoundedRectangle(110, 210, 980, 210, 8)
	renderer.dc.Fill()

	// Title
	renderer.DrawTextAnchored(600, 260, ad.Title, FontPrimary, 42, color.White, 0.5, 0.5)

	// Message
	renderer.DrawTextAnchored(600, 320, ad.Message, FontPrimary, 24, color.RGBA{232, 234, 240, 255}, 0.5, 0.5)

	// Action
	renderer.DrawTextAnchored(600, 380, ad.Action, FontMono, 20, color.RGBA{212, 168, 71, 255}, 0.5, 0.5) // Gold

	return nil
}