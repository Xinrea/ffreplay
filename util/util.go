package util

import (
	"image"
	"math"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// RangeLimit v into [-m, m], m should be positive.
func RangeLimit(v, m float64) float64 {
	if v < -m {
		return -m
	}

	if v > m {
		return m
	}

	return v
}

func IsWasm() bool {
	return runtime.GOOS == "js" && runtime.GOARCH == "wasm"
}

func WindowSize() (int, int) {
	return ebiten.WindowSize()
}

func Lerpf(a, b, t float64) float64 {
	return a + (b-a)*t
}

func NormalizeAngle(angle float64) float64 {
	normalized := math.Mod(angle, 360)
	if normalized < 0 {
		normalized += 360
	}

	return normalized
}

func NormalizeRadians(radian float64) float64 {
	for radian < 0 {
		radian += 2 * math.Pi
	}

	for radian > 2*math.Pi {
		radian -= 2 * math.Pi
	}

	return radian
}

func LerpRadians(a, b, t float64) float64 {
	a = NormalizeRadians(a)
	b = NormalizeRadians(b)
	// make diff into [-pi, pi]
	diff := math.Mod(((b-a)+math.Pi), (2*math.Pi)) - math.Pi

	return Lerpf(a, a+diff, t)
}

func TickToMS(ticks int64) int64 {
	return int64(float64(ticks) / ebiten.DefaultTPS * 1000)
}

func MSToTick(ms int64) int64 {
	return int64(float64(ms) / 1000 * ebiten.DefaultTPS)
}

func ScaleFrame(frame image.Rectangle) image.Rectangle {
	s := ebiten.Monitor().DeviceScaleFactor()

	return image.Rect(
		int(float64(frame.Min.X)*s),
		int(float64(frame.Min.Y)*s),
		int(float64(frame.Max.X)*s),
		int(float64(frame.Max.Y)*s),
	)
}
