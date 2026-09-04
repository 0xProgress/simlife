package layouts

import (
	"fmt"
	"image/color"
)

type ProfileData struct {
	Username   string
	AvatarURL  string
	DaysActive int
	RankPercent int
	TopSkills  []Skill
	Badges     []string
}

type Skill struct {
	Name  string
	Level int
}

func RenderProfile(c Canvas, data any) error {
	d, ok := data.(ProfileData)
	if !ok {
		return fmt.Errorf("invalid data type for profile layout")
	}

	c.DrawAvatar(60, 60, 160, d.AvatarURL)
	c.DrawText(250, 120, d.Username, FontPrimary, 48, color.White)
	c.DrawText(250, 170, fmt.Sprintf("Active for %d days", d.DaysActive), FontPrimary, 20, color.RGBA{180, 180, 180, 255})

	c.DrawText(60, 300, "WEALTH RANK", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	c.DrawText(60, 340, fmt.Sprintf("Top %d%%", d.RankPercent), FontMono, 36, color.RGBA{255, 215, 0, 255})

	c.DrawText(400, 300, "TOP SKILLS", FontPrimary, 18, color.RGBA{180, 180, 180, 255})
	y := 340.0
	for _, skill := range d.TopSkills {
		c.DrawText(400, y, skill.Name, FontPrimary, 24, color.White)
		c.DrawProgressBar(600, y-20, 400, 16, float64(skill.Level)/100.0, color.RGBA{0, 200, 255, 255})
		y += 40
	}

	return nil
}