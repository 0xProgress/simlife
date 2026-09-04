package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/rs/zerolog"
	"golang.org/x/image/font"

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
	assets   *Assets
	adChance float64 // Probability (0.0 to 1.0) of showing an ad instead of the requested layout
	log      zerolog.Logger
}

// NewComposer initializes the composer with pre-loaded assets and ad configuration.
func NewComposer(assets *Assets, adChance float64, log zerolog.Logger) *Composer {
	return &Composer{
		assets:   assets,
		adChance: adChance,
		log:      log.With().Str("component", "imaging").Logger(),
	}
}

// Compose generates a PNG byte slice for the given layout and data.
// If compositing fails, it returns an error so the caller can fall back to a plain text embed.
func (c *Composer) Compose(layout string, data any) ([]byte, error) {
	// Ad Injection Logic: Randomly replace the requested layout with an ad
	if c.adChance > 0 && rand.Float64() < c.adChance && layout != "ad" {
		c.log.Debug().Str("original_layout", layout).Msg("injecting ad layout")
		layout = "ad"
		data = layouts.AdData{
			Title:   "Aether City with Premium",
			Message: "Enjoy an ad-free experience and exclusive cosmetics with City Pass!",
			Action:  "Use /premium to upgrade",
			Color:   "#d4a847", // Gold accent
		}
	}

	bg, ok := c.assets.Backgrounds[layout]
	if !ok {
		c.log.Warn().Str("layout", layout).Msg("missing background asset, generating programmatic fallback")
		bg = c.generateFallbackBackground(layout)
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
	case "ad":
		err = layouts.RenderAd(renderer, data)
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

// generateFallbackBackground creates a programmatic gradient background if the PNG is missing.
// This ensures the bot never crashes due to missing assets and provides a clean, intentional look.
func (c *Composer) generateFallbackBackground(layout string) image.Image {
	dc := gg.NewContext(1200, 630)

	// Dark urban economy gradient (matches design tokens)
	gradient := gg.NewLinearGradient(0, 0, 1200, 630)
	gradient.AddColorStop(0, color.RGBA{15, 17, 23, 255})   // --bg-base
	gradient.AddColorStop(1, color.RGBA{30, 35, 50, 255})   // Slightly lighter dark

	dc.SetFillStyle(gradient)
	dc.Rectangle(0, 0, 1200, 630)
	dc.Fill()

	// Add a subtle accent line at the top
	dc.SetHexColor("#d4a847") // --accent-gold
	dc.Rectangle(0, 0, 1200, 4)
	dc.Fill()

	// Render layout name as a subtle watermark
	if c.assets.PrimaryFont != nil {
		face := truetype.NewFace(c.assets.PrimaryFont, &truetype.Options{Size: 72, DPI: 72, Hinting: font.HintingFull})
		dc.SetFontFace(face)
		dc.SetRGB(0.15, 0.15, 0.15) // Very dark gray for subtle watermark
		dc.DrawStringAnchored(fmt.Sprintf("%s LAYOUT", strings.ToUpper(layout)), 600, 315, 0.5, 0.5)
	}

	return dc.Image()
}

// LoadAssets loads all backgrounds and fonts from disk into memory at startup.
func LoadAssets(bgDir, fontDir string) (*Assets, error) {
	assets := &Assets{
		Backgrounds: make(map[string]image.Image),
	}

	// Load backgrounds (including the new 'ad' layout)
	bgFiles := []string{"balance", "business", "market", "property", "profile", "economic_news", "shop", "ad"}
	for _, name := range bgFiles {
		path := filepath.Join(bgDir, name+"_bg.png")
		img, err := gg.LoadImage(path)
		if err != nil {
			// We don't return an error here; Compose will use generateFallbackBackground
			// But we log it at startup so the developer knows which assets are missing.
			fmt.Printf("Warning: Could not load background %s: %v (will use fallback)\n", path, err)
		} else {
			assets.Backgrounds[name] = img
		}
	}

	// Load fonts
	primaryPath := filepath.Join(fontDir, "primary.ttf")
	monoPath := filepath.Join(fontDir, "mono.ttf")

	primaryData, err := os.ReadFile(primaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read primary font: %w", err)
	}
	primaryFont, err := truetype.Parse(primaryData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse primary font: %w", err)
	}
	assets.PrimaryFont = primaryFont

	monoData, err := os.ReadFile(monoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mono font: %w", err)
	}
	monoFont, err := truetype.Parse(monoData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mono font: %w", err)
	}
	assets.MonoFont = monoFont

	return assets, nil
}