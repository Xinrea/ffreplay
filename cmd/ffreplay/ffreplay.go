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
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/scene"
	"github.com/Xinrea/ffreplay/internal/scene/scenes"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	bounds       image.Rectangle
	sceneManager *scene.SceneManager
}

func NewGame(sceneName string, opt *scenes.FFLogsOpt) *Game {
	sceneManager := scene.NewSceneManager()
	sceneManager.AddScene(sceneName, opt)
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
	g.bounds = image.Rect(0, 0, width, height)
	g.sceneManager.Layout(width, height)

	s := ebiten.Monitor().DeviceScaleFactor()
	width, height = int(float64(width)*s), int(float64(height)*s)

	return width, height
}

var credential string

func main() {
	defer func() {
		if r := recover(); r != nil {
			util.SetExitMessage(fmt.Sprintf("error: %s\n%s", r, string(debug.Stack())))
			os.Exit(1)
		}
	}()
	// initialize work
	ui.InitializeFont()
	model.Init()
	// ffreplay -r report_code -f fight_id
	// ffpreplay -u https://www.fflogs.com/reports/wrLFVz2QtvnGT9j1?fight=9
	var scene string

	var report string

	var fight int

	var code string

	reportUrl := flag.String("u", "", "FFLogs fight url")
	flag.StringVar(&scene, "s", "replay", "Scene to start")
	flag.StringVar(&report, "r", "", "FFLogs report code")
	flag.IntVar(&fight, "f", 0, "FFlogs report fight code. Report may contains multiple fights")
	flag.StringVar(&code, "c", "", "FFLogs OAuth code")
	flag.Parse()

	log.Println(os.Args)

	if *reportUrl != "" {
		report, fight = parseFightURL(reportUrl)
	}

	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if scene == "playground" {
		startPlayground()

		return
	}

	if scene == "replay" {
		report, fight = parseFightURL(reportUrl)

		if credential == "" {
			credential = loadCredentialFromFile()
		}

		credentials := strings.Split(credential, ":")

		if len(credentials) != 2 || report == "" {
			log.Fatal("Invalid credential:", credential)
		}

		startReplay(code, credentials[0], credentials[1], report, fight)

		return
	}

	startCustomScene(scene)
}

func parseFightURL(reportUrl *string) (string, int) {
	if reportUrl == nil || *reportUrl == "" {
		return "", 0
	}

	parsedUrl, err := url.Parse(*reportUrl)
	if err != nil {
		log.Panic("Invalid report url:", err)
	}

	if !strings.HasPrefix(parsedUrl.Path, "/reports/") {
		log.Panic("Invalid report url:", *reportUrl)
	}

	report := parsedUrl.Path[len("/reports/"):]

	fight, err := strconv.Atoi(parsedUrl.Query().Get("fight"))
	if err != nil {
		if parsedUrl.Query().Get("fight") == "last" {
			fight = -1
		} else {
			log.Panic("Invalid fight id")
		}
	}

	log.Println("report:", report, "fight:", fight)

	return report, fight
}

func startPlayground() {
	ebiten.SetWindowTitle("FFReplay Playground")

	if err := ebiten.RunGame(NewGame("playground", nil)); err != nil {
		log.Panic(err)
	}
}

func startCustomScene(scene string) {
	ebiten.SetWindowTitle("FFReplay " + scene)

	if err := ebiten.RunGame(NewGame(scene, nil)); err != nil {
		log.Panic(err)
	}
}

func startReplay(code string, clientID string, clientSecret string, report string, fight int) {
	ebiten.SetWindowTitle(fmt.Sprintf("FFReplay %s-%d", report, fight))

	if err := ebiten.RunGame(
		NewGame(
			"replay",
			&scenes.FFLogsOpt{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Report:       report,
				Fight:        fight,
				AuthCode:     code,
			},
		),
	); err != nil {
		log.Panic(err)
	}
}

func loadCredentialFromFile() string {
	f, err := os.Open(".credential")
	if err != nil {
		log.Panic("Failed to open credential file:", err)
	}
	defer f.Close()

	cbytes, err := io.ReadAll(f)
	if err != nil {
		log.Panic("Failed to read credential file:", err)
	}

	return string(cbytes)
}
