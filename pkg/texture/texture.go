package texture

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Texture struct {
	asset *ebiten.Image
}

var textureCache = sync.Map{}

func NewBuffTexture(iconName string) *Texture {
	if texture, ok := textureCache.Load(iconName); ok {
		if value, ok := texture.(*Texture); ok {
			return value
		}
		return nil
	}
	u, err := url.Parse("https://assets.rpglogs.com/img/ff/abilities/" + iconName)
	if err != nil {
		log.Fatal(err)
	}
	finalUrl := u.ResolveReference(u).String()
	// not using local buff icon files
	if util.IsWasm() {
		img, err := ebitenutil.NewImageFromURL(finalUrl)
		if err != nil {
			log.Println("Load icon from fflogs failed")
			return nil
		}
		texture := &Texture{
			asset: img,
		}
		textureCache.Store(iconName, texture)
		return texture
	}
	img, _, err := ebitenutil.NewImageFromFile("asset/buff/" + iconName)
	if err != nil {
		// if not in wasm, save icon file
		log.Println("Buff icon not found, downloading from fflogs", iconName)
		// download image from fflogs
		resp, err := http.Get(finalUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		img, _, err = ebitenutil.NewImageFromReader(bytes.NewReader(data))
		if err != nil {
			log.Fatal("Create image from url failed:", finalUrl)
		}
		// save image to file
		filePath := "asset/buff/" + iconName
		file, err := os.Create(filePath)
		if err != nil {
			// it's ok to failed caching, for example: ../../warcraft/abilities/ability_hunter_assassinate2.jpg icon should not cached
			log.Println(err)
		} else {
			defer file.Close()
			_, err = file.Write(data)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	texture := &Texture{
		asset: img,
	}
	textureCache.Store(iconName, texture)
	return texture
}

func NewTextureFromFile(filepath string) *Texture {
	if texture, ok := textureCache.Load(filepath); ok {
		if value, ok := texture.(*Texture); ok {
			return value
		}
		return nil
	}
	var err error
	img, _, err := ebitenutil.NewImageFromFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	texture := &Texture{
		asset: img,
	}
	textureCache.Store(filepath, texture)
	return texture
}

func NewTextureFromImage(img *ebiten.Image) *Texture {
	return &Texture{
		asset: img,
	}
}

func (t *Texture) Img() *ebiten.Image {
	return t.asset
}

func (t Texture) GetGeoM() ebiten.GeoM {
	geoM := ebiten.GeoM{}
	geoM.Translate(-float64(t.asset.Bounds().Dx())/2, -float64(t.asset.Bounds().Dy())/2)
	return geoM
}
