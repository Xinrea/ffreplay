package ui

import (
	"image/color"
	"strings"

	"github.com/Xinrea/ffreplay/internal/entry"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	ResultColor = color.NRGBA{24, 169, 248, 128}
	PromptColor = color.NRGBA{255, 255, 255, 128}
	ErrorColor  = color.NRGBA{255, 0, 0, 128}
)

const InputHeight = 28

type CommandHandler struct {
	euiWrap      *widget.Container
	euiMessage   *widget.Container
	euiInput     *widget.TextInput
	euiScale     float64
	euiHistory   []string
	historyMode  bool
	historyIndex int
	player       PlayerCommand
}

type PlayerCommand struct {
	idcnt int64
}

func (c *CommandHandler) CommitCommand(cmd string) {
	if cmd == "" {
		return
	}

	c.Execute(cmd)
}

func (c *CommandHandler) Execute(cmd string) {
	c.AddEcho(cmd)

	commands := strings.Split(cmd, " ")
	switch commands[0] {
	case "/help":
		c.AddResult("可用命令：")
		c.AddResult("- /map list")
		c.AddResult("查看地图列表")
		c.AddResult("- /map phase <id>")
		c.AddResult("查看地图分段，例如 /map phase 77")
		c.AddResult("- /map set <id> [phase]")
		c.AddResult("设置地图，例如 /map set 77 4")
		c.AddResult("- /player add <name> <role>")
		c.AddResult("添加玩家，例如 /player add Xinrea Scholar")
		c.AddResult("- /player remove <id>")
		c.AddResult("移除玩家，例如 /player remove 1")
		c.AddResult("- /clear")
		c.AddResult("清空消息")
		c.AddResult("- /script run <filename>")
		c.AddResult("运行 Lua 脚本文件")
		c.AddResult("- /script exec <code>")
		c.AddResult("执行内联 Lua 代码")
	case "/map":
		c.mapHandler(commands[1:])
	case "/player":
		c.playerHandler(commands[1:])
	case "/script":
		c.scriptHandler(commands[1:])
	case "/clear":
		c.clearMessages()
	default:
		c.AddError("Invalid command: " + commands[0])
	}
}

func (c *CommandHandler) AddEcho(cmd string) {
	c.addEUIText("> "+cmd, color.White)
}

func (c *CommandHandler) AddResult(result string) {
	c.addEUIText(result, color.White)
}

func (c *CommandHandler) AddPrompt(prompt string) {
	c.addEUIText(prompt, PromptColor)
}

func (c *CommandHandler) AddError(err string) {
	c.addEUIText(err, color.NRGBA{255, 120, 120, 255})
}

func (c *CommandHandler) clearMessages() {
	c.euiMessage.RemoveChildren()
	c.AddPrompt("输入 /help 查看可用命令")
}

func (c *CommandHandler) addEUIText(content string, clr color.Color) {
	scale := c.euiScale
	if scale <= 0 {
		scale = 1
	}
	face := newEUIFace(12 * scale)
	text := widget.NewText(
		widget.TextOpts.Text(content, &face, clr),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch:  true,
				MaxWidth: int(380 * scale),
			}),
		),
	)
	c.euiMessage.AddChild(text)
}

func (c *CommandHandler) handleEUIInputUpdate() {
	if c.euiInput == nil || !c.euiInput.IsFocused() {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if !c.historyMode {
			c.historyMode = true
			c.historyIndex = len(c.euiHistory)
		}
		c.historyIndex--
		if c.historyIndex < 0 {
			c.historyIndex = 0
		}
		if c.historyIndex < len(c.euiHistory) {
			c.euiInput.SetText(c.euiHistory[c.historyIndex])
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && c.historyMode {
		c.historyIndex++
		if c.historyIndex >= len(c.euiHistory) {
			c.historyIndex = len(c.euiHistory)
			c.euiInput.SetText("")
			return
		}
		c.euiInput.SetText(c.euiHistory[c.historyIndex])
	}
}

func NewEUICommandView(scale float64) *widget.Container {
	handler := &CommandHandler{}
	width := int(400 * scale)
	pad := int(float64(UIPadding) * scale)
	messagePadX := int(10 * scale)
	messagePadY := int(5 * scale)
	fontSize := 12 * scale
	face := newEUIFace(fontSize)

	wrap := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left:   pad,
				Bottom: pad,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)

	message := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(nineSliceWithAlpha(messageTextureAtlas.GetNineSlice("message_bg.png"), 0.5)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    messagePadY,
				Bottom: messagePadY,
				Left:   messagePadX,
				Right:  messagePadX,
			}),
			widget.RowLayoutOpts.Spacing(int(5*scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(width, int(34*scale)),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	wrap.AddChild(message)

	inputBg := toEUINineSlice(messageTextureAtlas.GetNineSlice("input_bg.png"))
	inputWrap := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(inputBg),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(width, int(InputHeight*scale)),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	promptFace := newEUIFace(fontSize)
	inputWrap.AddChild(widget.NewText(
		widget.TextOpts.Text("> ", &promptFace, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding: &widget.Insets{
				Left: int(6 * scale),
			},
		})),
	))
	textInput := widget.NewTextInput(
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:      transparentNineSlice(),
			Disabled:  transparentNineSlice(),
			Highlight: euiimage.NewNineSliceColor(color.NRGBA{24, 169, 248, 96}),
		}),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.White,
			Disabled:      color.NRGBA{255, 255, 255, 120},
			Caret:         color.White,
			DisabledCaret: color.NRGBA{255, 255, 255, 120},
		}),
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Padding(&widget.Insets{
			Left:   int(18 * scale),
			Right:  int(6 * scale),
			Top:    int(8 * scale),
			Bottom: int(8 * scale),
		}),
		widget.TextInputOpts.SubmitOnEnter(true),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			cmd := strings.TrimSpace(args.InputText)
			if cmd == "" {
				args.TextInput.SetText("")
				return
			}
			handler.historyMode = false
			handler.euiHistory = append(handler.euiHistory, cmd)
			handler.CommitCommand(cmd)
			args.TextInput.SetText("")
		}),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(width, int(InputHeight*scale)),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
			widget.WidgetOpts.OnUpdate(func(w widget.HasWidget) {
				handler.handleEUIInputUpdate()
				if handler.euiInput != nil && handler.euiInput.IsFocused() {
					entry.GetGlobal(ecsInstance).UIFocus = true
				}
			}),
		),
	)
	inputWrap.AddChild(textInput)
	wrap.AddChild(inputWrap)

	handler.euiWrap = wrap
	handler.euiMessage = message
	handler.euiInput = textInput
	handler.euiScale = scale
	handler.AddPrompt("输入 /help 查看可用命令")

	return wrap
}
