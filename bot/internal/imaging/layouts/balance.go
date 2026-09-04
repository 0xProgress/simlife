package layouts

import (
	"fmt"
	"image/color"

	"github.com/shopspring/decimal"
)

// BalanceData defines the data structure for the balance image compositor.
// STRICT RULE: No float64 or int64 is used for financial values.
type BalanceData struct {
	Username    string
	AvatarURL   string
	Wallet      decimal.Decimal
	Bank        decimal.Decimal
	Escrow      decimal.Decimal
	NetWorth    decimal.Decimal
	Change24h   decimal.Decimal
	RankPercent int
	EconDay     int
}

// RenderBalance draws the player's financial standing onto the canvas.
func RenderBalance(c Canvas, data any) error {
	d, ok := data.(BalanceData)
	if !ok {
		return fmt.Errorf("invalid data type for balance layout: expected BalanceData")
	}

	c.DrawAvatar(40, 40, 80, d.AvatarURL)
	c.DrawText(140, 90, d.Username, FontPrimary, 32, color.White)

	c.DrawText(60, 220, "WALLET", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 280, d.Wallet, color.RGBA{255, 215, 0, 255}) // Gold

	c.DrawText(60, 360, "BANK", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 400, d.Bank, color.White)

	c.DrawText(60, 500, "NET WORTH", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 550, d.NetWorth, color.White)

	// 24h change indicator
	changeColor := color.RGBA{76, 175, 118, 255} // Green
	if d.Change24h.LessThan(decimal.Zero) {
		changeColor = color.RGBA{224, 85, 85, 255} // Red
	}

	changeStr := d.Change24h.String() + "%"
	if d.Change24h.GreaterThan(decimal.Zero) {
		changeStr = "+" + changeStr
	}

	c.DrawTextAnchored(1140, 550, changeStr, FontMono, 36, changeColor, 1.0, 0.5)
	c.DrawTextAnchored(600, 600, fmt.Sprintf("ECONOMIC DAY %d", d.EconDay), FontPrimary, 16, color.RGBA{150, 150, 150, 255}, 0.5, 0.5)

	return nil
}