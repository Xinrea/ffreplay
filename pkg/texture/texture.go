package texture

import (
	"image/color"
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

// abilityIconFileName normalizes RPG Logs ability icon names to a local filename.
func abilityIconFileName(iconName string) string {
	iconName = strings.TrimSpace(iconName)
	if iconName == "" {
		return ""
	}
	if strings.HasSuffix(strings.ToLower(iconName), ".png") {
		return iconName
	}
	return iconName + ".png"
}

func abilityIconCandidates(iconName string) []string {
	iconName = strings.TrimSpace(iconName)
	if iconName == "" {
		return nil
	}
	fileName := abilityIconFileName(iconName)
	if fileName == iconName {
		return []string{fileName}
	}
	return []string{fileName, iconName}
}

func abilityIconURL(iconName string) string {
	return "https://assets.rpglogs.com/img/ff/abilities/" + abilityIconFileName(iconName)
}

func missingAbilityTexture() *ebiten.Image {
	const cacheKey = "__missing_ability_icon__"

	if texture, ok := textureCache.Load(cacheKey); ok {
		if value, ok := texture.(*ebiten.Image); ok {
			return value
		}
	}

	img := ebiten.NewImage(24, 24)
	img.Fill(color.NRGBA{50, 50, 56, 220})
	textureCache.Store(cacheKey, img)

	return img
}

func loadFromFFreplay(iconName string) *ebiten.Image {
	for _, candidate := range abilityIconCandidates(iconName) {
		img, _, err := ebitenutil.NewImageFromFileSystem(asset.AssetFS, "asset/abilities/"+candidate)
		if err == nil && img != nil {
			return img
		}
	}
	log.Println("Load icon from ffreplay failed", iconName)

	return nil
}

func loadFromRPGLogs(iconName string) *ebiten.Image {
	u, err := url.Parse(abilityIconURL(iconName))
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
	for _, candidate := range abilityIconCandidates(iconName) {
		path := "asset/abilities/" + candidate
		if _, err := os.Stat(path); err != nil {
			continue
		}
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			log.Println("Load icon from local file failed", path, err)
			continue
		}
		return img
	}

	log.Println("Missing icon file", "asset/abilities/"+abilityIconFileName(iconName))

	return nil
}

func downloadAndLoadIcon(iconName string) *ebiten.Image {
	u, err := url.Parse(abilityIconURL(iconName))
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

	os.MkdirAll("asset/abilities/", os.ModePerm)

	localPath := "asset/abilities/" + abilityIconFileName(iconName)
	err = os.WriteFile(localPath, imgData, 0644)
	if err != nil {
		log.Println("Write icon to local file failed", finalUrl)

		return nil
	}

	img, _, err := ebitenutil.NewImageFromFile(localPath)
	if err != nil {
		log.Println("Load icon from local file failed", finalUrl, err)

		return nil
	}

	return img
}

func NewAbilityTexture(iconName string) *ebiten.Image {
	iconName = strings.TrimSpace(iconName)
	if iconName == "" {
		return missingAbilityTexture()
	}

	if strings.Contains(iconName, "warcraft") {
		return nil
	}

	cacheKey := abilityIconFileName(iconName)
	if texture, ok := textureCache.Load(cacheKey); ok {
		if value, ok := texture.(*ebiten.Image); ok {
			return value
		}

		return nil
	}

	store := func(img *ebiten.Image) *ebiten.Image {
		if img != nil {
			textureCache.Store(cacheKey, img)
		}
		return img
	}

	// not using local buff icon files
	if util.IsWasm() {
		if img := store(loadFromFFreplay(iconName)); img != nil {
			return img
		}
		if img := store(loadFromRPGLogs(iconName)); img != nil {
			return img
		}

		log.Println("Load icon from ffreplay and fflogs failed", cacheKey)

		img := missingAbilityTexture()
		textureCache.Store(cacheKey, img)

		return img
	}

	if img := store(loadFromLocal(iconName)); img != nil {
		return img
	}
	if img := store(loadFromFFreplay(iconName)); img != nil {
		return img
	}
	if img := store(downloadAndLoadIcon(iconName)); img != nil {
		return img
	}

	log.Println("Load icon from local file failed", cacheKey)

	img := missingAbilityTexture()
	textureCache.Store(cacheKey, img)

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
