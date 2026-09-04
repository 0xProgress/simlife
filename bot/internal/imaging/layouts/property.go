package layouts

import (
	"fmt"
	"image/color"
)

type PropertyData struct {
	Coordinates      string
	ZoneType         string
	DevelopmentLevel int
	AssessedValue    int64
	LastTaxPaid      string
}

func RenderProperty(c Canvas, data any) error {
	d, ok := data.(PropertyData)
	if !ok {
		return fmt.Errorf("invalid data type for property layout")
	}

	c.DrawText(60, 80, "PROPERTY DEED", FontPrimary, 48, color.White)
	c.DrawText(60, 140, d.Coordinates, FontMono, 24, color.RGBA{200, 200, 200, 255})

	c.DrawText(60, 220, "ZONE", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 260, d.ZoneType, FontPrimary, 32, color.White)

	c.DrawText(400, 220, "DEVELOPMENT LVL", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(400, 260, fmt.Sprintf("%d", d.DevelopmentLevel), FontMono, 32, color.White)

	c.DrawText(60, 360, "ASSESSED VALUE", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 400, d.AssessedValue, color.RGBA{255, 215, 0, 255})

	c.DrawText(60, 500, "LAST TAX PAID", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 540, d.LastTaxPaid, FontMono, 24, color.White)

	return nil
}