package renderer

import (
	"fmt"
	"image/color"
	"io/fs"
	"log"

	asset "github.com/Xinrea/ffreplay"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	TIMELINE_WIDTH  = 800
	TIMELINE_HEIGHT = 130
	NONEGCD_GAP     = 50
)

var (
	shader     *ebiten.Shader
	skillLayer *ebiten.Image
)

type SkillTimeline struct {
	Periods []GCDPeriod
}

type GCDPeriod struct {
	StartTick     int64
	GCDSkill      *model.Skill
	NoneGCDSkills []*model.Skill
}

func initBackground() {
	s := ebiten.Monitor().DeviceScaleFactor()
	skillLayer = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))

	// The browser build fetches this file at startup. A non-200 used to reach
	// NewShader as an empty source and panic, which exits the replay. The fade
	// is decorative; draw the timeline directly when the shader is unavailable.
	shaderSrc, err := fs.ReadFile(asset.AssetFS, "asset/shader/easeinout.kage")
	if err != nil {
		log.Printf("Timeline shader unavailable, drawing without easing: %v", err)

		return
	}

	compiled, err := ebiten.NewShader(shaderSrc)
	if err != nil {
		log.Printf("Timeline shader failed to compile, drawing without easing: %v", err)

		return
	}

	shader = compiled
}

func RenderCasting(debug bool, canvas *ebiten.Image, tick int64, cast *model.Skill, x, y float64) {
	s := ebiten.Monitor().DeviceScaleFactor()

	textSize := 12.0 * s
	yOffset := 30.0 * s

	iconTexture := cast.Texture()
	if iconTexture == nil {
		log.Printf(
			"Ability texture is nil, using fallback texture: id=%d name=%q icon=%q start_tick=%d render_tick=%d",
			cast.ID,
			cast.Name,
			cast.Icon,
			cast.StartTick,
			tick,
		)
		iconTexture = texture.NewAbilityTexture("")
	}
	geoM := texture.CenterGeoM(iconTexture)
	geoM.Scale(s, s)

	borderGeoM := texture.CenterGeoM(model.BorderTexture)
	borderGeoM.Scale(s, s)

	if !cast.IsGCD {
		geoM.Scale(0.8, 0.8)
		borderGeoM.Scale(0.8, 0.8)
		geoM.Translate(0, -30*s)
		borderGeoM.Translate(0, -30*s)

		textSize = 10.0 * s
		yOffset = -55 * s
	}

	geoM.Translate(x, y)
	borderGeoM.Translate(x, y)
	canvas.DrawImage(iconTexture, &ebiten.DrawImageOptions{GeoM: geoM})
	canvas.DrawImage(model.BorderTexture, &ebiten.DrawImageOptions{GeoM: borderGeoM})

	name := cast.Name
	if debug {
		name = fmt.Sprintf("[%d]%s", cast.ID, cast.Name)
	}

	ui.DrawText(canvas, name, textSize, x, y+yOffset, color.White, ui.AlignCenter, textShdowOpt)
}

func (g GCDPeriod) Render(debug bool, canvas *ebiten.Image, x, y float64, tick int64) {
	s := ebiten.Monitor().DeviceScaleFactor()

	if g.GCDSkill != nil {
		// tailLength := tickToLength(util.MSToTick(g.GCDSkill.Cast))
		// DrawFilledRect(canvas, x, y-6, tailLength, 12, color.NRGBA{119, 123, 131, 255})
		RenderCasting(debug, canvas, tick, g.GCDSkill, x, y)
	}

	for i := range g.NoneGCDSkills {
		RenderCasting(debug, canvas, tick, g.NoneGCDSkills[i], x+float64(i+1)*NONEGCD_GAP*s, y)
	}
}

func NewSkillTimeline(casts []*model.Skill) SkillTimeline {
	periods := []GCDPeriod{}
	currentPeriod := GCDPeriod{
		StartTick: -1,
	}

	for _, c := range casts {
		if c.IsGCD {
			if currentPeriod.StartTick != -1 {
				periods = append(periods, currentPeriod)
			}

			currentPeriod = GCDPeriod{
				StartTick: c.StartTick,
				GCDSkill:  c,
			}

			continue
		}

		if currentPeriod.StartTick == -1 {
			currentPeriod.StartTick = c.StartTick
		}

		currentPeriod.NoneGCDSkills = append(currentPeriod.NoneGCDSkills, c)
	}

	if currentPeriod.StartTick != -1 {
		periods = append(periods, currentPeriod)
	}

	return SkillTimeline{
		Periods: periods,
	}
}

func tickToLength(tick int64) float64 {
	return float64(tick) * 1.3
}

func (st SkillTimeline) Render(debug bool, canvas *ebiten.Image, x, y float64, tick int64) {
	s := ebiten.Monitor().DeviceScaleFactor()

	if len(st.Periods) == 0 {
		return
	}

	if int(TIMELINE_HEIGHT*s) != skillLayer.Bounds().Dy() || int(TIMELINE_WIDTH*s) != skillLayer.Bounds().Dx() {
		skillLayer = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))

		log.Println("resize skillLayer")
	}

	x = x * s
	y = y * s

	x = x - TIMELINE_WIDTH*s/2

	skillLayer.Clear()

	for i := range st.Periods {
		offset := tickToLength(tick - st.Periods[i].StartTick)
		px := -offset*s + TIMELINE_WIDTH*s
		py := TIMELINE_HEIGHT*s/2.0 + 15*s
		st.Periods[i].Render(debug, skillLayer, px, py, tick)
	}

	if shader == nil {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(x, y)
		canvas.DrawImage(skillLayer, &op)

		return
	}

	sop := &ebiten.DrawRectShaderOptions{}
	sop.GeoM.Translate(x, y)
	sop.Images[0] = skillLayer
	canvas.DrawRectShader(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s), shader, sop)
}
