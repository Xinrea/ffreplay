package texture

import (
	"log"
	"net/url"
	"sync"

	asset "github.com/Xinrea/ffreplay"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var textureCache = sync.Map{}

func NewAbilityTexture(iconName string) *ebiten.Image {
	if texture, ok := textureCache.Load(iconName); ok {
		if value, ok := texture.(*ebiten.Image); ok {
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
	img, err := ebitenutil.NewImageFromURL(finalUrl)
	if err != nil {
		log.Println("Load icon from fflogs failed")
		return nil
	}
	textureCache.Store(iconName, img)
	return img
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
		log.Fatal(err)
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
	var err error
	f, err := asset.AssetFS.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := ebitenutil.NewImageFromReader(f)
	if err != nil {
		log.Fatal(err)
	}
	textureCache.Store(filepath, img)
	return img
}

func CenterGeoM(t *ebiten.Image) ebiten.GeoM {
	geoM := ebiten.GeoM{}
	geoM.Translate(-float64(t.Bounds().Dx())/2, -float64(t.Bounds().Dy())/2)
	return geoM
}
