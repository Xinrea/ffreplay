package texture

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	asset "github.com/Xinrea/ffreplay"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var textureCache = sync.Map{}

func loadFromFFreplay(iconName string) *ebiten.Image {
	u, err := url.Parse("https://ffreplay.vjoi.cn/img/ff/abilities/" + iconName)
	if err != nil {
		log.Println(err)

		return nil
	}

	finalUrl := u.ResolveReference(u).String()

	img, err := ebitenutil.NewImageFromURL(finalUrl)
	if err != nil {
		log.Println("Load icon from ffreplay failed", finalUrl, err)

		return nil
	}

	return img
}

func loadFromRPGLogs(iconName string) *ebiten.Image {
	u, err := url.Parse("https://assets.rpglogs.com/img/ff/abilities/" + iconName)
	if err != nil {
		log.Println(err)

		return nil
	}

	finalUrl := u.ResolveReference(u).String()

	img, err := ebitenutil.NewImageFromURL(finalUrl)
	if err != nil {
		log.Println("Load icon from fflogs failed", finalUrl, err)

		return nil
	}

	return img
}

func loadFromLocal(iconName string) *ebiten.Image {
	if _, err := os.Stat("public/img/ff/abilities/" + iconName); err == nil {
		img, _, err := ebitenutil.NewImageFromFile("public/img/ff/abilities/" + iconName)
		if err != nil {
			log.Println("Load icon from local file failed", err)

			return nil
		}

		return img
	}

	log.Println("Missing icon file", "public/img/ff/abilities/"+iconName)

	return nil
}

func downloadAndLoadIcon(iconName string) *ebiten.Image {
	u, err := url.Parse("https://assets.rpglogs.com/img/ff/abilities/" + iconName)
	if err != nil {
		log.Panic(err)
	}

	finalUrl := u.ResolveReference(u).String()

	// download image to local file
	resp, err := http.Get(finalUrl)
	if err != nil {
		log.Println("Load icon from fflogs failed", finalUrl)

		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Load icon from fflogs failed", finalUrl, resp.StatusCode)

		return nil
	}

	// write to local file
	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Dump response failed", finalUrl)

		return nil
	}

	os.MkdirAll("public/img/ff/abilities", os.ModePerm)

	err = os.WriteFile("public/img/ff/abilities/"+iconName, imgData, 0644)
	if err != nil {
		log.Println("Write icon to local file failed", finalUrl)

		return nil
	}

	img, _, err := ebitenutil.NewImageFromFile("public/img/ff/abilities/" + iconName)
	if err != nil {
		log.Println("Load icon from local file failed", finalUrl, err)

		return nil
	}

	return img
}

func NewAbilityTexture(iconName string) *ebiten.Image {
	if strings.Contains(iconName, "warcraft") {
		return nil
	}

	if texture, ok := textureCache.Load(iconName); ok {
		if value, ok := texture.(*ebiten.Image); ok {
			return value
		}

		return nil
	}

	// not using local buff icon files
	if util.IsWasm() {
		img := loadFromFFreplay(iconName)
		if img != nil {
			textureCache.Store(iconName, img)

			return img
		}

		img = loadFromRPGLogs(iconName)
		if img != nil {
			textureCache.Store(iconName, img)

			return img
		}

		log.Println("Load icon from ffreplay and fflogs failed", iconName)

		return nil
	}

	// try load from local

	img := loadFromLocal(iconName)
	if img != nil {
		textureCache.Store(iconName, img)

		return img
	}

	// try download from fflogs
	img = downloadAndLoadIcon(iconName)
	if img != nil {
		textureCache.Store(iconName, img)

		return img
	}

	log.Println("Load icon from local file failed", iconName)

	return nil
}

func NewMapTexture(mapName string) *ebiten.Image {
	if texture, ok := textureCache.Load(mapName); ok {
		if value, ok := texture.(*ebiten.Image); ok {
			return value
		}

		return nil
	}

	u, err := url.Parse("https://assets.rpglogs.com/img/ff/maps/" + mapName)
	if err != nil {
		log.Panic(err)
	}

	finalUrl := u.ResolveReference(u).String()
	// not using local buff icon files
	img, err := ebitenutil.NewImageFromURL(finalUrl)
	if err != nil {
		log.Println("Load map from fflogs failed", finalUrl)

		return nil
	}

	textureCache.Store(mapName, img)

	return img
}

func NewTextureFromFile(filepath string) *ebiten.Image {
	if texture, ok := textureCache.Load(filepath); ok {
		if value, ok := texture.(*ebiten.Image); ok {
			return value
		}

		return nil
	}

	f, err := asset.AssetFS.Open(filepath)
	if err != nil {
		log.Panic(err)
	}

	img, _, err := ebitenutil.NewImageFromReader(f)
	if err != nil {
		log.Panic(err)
	}

	textureCache.Store(filepath, img)

	return img
}

func CenterGeoM(t *ebiten.Image) ebiten.GeoM {
	geoM := ebiten.GeoM{}
	geoM.Translate(-float64(t.Bounds().Dx())/2, -float64(t.Bounds().Dy())/2)

	return geoM
}
