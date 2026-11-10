package texture

import (
	"encoding/xml"
	"image"
	"io/fs"
	"log"

	asset "github.com/Xinrea/ffreplay"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
)

type TextureAtlas struct {
	MainTexture *ebiten.Image         `xml:"-"`
	SubTextures []*SubTexture         `xml:"SubTexture"`
	seachMap    map[string]*NineSlice `xml:"-"`
	fallback    bool
	fallbackNS  *NineSlice
}

func (a *TextureAtlas) GetNineSlice(name string) *NineSlice {
	if a.seachMap == nil {
		a.seachMap = map[string]*NineSlice{}
		for _, sub := range a.SubTextures {
			a.seachMap[sub.Name] = sub.NiceSlice
		}
	}

	if slice, ok := a.seachMap[name]; ok && slice != nil {
		return slice
	}

	if a.fallback && a.MainTexture != nil {
		if a.fallbackNS == nil {
			a.fallbackNS = NewNineSlice(a.MainTexture, 0, 0, 0, 0)
		}

		return a.fallbackNS
	}

	return nil
}

type SubTexture struct {
	Name      string     `xml:"name,attr"`
	X         int        `xml:"x,attr"`
	Y         int        `xml:"y,attr"`
	W         int        `xml:"w,attr"`
	H         int        `xml:"h,attr"`
	Top       int        `xml:"top,attr"`
	Bottom    int        `xml:"bottom,attr"`
	Left      int        `xml:"left,attr"`
	Right     int        `xml:"right,attr"`
	NiceSlice *NineSlice `xml:"-"`
}

var atlasCache = map[string]*TextureAtlas{}

func NewTextureAtlasFromFile(file string) *TextureAtlas {
	if atlas, ok := atlasCache[file]; ok {
		return atlas
	}

	bytes, err := fs.ReadFile(asset.AssetFS, file)
	if err != nil {
		return fallbackTextureAtlas(file, err)
	}

	// parse xml
	textureAtlas := &TextureAtlas{}

	err = xml.Unmarshal(bytes, textureAtlas)
	if err != nil {
		log.Println(string(bytes))

		return fallbackTextureAtlas(file, err)
	}

	// load main texture
	pngFile := file[:len(file)-len("xml")] + "png"

	mainTexture, err := loadTextureFile(pngFile)
	if err != nil {
		return fallbackTextureAtlas(file, err)
	}

	textureAtlas.MainTexture = mainTexture

	for i := range textureAtlas.SubTextures {
		textureAtlas.SubTextures[i].InitNineSlice(textureAtlas.MainTexture)
	}

	atlasCache[file] = textureAtlas

	return textureAtlas
}

func fallbackTextureAtlas(file string, err error) *TextureAtlas {
	if !util.IsWasm() {
		panic(err)
	}

	log.Printf("Load texture atlas %s failed, using placeholder: %v", file, err)

	atlas := &TextureAtlas{MainTexture: missingAbilityTexture(), fallback: true}
	atlasCache[file] = atlas

	return atlas
}

func (s *SubTexture) InitNineSlice(mainTexture *ebiten.Image) {
	bounds := mainTexture.Bounds()

	if s.W <= 0 || s.H <= 0 || s.X < 0 || s.Y < 0 || s.X+s.W > bounds.Dx() || s.Y+s.H > bounds.Dy() {
		log.Printf("Skip out-of-bounds atlas region %q", s.Name)

		return
	}

	subImage := ebiten.NewImageFromImage(mainTexture.SubImage(image.Rect(s.X, s.Y, s.X+s.W, s.Y+s.H)))
	s.NiceSlice = NewNineSlice(subImage, s.Top, s.Bottom, s.Left, s.Right)
}
