package ui

import (
	"fmt"
	"strings"

	"github.com/Xinrea/ffreplay/internal/gameboard"
	"github.com/Xinrea/ffreplay/internal/planner"
)

func (b *StrategyBoard) gameCodeButtons(x, y int) {
	b.button(x, y, 83, "导入游戏码", false, func() {
		code, err := readGameCode()
		if err != nil {
			b.status = err.Error()

			return
		}

		if code != "" {
			b.importGame(code)
		}
	})
	b.button(x+88, y, 85, "导出本步码", false, func() {
		code, err := planner.ExportGame(b.editor.Plan, b.editor.Current)
		if err == nil {
			err = writeGameCode(code)
		}

		if err != nil {
			b.status = "未导出：" + err.Error()

			return
		}

		b.status = "已生成当前步骤的游戏码；Buff、底图、自定义名称不随码传递，连线转为普通线段"
	})
}

func (b *StrategyBoard) importGame(text string) {
	if i := strings.Index(text, gameboard.Prefix); i >= 0 {
		text = text[i:]
		if j := strings.IndexByte(text, ']'); j >= 0 {
			text = text[:j+1]
		}
	}

	p, note, err := planner.ImportGame(text)
	if err != nil {
		b.status = "未导入：" + err.Error()

		return
	}

	b.editor.Checkpoint()
	b.editor.Plan, b.editor.Current = p, 0
	b.resetSelection()
	b.panel = ""
	b.status = fmt.Sprintf("已导入 %s，可撤销%s", p.Steps[0].Name, note)
}
