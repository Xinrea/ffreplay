package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/scene"
	"github.com/Xinrea/ffreplay/internal/scene/scenes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	bounds       image.Rectangle
	sceneManager *scene.SceneManager
}

func NewGame(clientID, clientSecret, report string, fight int) *Game {
	sceneManager := scene.NewSceneManager()
	sceneManager.AddScene(report, scenes.NewFFScene(fflogs.NewFFLogsClient(clientID, clientSecret), report, fight))
	g := &Game{
		bounds:       image.Rectangle{},
		sceneManager: sceneManager,
	}
	return g
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.sceneManager.ResetScene()
	}
	g.sceneManager.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	g.sceneManager.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	w := int(float64(width) * s)
	h := int(float64(height) * s)
	g.bounds = image.Rect(0, 0, w, h)
	g.sceneManager.Layout(width, height)
	return w, h
}

var credential string

func main() {
	// ffreplay -r report_code -f fight_id
	// ffpreplay -u https://www.fflogs.com/reports/wrLFVz2QtvnGT9j1?fight=9
	var report string
	var fight int
	var timelinePath string
	reportUrl := flag.String("u", "", "FFLogs fight url")
	flag.StringVar(&report, "r", "", "FFLogs report code")
	flag.IntVar(&fight, "f", 0, "FFlogs report fight code. Report may contains multiple fights")
	flag.StringVar(&timelinePath, "t", "", "Cutom scene with timeline")
	flag.Parse()
	log.Println(os.Args)
	if reportUrl != nil && *reportUrl != "" {
		parsedUrl, err := url.Parse(*reportUrl)
		if err != nil {
			log.Fatal("Invalid report url:", err)
		}
		report = parsedUrl.Path[len("/reports/"):]
		fight, err = strconv.Atoi(parsedUrl.Query().Get("fight"))
		if err != nil {
			if parsedUrl.Query().Get("fight") == "last" {
				fight = -1
			} else {
				log.Fatal("Invalid fight id")
			}
		}
	}
	if report == "" {
		log.Println("Invalid report")
		return
	}
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle(fmt.Sprintf("FFReplay %s-%d", report, fight))
	if credential == "" {
		f, err := os.Open(".credential")
		if err != nil {
			log.Fatal("Failed to open credential file:", err)
		}
		defer f.Close()
		cbytes, err := io.ReadAll(f)
		if err != nil {
			log.Fatal("Failed to read credential file:", err)
		}
		credential = string(cbytes)
	}
	credentials := strings.Split(credential, ":")
	if err := ebiten.RunGame(NewGame(credentials[0], credentials[1], report, fight)); err != nil {
		log.Fatal(err)
	}
}
