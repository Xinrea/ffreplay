package ui

import (
	"image/color"
	"strings"

	"github.com/yohamta/furex/v2"
)

var (
	ResultColor = color.NRGBA{24, 169, 248, 128}
	PromptColor = color.NRGBA{255, 255, 255, 128}
	ErrorColor  = color.NRGBA{255, 0, 0, 128}
)

type CommandHandler struct {
	wrap    *furex.View
	message *furex.View
	input   *furex.View
	player  PlayerCommand
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
	case "/map":
		c.mapHandler(commands[1:])
	case "/player":
		c.playerHandler(commands[1:])
	case "/clear":
		c.message.RemoveAll()
		c.message.SetHeight(12)
	default:
		c.AddError("Invalid command: " + commands[0])
	}
}

func (c *CommandHandler) AddEcho(cmd string) {
	// add echo message
	text := &Text{
		Align:   furex.AlignItemStart,
		Content: "> " + cmd,
		Color:   color.White,
	}
	c.message.AddChild(furex.NewView(furex.MarginLeft(10), furex.MarginTop(5), furex.Height(12), furex.Handler(text)))
	c.message.SetHeight(c.message.Attrs.Height + 12 + 5)
}

func (c *CommandHandler) AddResult(result string) {
	// add echo message
	text := &Text{
		Align:        furex.AlignItemStart,
		Content:      result,
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  ResultColor,
	}
	c.message.AddChild(furex.NewView(furex.MarginLeft(10), furex.MarginTop(5), furex.Height(12), furex.Handler(text)))
	c.message.SetHeight(c.message.Attrs.Height + 12 + 5)
}

func (c *CommandHandler) AddPrompt(prompt string) {
	// add echo message
	text := &Text{
		Align:   furex.AlignItemStart,
		Content: prompt,
		Color:   PromptColor,
	}
	c.message.AddChild(furex.NewView(furex.MarginLeft(10), furex.MarginTop(5), furex.Height(12), furex.Handler(text)))
	c.message.SetHeight(c.message.Attrs.Height + 12 + 5)
}

func (c *CommandHandler) AddError(err string) {
	// add echo message
	text := &Text{
		Align:        furex.AlignItemStart,
		Content:      err,
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  ErrorColor,
	}
	c.message.AddChild(furex.NewView(furex.MarginLeft(10), furex.MarginTop(5), furex.Height(12), furex.Handler(text)))
	c.message.SetHeight(c.message.Attrs.Height + 12 + 5)
}

func CommandView() *furex.View {
	handler := &CommandHandler{}
	view := furex.NewView(
		furex.TagName("command"),
		furex.Direction(furex.Column),
		furex.Justify(furex.JustifyEnd),
	)

	message := furex.NewView(
		furex.TagName("message"),
		furex.Direction(furex.Column),
		furex.Width(400),
		furex.Height(34),
		furex.Handler(&Sprite{
			NineSliceTexture: messageTextureAtlas.GetNineSlice("message_bg.png"),
			BlendAlpha:       true,
			Alpha:            0.5,
		}))
	text := &Text{
		Align:   furex.AlignItemStart,
		Content: "输入 /help 查看可用命令",
		Color:   PromptColor,
	}
	message.AddChild(furex.NewView(furex.MarginLeft(10), furex.MarginTop(5), furex.Height(12), furex.Handler(text)))
	view.AddChild(message)

	input := InputView("> ", 400, handler.CommitCommand)
	view.AddChild(input)
	handler.wrap = view
	handler.message = message
	handler.input = input

	return view
}
