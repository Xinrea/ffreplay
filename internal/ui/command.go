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

var ResultColor = color.NRGBA{24, 169, 248, 128}
var PromptColor = color.NRGBA{255, 255, 255, 128}
var ErrorColor = color.NRGBA{255, 0, 0, 128}

type CommandHandler struct {
	wrap    *furex.View
	message *furex.View
	input   *furex.View
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
		c.AddResult("/map set <id> - 设置地图")
		c.AddResult("/clear - 清空记录")
	case "/map":
		c.mapHandler(commands[1:])
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
	case "set":
		mapData := component.Map.Get(component.Map.MustFirst(ecsInstance.World))
		cameraData := entry.GetCamera(ecsInstance)
		id, _ := strconv.Atoi(cmds[1])
		if m, ok := model.MapCache[id]; ok {
			mapData.Config = m.Load()
			current := mapData.Config.Maps[mapData.Config.CurrentMap]
			cameraData.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)
			c.AddResult("Setup map " + cmds[1])
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
	c.message.AddChild(&furex.View{
		MarginLeft: 10,
		MarginTop:  5,
		Height:     12,
		Handler:    text,
	})
	c.message.SetHeight(c.message.Height + 12 + 5)
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
	c.message.AddChild(&furex.View{
		MarginLeft: 10,
		MarginTop:  5,
		Height:     12,
		Handler:    text,
	})
	c.message.SetHeight(c.message.Height + 12 + 5)
}

func (c *CommandHandler) AddPrompt(prompt string) {
	// add echo message
	text := &Text{
		Align:   furex.AlignItemStart,
		Content: prompt,
		Color:   PromptColor,
	}
	c.message.AddChild(&furex.View{
		MarginLeft: 10,
		MarginTop:  5,
		Height:     12,
		Handler:    text,
	})
	c.message.SetHeight(c.message.Height + 12 + 5)
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
	c.message.AddChild(&furex.View{
		MarginLeft: 10,
		MarginTop:  5,
		Height:     12,
		Handler:    text,
	})
	c.message.SetHeight(c.message.Height + 12 + 5)
}

func CommandView() *furex.View {
	view := &furex.View{
		Direction: furex.Column,
		Justify:   furex.JustifyEnd,
	}
	handler := &CommandHandler{}
	view.Handler = handler

	view.Width = 400
	message := &furex.View{
		Direction: furex.Column,
		Width:     400,
		Height:    34,
		Handler: &Sprite{
			NineSliceTexture: messageTextureAtlas.GetNineSlice("message_bg.png"),
			BlendAlpha:       true,
			Alpha:            0.5,
		},
	}
	text := &Text{
		Align:   furex.AlignItemStart,
		Content: "输入 /help 查看可用命令",
		Color:   PromptColor,
	}
	message.AddChild(&furex.View{
		MarginLeft: 10,
		MarginTop:  10,
		Height:     12,
		Handler:    text,
	})
	view.AddChild(message)
	input := InputView("> ", 400, handler.CommitCommand)
	view.AddChild(input)
	handler.wrap = view
	handler.message = message
	handler.input = input
	return view
}
