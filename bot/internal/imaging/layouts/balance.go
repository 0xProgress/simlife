package layouts

import (
	"fmt"
	"image/color"
)

type BalanceData struct {
	Username    string
	AvatarURL   string
	Wallet      int64
	Bank        int64
	Escrow      int64
	NetWorth    int64
	Change24h   float64
	RankPercent int
	EconDay     int
}

func RenderBalance(c Canvas, data any) error {
	d, ok := data.(BalanceData)
	if !ok {
		return fmt.Errorf("invalid data type for balance layout")
	}

	c.DrawAvatar(40, 40, 80, d.AvatarURL)
	c.DrawText(140, 90, d.Username, FontPrimary, 32, color.White)

	c.DrawText(60, 220, "WALLET", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 280, d.Wallet, color.RGBA{255, 215, 0, 255}) // Gold

	c.DrawText(60, 360, "BANK", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 400, d.Bank, color.White)

	c.DrawText(60, 500, "NET WORTH", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawCurrencyValue(60, 550, d.NetWorth, color.White)

	changeColor := color.RGBA{0, 255, 0, 255} // Green
	if d.Change24h < 0 {
		changeColor = color.RGBA{255, 69, 0, 255} // Red
	}
	changeStr := fmt.Sprintf("%.2f%%", d.Change24h)
	if d.Change24h > 0 {
		changeStr = "+" + changeStr
	}
	c.DrawTextAnchored(1140, 550, changeStr, FontMono, 36, changeColor, 1.0, 0.5)

	c.DrawTextAnchored(600, 600, fmt.Sprintf("ECONOMIC DAY %d", d.EconDay), FontPrimary, 16, color.RGBA{150, 150, 150, 255}, 0.5, 0.5)

	return nil
}
