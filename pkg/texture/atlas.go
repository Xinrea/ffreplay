package texture

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"sync"

	asset "github.com/Xinrea/ffreplay"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// The ability atlas is a single file: a small JSON header describing the icons
// followed by the PNG pages. Keeping the index and the images in one file means
// a cached copy is always self-consistent, even when a CDN serves different
// versions of different files after a deploy.
const (
	abilityAtlasPath   = "asset/abilities/atlas-v1.bin"
	abilityAtlasMagic  = "FATL"
	abilityAtlasFormat = 1
)

type abilityAtlasRect struct {
	Page int `json:"p"`
	X    int `json:"x"`
	Y    int `json:"y"`
	W    int `json:"w"`
	H    int `json:"h"`
}

type abilityAtlasPage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Offset int `json:"offset"`
	Length int `json:"length"`
}

type abilityAtlasManifest struct {
	Format int                         `json:"format"`
	Pages  []abilityAtlasPage          `json:"pages"`
	Icons  map[string]abilityAtlasRect `json:"icons"`
}

var (
	abilityAtlasOnce sync.Once
	abilityAtlas     abilityAtlasManifest
	abilityAtlasBlob []byte
	abilityAtlasErr  error
)

// PrefetchAbilityIcons warms the icon atlas in the background so the first
// rendered skill does not have to wait for the file download.
func PrefetchAbilityIcons() {
	if !util.IsWasm() {
		return
	}

	go func() {
		_, _ = loadAbilityAtlas()
	}()
}

func loadAbilityAtlas() (*abilityAtlasManifest, error) {
	abilityAtlasOnce.Do(func() {
		data, err := fs.ReadFile(asset.AssetFS, abilityAtlasPath)
		if err != nil {
			abilityAtlasErr = err

			return
		}

		abilityAtlas, abilityAtlasBlob, abilityAtlasErr = splitAbilityAtlas(data)
	})

	if abilityAtlasErr != nil {
		return nil, abilityAtlasErr
	}

	return &abilityAtlas, nil
}

func splitAbilityAtlas(data []byte) (abilityAtlasManifest, []byte, error) {
	var manifest abilityAtlasManifest

	if len(data) < 8 || string(data[:4]) != abilityAtlasMagic {
		return manifest, nil, errors.New("invalid ability atlas header")
	}

	headerLength := int(binary.BigEndian.Uint32(data[4:8]))
	if headerLength <= 0 || 8+headerLength > len(data) {
		return manifest, nil, fmt.Errorf("invalid ability atlas header length %d", headerLength)
	}

	err := json.Unmarshal(data[8:8+headerLength], &manifest)
	if err != nil {
		return manifest, nil, fmt.Errorf("parse ability atlas header: %w", err)
	}

	if manifest.Format != abilityAtlasFormat {
		return manifest, nil, fmt.Errorf("unsupported ability atlas format %d", manifest.Format)
	}

	blob := data[8+headerLength:]

	err = validateAbilityAtlasPages(manifest.Pages, len(blob))
	if err != nil {
		return manifest, nil, err
	}

	return manifest, blob, nil
}

func validateAbilityAtlasPages(pages []abilityAtlasPage, blobLength int) error {
	for _, page := range pages {
		if page.Offset < 0 || page.Length <= 0 || page.Offset+page.Length > blobLength {
			return errors.New("ability atlas page out of bounds")
		}
	}

	return nil
}

func loadFromAbilityAtlas(iconName string) *ebiten.Image {
	manifest, err := loadAbilityAtlas()
	if err != nil {
		return nil
	}

	for _, candidate := range abilityIconCandidates(iconName) {
		rect, ok := manifest.Icons[candidate]
		if !ok {
			continue
		}

		page := loadAbilityAtlasPage(rect.Page)
		if page == nil {
			continue
		}

		if img, ok := page.SubImage(image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)).(*ebiten.Image); ok {
			return img
		}
	}

	return nil
}

func loadAbilityAtlasPage(index int) *ebiten.Image {
	key := fmt.Sprintf("__ability_atlas_page_%d__", index)

	if cached, ok := textureCache.Load(key); ok {
		if img, ok := cached.(*ebiten.Image); ok {
			return img
		}

		return nil
	}

	manifest, err := loadAbilityAtlas()
	if err != nil || index < 0 || index >= len(manifest.Pages) {
		return nil
	}

	page := manifest.Pages[index]
	data := abilityAtlasBlob[page.Offset : page.Offset+page.Length]

	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	textureCache.Store(key, img)

	return img
}
