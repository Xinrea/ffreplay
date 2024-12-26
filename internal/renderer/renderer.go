package renderer

import (
	"bytes"
	_ "embed"
	"image/color"
	"log"

	"github.com/Xinrea/ffreplay/internal/layer"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/text/language"
)

//go:embed OPPOSans-Regular.ttf
var fontTTF []byte
var fontSource *text.GoTextFaceSource

type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignRight
	AlignCenter
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fontTTF))
	if err != nil {
		log.Fatal(err)
	}
	fontSource = s
}

type Renderer struct {
	EnemyHealthBar *ProgressBar
	EnemyCasting   *ProgressBar
	PartyList      *PartyList
	PlayProgress   *ProgressBar
}

func NewRenderer() *Renderer {
	pcolor := color.NRGBA{255, 181, 176, 255}
	healthBar := NewProgressBar(400, 6, pcolor)
	return &Renderer{
		EnemyHealthBar: healthBar,
		EnemyCasting:   NewProgressBar(100, 4, pcolor),
		PartyList:      NewPartyList(40, 70),
		PlayProgress:   NewProgressBar(300, 6, color.NRGBA{230, 255, 255, 255}),
	}
}

func DrawText(dst *ebiten.Image, str string, fontSize float64, x, y float64, clr color.Color, align TextAlign) {
	s := ebiten.Monitor().DeviceScaleFactor()
	DrawTextScale(dst, str, fontSize, x, y, clr, align, s)
}

func DrawTextScale(dst *ebiten.Image, str string, fontSize float64, x, y float64, clr color.Color, align TextAlign, s float64) {
	x *= s
	y *= s
	f := &text.GoTextFace{
		Source:    fontSource,
		Direction: text.DirectionLeftToRight,
		Size:      fontSize * s,
		Language:  language.SimplifiedChinese,
	}
	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(clr)
	switch align {
	case AlignLeft:
		op.GeoM.Translate(x, y)
	case AlignRight:
		w, _ := text.Measure(str, f, 0)
		op.GeoM.Translate(x-w, y)
	case AlignCenter:
		w, _ := text.Measure(str, f, 0)
		op.GeoM.Translate(x-w/2, y)
	}
	text.Draw(dst, str, f, op)
}
func DrawFilledRect(canvas *ebiten.Image, x float64, y float64, w float64, h float64, color color.Color) {
	s := ebiten.Monitor().DeviceScaleFactor()
	vector.DrawFilledRect(canvas, float32(x*s), float32(y*s), float32(w*s), float32(h*s), color, true)
}

func StrokeRect(canvas *ebiten.Image, x float64, y float64, w float64, h float64, width float64, color color.Color) {
	s := ebiten.Monitor().DeviceScaleFactor()
	vector.StrokeRect(canvas, float32(x*s), float32(y*s), float32(w*s), float32(h*s), float32(width*s), color, true)
}

func RenderBuffList(canvas *ebiten.Image, tick int64, buffs []model.Buff, x, y, s float64) {
	// render buff icons
	for i, buff := range buffs {
		iconTexture := buff.Texture()
		geoM := iconTexture.GetGeoM()
		geoM.Translate(x+float64(i*25), y)
		geoM.Scale(s, s)
		canvas.DrawImage(iconTexture.Img(), &ebiten.DrawImageOptions{GeoM: geoM})
		remain := buff.Remain(tick)
		if remain > 0 {
			DrawText(canvas, formatSeconds(remain), 14, x+float64(i*25), y+5, color.White, AlignCenter)
		}
	}
}

func (r *Renderer) Init(ecs *ecs.ECS) {
	ecs.AddRenderer(layer.Background, r.BackgroundRender)
	ecs.AddRenderer(layer.SkillRange, r.RangeRender)
	ecs.AddRenderer(layer.Background, r.WorldMarkerRender)
	ecs.AddRenderer(layer.Player, r.EnemyRender)
	ecs.AddRenderer(layer.Player, r.PlayerRender)
	ecs.AddRenderer(layer.UI, r.UIRender)
}
