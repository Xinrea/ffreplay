package texture

import (
	"encoding/xml"
	"image"

	asset "github.com/Xinrea/ffreplay"
	"github.com/hajimehoshi/ebiten/v2"
)

type TextureAtlas struct {
	MainTexture *ebiten.Image         `xml:"-"`
	SubTextures []*SubTexture         `xml:"SubTexture"`
	seachMap    map[string]*NineSlice `xml:"-"`
}

func (a *TextureAtlas) GetNineSlice(name string) *NineSlice {
	if a.seachMap == nil {
		a.seachMap = map[string]*NineSlice{}
		for _, sub := range a.SubTextures {
			a.seachMap[sub.Name] = sub.NiceSlice
		}
	}
	return a.seachMap[name]
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
	bytes, err := asset.AssetFS.ReadFile(file)
	if err != nil {
		panic(err)
	}
	// parse xml
	textureAtlas := &TextureAtlas{}
	err = xml.Unmarshal(bytes, textureAtlas)
	if err != nil {
		panic(err)
	}
	// load main texture
	pngFile := file[:len(file)-len("xml")] + "png"
	textureAtlas.MainTexture = NewTextureFromFile(pngFile)
	for i := range textureAtlas.SubTextures {
		textureAtlas.SubTextures[i].InitNineSlice(textureAtlas.MainTexture)
	}
	atlasCache[file] = textureAtlas
	return textureAtlas
}

func (s *SubTexture) InitNineSlice(mainTexture *ebiten.Image) {
	subImage := ebiten.NewImageFromImage(mainTexture.SubImage(image.Rect(s.X, s.Y, s.X+s.W, s.Y+s.H)))
	s.NiceSlice = NewNineSlice(subImage, s.Top, s.Bottom, s.Left, s.Right)
}
