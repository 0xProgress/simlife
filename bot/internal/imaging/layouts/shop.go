package layouts

import (
	"fmt"
	"image/color"
)

type ShopItem struct {
	Name  string
	Price int64
}

type ShopData struct {
	Items []ShopItem
}

func RenderShop(c Canvas, data any) error {
	d, ok := data.(ShopData)
	if !ok {
		return fmt.Errorf("invalid data type for shop layout")
	}

	c.DrawTextAnchored(600, 60, "GLOBAL MARKET SHOPFRONT", FontPrimary, 42, color.White, 0.5, 0.5)

	y := 140.0
	for _, item := range d.Items {
		c.DrawText(100, y, item.Name, FontPrimary, 28, color.White)
		c.DrawTextAnchored(1100, y, fmt.Sprintf("$%d", item.Price), FontMono, 28, color.RGBA{255, 215, 0, 255}, 1.0, 0.5)
		y += 60
	}

	c.DrawTextAnchored(600, 580, "Prices update dynamically at daily settlement", FontPrimary, 18, color.RGBA{150, 150, 150, 255}, 0.5, 0.5)

	return nil
}