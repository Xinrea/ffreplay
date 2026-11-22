//go:build js

package ui

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/Xinrea/ffreplay/internal/planner"
)

const boardKey = "ffreplay.planner.draft.v1"
const boardPrefix = "#ffplan="

func readGameCode() (_ string, err error) {
	defer catchBoardError(&err)
	v := js.Global().Call("prompt", "粘贴游戏战术板分享码（导入可撤销）", "")
	if v.IsNull() {
		return "", nil
	}
	return v.String(), nil
}
func writeGameCode(code string) (err error) {
	defer catchBoardError(&err)
	js.Global().Call("prompt", "复制当前步骤的游戏码。头顶 Buff 和自定义底图不包含在内，连线会变为普通线段。", code)
	return nil
}

var leaveBoard js.Func

// Register only while dirty. A clean board should not prompt on navigation
// or prevent browsers from keeping the page in their back/forward cache.
func guardBoardExit(dirty bool) {
	if dirty {
		leaveBoard = js.FuncOf(func(_ js.Value, args []js.Value) any {
			args[0].Call("preventDefault")
			args[0].Set("returnValue", "")
			return nil
		})
		js.Global().Call("addEventListener", "beforeunload", leaveBoard)
	} else {
		js.Global().Call("removeEventListener", "beforeunload", leaveBoard)
		leaveBoard.Release()
	}
}

func openedSharedBoard() bool {
	return strings.HasPrefix(js.Global().Get("location").Get("hash").String(), boardPrefix)
}

// Storage, history and downloads can be unavailable in restricted browsers.
func catchBoardError(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("浏览器操作失败：%v", r)
	}
}

func loadBoard() (_ []byte, err error) {
	defer catchBoardError(&err)
	hash := js.Global().Get("location").Get("hash").String()
	if encoded, ok := strings.CutPrefix(hash, boardPrefix); ok {
		if len(encoded) > base64.RawURLEncoding.EncodedLen(planner.MaxFileSize) {
			return nil, errors.New("分享链接过大")
		}
		return base64.RawURLEncoding.DecodeString(encoded)
	}
	value := js.Global().Get("localStorage").Call("getItem", boardKey)
	if value.IsNull() {
		return nil, nil
	}
	return []byte(value.String()), nil
}

func saveBoard(data []byte) (_ string, err error) {
	defer catchBoardError(&err)
	g := js.Global()
	g.Get("localStorage").Call("setItem", boardKey, string(data))
	if strings.HasPrefix(g.Get("location").Get("hash").String(), boardPrefix) {
		g.Get("history").Call("replaceState", nil, "", g.Get("location").Get("pathname"))
	}
	return "已保存至此浏览器，下次打开会恢复", nil
}

func exportBoard(data []byte) (_ string, err error) {
	defer catchBoardError(&err)
	g := js.Global()
	a := g.Get("document").Call("createElement", "a")
	a.Set("download", "strategy.ffplan")
	a.Set("href", "data:application/json;base64,"+base64.StdEncoding.EncodeToString(data))
	a.Call("click")
	return "已请求下载 strategy.ffplan", nil
}

func shareBoard(data []byte) (_ string, err error) {
	defer catchBoardError(&err)
	if len(data) > 24*1024 {
		return "", errors.New("方案较大，请用「导出文件」分享")
	}
	g := js.Global()
	link := g.Get("location").Get("pathname").String() + boardPrefix + base64.RawURLEncoding.EncodeToString(data)
	g.Get("history").Call("replaceState", nil, "", link)
	return "地址栏已生成分享链接，复制当前网址即可。链接包含整个方案。", nil
}

func (b *StrategyBoard) platformControls(x int) {
	b.button(x, b.fileActionsY(), 80, "分享链接", false, func() { b.store(shareBoard) })
}
