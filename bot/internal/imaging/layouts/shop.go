package layouts

import (
	"fmt"
	"image/color"

	"github.com/shopspring/decimal"
)

// ShopItem represents a single item in the global shop.
// STRICT RULE: No float64 or int64 is used for financial values.
type ShopItem struct {
	Name  string
	Price decimal.Decimal
}

// ShopData defines the data structure for the shop listing image compositor.
type ShopData struct {
	Items []ShopItem
}

// RenderShop draws the global shopfront listing onto the canvas.
func RenderShop(c Canvas, data any) error {
	d, ok := data.(ShopData)
	if !ok {
		return fmt.Errorf("invalid data type for shop layout: expected ShopData")
	}

	c.DrawTextAnchored(600, 60, "GLOBAL MARKET SHOPFRONT", FontPrimary, 42, color.White, 0.5, 0.5)

	y := 140.0
	for _, item := range d.Items {
		c.DrawText(100, y, item.Name, FontPrimary, 28, color.White)
		c.DrawCurrencyValue(1100, y, item.Price, color.RGBA{255, 215, 0, 255}) // Gold for prices
		y += 60
	}

	c.DrawTextAnchored(600, 580, "Prices update dynamically at daily settlement", FontPrimary, 18, color.RGBA{150, 150, 150, 255}, 0.5, 0.5)
	
	return nil
}