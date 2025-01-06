package ui

import (
	"image"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/entry"
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
	Width         int
	focused       bool
	runes         []rune
	content       string
	counter       int64
	historyMode   bool
	historyIndex  int
	history       []string
	CommitHandler func(string)

	handler furex.ViewHandler
}

func (i *InputHandler) Handler() furex.ViewHandler {
	i.handler.Extra = i
	i.handler.Update = i.Update
	i.handler.JustPressedMouseButtonLeft = i.HandleJustPressedMouseButtonLeft
	i.handler.JustReleasedMouseButtonLeft = i.HandleJustReleasedMouseButtonLeft
	i.handler.MouseEnter = i.HandleMouseEnter
	i.handler.MouseLeave = i.HandleMouseLeave
	return i.handler
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
	entry.GetGlobal(ecsInstance).UIFocus = true
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
	currentWidth := v.MustGetByID("content").Attrs.Width
	if currentWidth+18 <= i.Width {
		i.runes = ebiten.AppendInputChars(i.runes[:0])
		if len(i.runes) > 0 {
			i.historyMode = false
		}
		i.content += string(i.runes)
	} else {
		runes := []rune(i.content)
		i.content = string(runes[:len(runes)-1])
	}
	// If the enter key is pressed, commit this
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyNumpadEnter) {
		i.historyMode = false
		if i.CommitHandler != nil {
			i.CommitHandler(i.content)
		}
		i.history = append(i.history, i.content)
		i.content = ""
	}

	if repeatingKeyPressed(ebiten.KeyArrowUp) {
		if !i.historyMode {
			i.historyMode = true
			i.historyIndex = len(i.history)
		}
		i.historyIndex -= 1
		if i.historyIndex >= 0 {
			i.content = i.history[i.historyIndex]
		} else {
			i.historyIndex++
		}
	}
	if repeatingKeyPressed(ebiten.KeyArrowDown) {
		if i.historyMode {
			i.historyIndex += 1
			if i.historyIndex < len(i.history) {
				i.content = i.history[i.historyIndex]
			} else {
				i.historyIndex--
			}
		}
	}

	// If the backspace key is pressed, remove one character.
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		i.historyMode = false
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

func InputView(prefix string, width int, commitHandler func(string)) *furex.View {
	handler := &InputHandler{
		Width:         width,
		CommitHandler: commitHandler,
	}
	view := furex.NewView(furex.TagName("input"), furex.Direction(furex.Column), furex.Handler(handler))
	view.AddChild(furex.NewView(furex.Height(28), furex.Width(width), furex.Handler(&Sprite{
		NineSliceTexture: messageTextureAtlas.GetNineSlice("input_bg.png"),
	})))
	view.AddChild(furex.NewView(furex.ID("content"), furex.Position(furex.PositionAbsolute), furex.Top(8), furex.Left(6), furex.Height(12), furex.Handler(&Text{
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
	})))
	return view
}
