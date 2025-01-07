package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
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
		c.AddResult("/map list - 查看地图列表")
		c.AddResult("/map phase <id> - 查看地图分段")
		c.AddResult("/map set <id> [phase] - 设置地图")
		c.AddResult("/player add <name> <role> - 添加玩家")
		c.AddResult("/player remove <id> - 移除玩家")
		c.AddResult("/clear - 清空记录")
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

func (c *CommandHandler) mapHandler(cmds []string) {
	if len(cmds) == 0 {
		c.AddError("Invalid map command")

		return
	}

	switch cmds[0] {
	case "list":
		mapids := []string{}
		for k, m := range model.MapCache {
			mapids = append(mapids, fmt.Sprintf("%d-%s", k, m.Path))
		}

		sort.Strings(mapids)

		for _, m := range mapids {
			c.AddResult(m)
		}
	case "phase":
		id, _ := strconv.Atoi(cmds[1])
		if m, ok := model.MapCache[id]; ok {
			for i := range m.Phases {
				c.AddResult(fmt.Sprintf("Phase %d", i))
			}
		} else {
			c.AddError("Invalid map id")
		}
	case "set":
		mapData := component.Map.Get(component.Map.MustFirst(ecsInstance.World))
		cameraData := entry.GetCamera(ecsInstance)
		id, _ := strconv.Atoi(cmds[1])

		if m, ok := model.MapCache[id]; ok {
			mapData.Config = m.Load()
			current := mapData.Config.Maps[mapData.Config.CurrentMap]
			cameraData.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)

			c.AddResult("Setup map " + cmds[1])

			if len(cmds) > 2 {
				phase, _ := strconv.Atoi(cmds[2])
				if phase < len(mapData.Config.Phases) && phase >= 0 {
					mapData.Config.CurrentPhase = phase

					c.AddResult("Setup phase " + cmds[2])
				}
			}
		} else {
			c.AddError("Invalid map id")
		}
	default:
		c.AddError("Invalid map command")
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
