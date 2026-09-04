package layouts

import (
	"fmt"
	"image/color"

	"github.com/shopspring/decimal"
)

// BusinessData defines the data structure for the business status image compositor.
type BusinessData struct {
	Name             string
	Type             string
	Status           string // PRODUCING, STALLED, CLOSED
	ProgressPercent  float64 // Allowed: visual percentage, not a financial value
	WorkerCount      int
	ProjectedRevenue decimal.Decimal
	ProjectedCost    decimal.Decimal
}

// RenderBusiness draws the business operational status onto the canvas.
func RenderBusiness(c Canvas, data any) error {
	d, ok := data.(BusinessData)
	if !ok {
		return fmt.Errorf("invalid data type for business layout: expected BusinessData")
	}

	c.DrawText(60, 80, d.Name, FontPrimary, 48, color.White)
	c.DrawText(60, 130, d.Type, FontPrimary, 24, color.RGBA{200, 200, 200, 255})

	statusColor := color.RGBA{76, 175, 118, 255} // Green
	if d.Status == "STALLED" {
		statusColor = color.RGBA{255, 165, 0, 255} // Orange
	}
	if d.Status == "CLOSED" {
		statusColor = color.RGBA{224, 85, 85, 255} // Red
	}
	c.DrawText(60, 550, d.Status, FontMono, 32, statusColor)

	c.DrawProgressBar(60, 220, 1080, 24, d.ProgressPercent, color.RGBA{74, 144, 217, 255}) // Blue

	c.DrawText(60, 300, "WORKERS", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 340, fmt.Sprintf("%d", d.WorkerCount), FontMono, 48, color.White)

	c.DrawText(400, 300, "PROJECTED REVENUE", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(400, 340, d.ProjectedRevenue, color.RGBA{76, 175, 118, 255}) // Green

	c.DrawText(800, 300, "PROJECTED COST", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(800, 340, d.ProjectedCost, color.RGBA{224, 85, 85, 255}) // Red

	return nil
}