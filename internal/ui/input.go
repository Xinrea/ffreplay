package ui

import (
	"image"
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/furex/v2"
)

var messageTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/message.xml")

type Focusable interface {
	SetFocus(bool)
}

type InputHandler struct {
	focused       bool
	runes         []rune
	content       string
	counter       int64
	CommitHandler func(string)
}

// HandleMouseEnter implements furex.MouseEnterLeaveHandler.
func (i *InputHandler) HandleMouseEnter(x int, y int) bool {
	ebiten.SetCursorShape(ebiten.CursorShapeText)
	return true
}

// HandleMouseLeave implements furex.MouseEnterLeaveHandler.
func (i *InputHandler) HandleMouseLeave() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}

// SetFocus implements Focusable.
func (i *InputHandler) SetFocus(b bool) {
	i.focused = b
}

// HandleJustPressedMouseButtonLeft implements furex.MouseLeftButtonHandler.
func (i *InputHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x int, y int) bool {
	i.focused = true
	return true
}

// HandleJustReleasedMouseButtonLeft implements furex.MouseLeftButtonHandler.
func (i *InputHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x int, y int) {
}

// repeatingKeyPressed return true when key is pressed considering the repeat state.
func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 20
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (i *InputHandler) Update(v *furex.View) {
	if !i.focused {
		return
	}
	i.runes = ebiten.AppendInputChars(i.runes[:0])
	i.content += string(i.runes)
	// If the enter key is pressed, commit this
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyNumpadEnter) {
		if i.CommitHandler != nil {
			i.CommitHandler(i.content)
		}
		i.content = ""
	}

	// If the backspace key is pressed, remove one character.
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(i.content) >= 1 {
			runes := []rune(i.content)
			i.content = string(runes[:len(runes)-1])
		}
	}

	i.counter += 1
}

func (i *InputHandler) Content() string {
	return i.content
}

var _ Focusable = (*InputHandler)(nil)
var _ furex.MouseLeftButtonHandler = (*InputHandler)(nil)
var _ furex.MouseEnterLeaveHandler = (*InputHandler)(nil)

func InputView(prefix string) *furex.View {
	handler := &InputHandler{}
	view := &furex.View{
		Direction: furex.Column,
		Handler:   handler,
	}
	view.AddChild(&furex.View{
		Height: 28,
		Handler: &Sprite{
			NineSliceTexture: messageTextureAtlas.GetNineSlice("input_bg.png"),
		},
	})
	view.AddChild(&furex.View{
		Position: furex.PositionAbsolute,
		Top:      8,
		Left:     6,
		Height:   12,
		Handler: &Text{
			Align: furex.AlignItemStart,
			Content: func() string {
				if handler.focused && handler.counter%60 > 30 {
					return prefix + handler.Content() + "|"
				}
				return prefix + handler.Content()
			},
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		},
	})
	return view
}
