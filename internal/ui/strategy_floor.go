package ui

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/ebitenui/ebitenui/widget"
)

func (b *StrategyBoard) loadFloors() {
	b.floors = map[string]string{"": ""}
	catalog, err := model.ReadMapPreset()
	if err != nil {
		b.status = "场地素材目录未能加载，仍可使用空白场地"
		return
	}
	add := func(p string) {
		key := strings.TrimSuffix(path.Base(p), path.Ext(p))
		b.floors[key] = p
	}
	for _, floor := range catalog.Maps {
		add(floor.Path)
		for _, phase := range floor.Phases {
			add(phase.Path)
		}
	}
}

func (b *StrategyBoard) decode(data []byte) (planner.Plan, error) {
	p, err := planner.Decode(data)
	if err != nil {
		return p, err
	}
	for _, s := range p.Steps {
		if _, ok := b.floors[s.Floor]; !ok {
			return planner.Plan{}, fmt.Errorf("未知场地背景 %q", s.Floor)
		}
	}
	return p, nil
}

func (b *StrategyBoard) floorPicker() {
	keys := slices.Sorted(maps.Keys(b.floors))
	entries := make([]any, len(keys))
	for i, k := range keys {
		entries[i] = k
	}
	label := func(entry any) string {
		if k := entry.(string); k != "" {
			for _, p := range []struct{ prefix, label string }{{"dragonsongs_reprise_", "绝龙诗"}, {"futures_rewritten_", "绝伊甸"}, {"omega_", "绝欧米茄"}, {"tea-p", "绝亚历山大"}} {
				if n, ok := strings.CutPrefix(k, p.prefix); ok {
					return p.label + "  P" + strings.ReplaceAll(n, "_", "/")
				}
			}
			return strings.ToUpper(strings.ReplaceAll(k, "_", " · "))
		}
		return "无场地背景"
	}
	s := b.editor.Step()
	st := newPropInspectorStyle(1)
	button, list := hudListParams(st)
	b.ui.Container.AddChild(widget.NewListComboButton(
		widget.ListComboButtonOpts.WidgetOpts(boardAt(16, 184, 170, 28)...),
		widget.ListComboButtonOpts.ButtonParams(button),
		widget.ListComboButtonOpts.ListParams(list),
		widget.ListComboButtonOpts.Entries(entries), widget.ListComboButtonOpts.InitialEntry(s.Floor),
		widget.ListComboButtonOpts.MaxContentHeight(260),
		widget.ListComboButtonOpts.EntryLabelFunc(label, label),
		widget.ListComboButtonOpts.EntrySelectedHandler(func(a *widget.ListComboButtonEntrySelectedEventArgs) {
			k := a.Entry.(string)
			if k != s.Floor {
				b.editor.Checkpoint()
				s.Floor = k
				b.refreshOnIdle = true
			}
		}),
	))
}
