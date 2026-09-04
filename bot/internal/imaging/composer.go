package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"

	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// Assets holds all pre-loaded static resources required for compositing.
// These are loaded once at startup to ensure zero disk I/O during request handling.
type Assets struct {
	Backgrounds map[string]image.Image
	PrimaryFont *truetype.Font
	MonoFont    *truetype.Font
}

// Composer is the core compositing engine.
type Composer struct {
	assets *Assets
}

// NewComposer initializes the composer with pre-loaded assets.
func NewComposer(assets *Assets) *Composer {
	return &Composer{assets: assets}
}

// Compose generates a PNG byte slice for the given layout and data.
// If compositing fails, it returns an error so the caller can fall back to a plain text embed.
func (c *Composer) Compose(layout string, data any) ([]byte, error) {
	bg, ok := c.assets.Backgrounds[layout]
	if !ok {
		return nil, fmt.Errorf("unknown layout or missing background asset: %s", layout)
	}

	// Standard Discord embed image ratio
	dc := gg.NewContext(1200, 630)
	dc.DrawImage(bg, 0, 0)

	renderer := NewRenderer(dc, c.assets)

	var err error
	switch layout {
	case "balance":
		err = layouts.RenderBalance(renderer, data)
	case "business":
		err = layouts.RenderBusiness(renderer, data)
	case "market":
		err = layouts.RenderMarket(renderer, data)
	case "property":
		err = layouts.RenderProperty(renderer, data)
	case "profile":
		err = layouts.RenderProfile(renderer, data)
	case "economic_news":
		err = layouts.RenderEconomicNews(renderer, data)
	case "shop":
		err = layouts.RenderShop(renderer, data)
	default:
		return nil, fmt.Errorf("layout %s not implemented", layout)
	}

	if err != nil {
		return nil, fmt.Errorf("layout rendering failed for %s: %w", layout, err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}