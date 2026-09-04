package layouts

import (
	"fmt"
	"image/color"
)

type MarketData struct {
	ItemType      string
	Quantity      int
	PricePerUnit  int64
	ExpiryHours   int
	HistoryValues []float64
}

func RenderMarket(c Canvas, data any) error {
	d, ok := data.(MarketData)
	if !ok {
		return fmt.Errorf("invalid data type for market layout")
	}

	c.DrawText(60, 80, d.ItemType, FontPrimary, 48, color.White)

	c.DrawText(60, 180, "QUANTITY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 220, fmt.Sprintf("%d", d.Quantity), FontMono, 36, color.White)

	c.DrawText(300, 180, "ASKING PRICE", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(300, 220, d.PricePerUnit, color.RGBA{255, 215, 0, 255})

	c.DrawText(600, 180, "EXPIRES IN", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(600, 220, fmt.Sprintf("%dh", d.ExpiryHours), FontMono, 36, color.White)

	if len(d.HistoryValues) > 0 {
		c.DrawText(60, 320, "30-DAY PRICE HISTORY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
		c.DrawMiniChart(60, 360, d.HistoryValues, color.RGBA{0, 200, 255, 255}, 1080, 200)
	}

	return nil
}