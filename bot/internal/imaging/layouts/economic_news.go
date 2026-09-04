package layouts

import (
	"fmt"
	"image/color"

	"github.com/shopspring/decimal"
)

// EconomicNewsData defines the data structure for the daily economic report image.
type EconomicNewsData struct {
	EconDay     int
	PriceIndex  float64 // Allowed: statistical index, not a direct currency amount
	IndexChange float64 // Allowed: statistical percentage
	Velocity    float64 // Allowed: statistical ratio
	GiniCoeff   float64 // Allowed: statistical ratio
	TopEarners  []string
	MoneySupply decimal.Decimal // STRICT: Financial aggregate must be decimal
}

// RenderEconomicNews draws the daily economic bulletin onto the canvas.
func RenderEconomicNews(c Canvas, data any) error {
	d, ok := data.(EconomicNewsData)
	if !ok {
		return fmt.Errorf("invalid data type for economic news layout: expected EconomicNewsData")
	}

	c.DrawTextAnchored(600, 80, fmt.Sprintf("DAY %d ECONOMIC REPORT", d.EconDay), FontPrimary, 48, color.White, 0.5, 0.5)

	c.DrawText(60, 180, "PRICE INDEX", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 220, fmt.Sprintf("%.2f", d.PriceIndex), FontMono, 36, color.White)
	
	changeColor := color.RGBA{76, 175, 118, 255} // Green
	if d.IndexChange < 0 {
		changeColor = color.RGBA{224, 85, 85, 255} // Red
	}
	c.DrawText(250, 220, fmt.Sprintf("(%.2f%%)", d.IndexChange), FontMono, 24, changeColor)

	c.DrawText(60, 300, "VELOCITY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawProgressBar(60, 330, 400, 20, d.Velocity/10.0, color.RGBA{74, 144, 217, 255}) // Blue

	c.DrawText(600, 300, "GINI COEFFICIENT", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawProgressBar(600, 330, 400, 20, d.GiniCoeff, color.RGBA{255, 165, 0, 255}) // Orange

	c.DrawText(60, 420, "TOP EARNERS TODAY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	y := 460.0
	for i, earner := range d.TopEarners {
		c.DrawText(60, y, fmt.Sprintf("%d. %s", i+1, earner), FontPrimary, 24, color.White)
		y += 30
	}

	c.DrawTextAnchored(600, 580, fmt.Sprintf("TOTAL MONEY SUPPLY: ⊄%s", d.MoneySupply.String()), FontMono, 24, color.RGBA{255, 215, 0, 255}, 0.5, 0.5)

	return nil
}