package layouts

import (
	"fmt"
	"image/color"
)

type EconomicNewsData struct {
	EconDay     int
	PriceIndex  float64
	IndexChange float64
	Velocity    float64
	GiniCoeff   float64
	TopEarners  []string
	MoneySupply int64
}

func RenderEconomicNews(c Canvas, data any) error {
	d, ok := data.(EconomicNewsData)
	if !ok {
		return fmt.Errorf("invalid data type for economic news layout")
	}

	c.DrawTextAnchored(600, 80, fmt.Sprintf("DAY %d ECONOMIC REPORT", d.EconDay), FontPrimary, 48, color.White, 0.5, 0.5)

	c.DrawText(60, 180, "PRICE INDEX", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 220, fmt.Sprintf("%.2f", d.PriceIndex), FontMono, 36, color.White)

	changeColor := color.RGBA{0, 255, 0, 255}
	if d.IndexChange < 0 {
		changeColor = color.RGBA{255, 69, 0, 255}
	}
	c.DrawText(250, 220, fmt.Sprintf("(%.2f%%)", d.IndexChange), FontMono, 24, changeColor)

	c.DrawText(60, 300, "VELOCITY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawProgressBar(60, 330, 400, 20, d.Velocity/10.0, color.RGBA{0, 200, 255, 255})

	c.DrawText(600, 300, "GINI COEFFICIENT", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawProgressBar(600, 330, 400, 20, d.GiniCoeff, color.RGBA{255, 165, 0, 255})

	c.DrawText(60, 420, "TOP EARNERS TODAY", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	y := 460.0
	for i, earner := range d.TopEarners {
		c.DrawText(60, y, fmt.Sprintf("%d. %s", i+1, earner), FontPrimary, 24, color.White)
		y += 30
	}

	c.DrawTextAnchored(600, 580, fmt.Sprintf("TOTAL MONEY SUPPLY: $%s", formatCurrencySimple(d.MoneySupply)), FontMono, 24, color.RGBA{255, 215, 0, 255}, 0.5, 0.5)

	return nil
}

func formatCurrencySimple(amount int64) string {
	if amount < 0 {
		return "-" + formatCurrencySimple(-amount)
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