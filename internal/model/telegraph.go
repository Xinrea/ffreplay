package model

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

// TelegraphType defines the visual shape of a telegraph.
type TelegraphType int

const (
	TelegraphCircle TelegraphType = iota
	TelegraphRect
	TelegraphCone
	TelegraphText
)

// TelegraphData holds data for a temporary visual telegraph (AoE indicator, text annotation, etc.).
// Telegraghs are created by scripts and automatically removed when their duration expires.
// A duration of 0 means permanent (until manually removed).
type TelegraphData struct {
	Type     TelegraphType
	Object   object.Object
	Text     string  // only for TelegraphText
	Color    color.NRGBA
	Duration int64   // remaining in ticks; 0 = permanent
	MaxDuration int64 // for alpha fade-out calculation
}

// NewTelegraphCircle creates a circular telegraph at the given position.
func NewTelegraphCircle(pos vector.Vector, radius float64, fill color.NRGBA, stroke color.NRGBA, durationMs int64) *TelegraphData {
	opt := object.ObjectOption{
		FillColor:   fill,
		StrokeColor: stroke,
		StrokeWidth: 4,
	}
	obj := object.NewCircleObject(opt, pos, radius)
	duration := msToTick(durationMs)

	return &TelegraphData{
		Type:     TelegraphCircle,
		Object:   obj,
		Duration: duration,
		MaxDuration: duration,
		Color:    fill,
	}
}

// NewTelegraphRect creates a rectangular telegraph at the given position.
// anchor: 0=middle, 1=bottom-middle (same as RectObject constants)
func NewTelegraphRect(pos vector.Vector, anchor int, width, height float64, fill color.NRGBA, stroke color.NRGBA, durationMs int64) *TelegraphData {
	opt := object.ObjectOption{
		FillColor:   fill,
		StrokeColor: stroke,
		StrokeWidth: 4,
	}
	obj := object.NewRectObject(opt, pos, anchor, width, height)
	duration := msToTick(durationMs)

	return &TelegraphData{
		Type:     TelegraphRect,
		Object:   obj,
		Duration: duration,
		MaxDuration: duration,
		Color:    fill,
	}
}

// NewTelegraphText creates a text annotation telegraph.
func NewTelegraphText(pos vector.Vector, text string, c color.NRGBA, durationMs int64) *TelegraphData {
	duration := msToTick(durationMs)
	// Text telegraphs use a point object for positioning only
	obj := object.NewPointObject(pos)

	return &TelegraphData{
		Type:     TelegraphText,
		Object:   obj,
		Text:     text,
		Duration: duration,
		MaxDuration: duration,
		Color:    c,
	}
}

// msToTick converts milliseconds to game ticks (60 ticks/second).
func msToTick(ms int64) int64 {
	if ms <= 0 {
		return 0
	}
	return ms * 60 / 1000
}
