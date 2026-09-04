package imaging

import (
	"fmt"
	"image/color"
	"net/http"
	"time"

	imgutil "github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// Renderer implements the layouts.Canvas interface using the gg drawing context.
type Renderer struct {
	dc     *gg.Context
	assets *Assets
}

// NewRenderer creates a new rendering context for a specific image composition.
func NewRenderer(dc *gg.Context, assets *Assets) *Renderer {
	return &Renderer{dc: dc, assets: assets}
}

func (r *Renderer) setFont(fontType int, size float64) {
	var ttf *truetype.Font
	if fontType == layouts.FontMono {
		ttf = r.assets.MonoFont
	} else {
		ttf = r.assets.PrimaryFont
	}

	if ttf == nil {
		return
	}

	face := truetype.NewFace(ttf, &truetype.Options{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	r.dc.SetFontFace(face)
}

func (r *Renderer) DrawText(x, y float64, text string, fontType int, size float64, clr color.Color) {
	r.setFont(fontType, size)
	r.dc.SetColor(clr)
	r.dc.DrawString(text, x, y)
}

func (r *Renderer) DrawTextAnchored(x, y float64, text string, fontType int, size float64, clr color.Color, anchorX, anchorY float64) {
	r.setFont(fontType, size)
	r.dc.SetColor(clr)
	r.dc.DrawStringAnchored(text, x, y, anchorX, anchorY)
}

func (r *Renderer) DrawCurrencyValue(x, y float64, amount int64, clr color.Color) {
	r.setFont(layouts.FontMono, 48)
	r.dc.SetColor(clr)
	r.dc.DrawString(fmt.Sprintf("$%s", formatCurrency(amount)), x, y)
}

func (r *Renderer) DrawProgressBar(x, y, width, height, percent float64, clr color.Color) {
	if percent > 1.0 {
		percent = 1.0
	}
	if percent < 0.0 {
		percent = 0.0
	}

	// Background track
	r.dc.SetHexColor("#222222")
	r.dc.DrawRoundedRectangle(x, y, width, height, height/2)
	r.dc.Fill()

	// Fill track
	fillWidth := width * percent
	if fillWidth > 0 {
		r.dc.SetColor(clr)
		r.dc.DrawRoundedRectangle(x, y, fillWidth, height, height/2)
		r.dc.Fill()
	}
}

func (r *Renderer) DrawMiniChart(x, y float64, values []float64, clr color.Color, width, height float64) {
	if len(values) < 2 {
		return
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	rng := max - min
	if rng == 0 {
		rng = 1 // Prevent division by zero for flat lines
	}

	stepX := width / float64(len(values)-1)

	r.dc.SetColor(clr)
	r.dc.SetLineWidth(3)

	for i := 0; i < len(values)-1; i++ {
		x1 := x + float64(i)*stepX
		y1 := y + height - ((values[i]-min)/rng)*height
		x2 := x + float64(i+1)*stepX
		y2 := y + height - ((values[i+1]-min)/rng)*height
		r.dc.DrawLine(x1, y1, x2, y2)
		r.dc.Stroke()
	}
}

func (r *Renderer) DrawAvatar(x, y, size float64, imageURL string) {
	// Strict timeout to ensure compositing never exceeds 30ms target
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		// Fallback placeholder
		r.dc.SetHexColor("#555555")
		r.dc.DrawCircle(x+size/2, y+size/2, size/2)
		r.dc.Fill()
		return
	}
	defer resp.Body.Close()

	img, err := imgutil.Decode(resp.Body)
	if err != nil {
		return
	}

	resized := imgutil.Resize(img, int(size), int(size), imgutil.Lanczos)

	// Clip to circle
	r.dc.Push()
	r.dc.DrawCircle(x+size/2, y+size/2, size/2)
	r.dc.Clip()
	r.dc.DrawImage(resized, int(x), int(y))
	r.dc.Pop()
}

// formatCurrency adds comma separators to large integers.
func formatCurrency(amount int64) string {
	if amount < 0 {
		return "-" + formatCurrency(-amount)
	}
	str := fmt.Sprintf("%d", amount)
	if len(str) <= 3 {
		return str
	}

	var res []byte
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			res = append(res, ',')
		}
		res = append(res, byte(c))
	}
	return string(res)
}
