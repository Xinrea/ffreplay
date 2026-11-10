// Command abilityatlas packs the ability icon PNGs into a single self-contained
// atlas file so the browser build can load and slice one file instead of
// fetching every icon separately. The icon index and the image data live in
// the same file, so a cached index can never point at a mismatched image.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
)

const (
	atlasMagic    = "FATL"
	atlasFormat   = 1
	atlasFileName = "atlas-v1.bin"
)

type atlasRect struct {
	Page int `json:"p"`
	X    int `json:"x"`
	Y    int `json:"y"`
	W    int `json:"w"`
	H    int `json:"h"`
}

type atlasPage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Offset int `json:"offset"`
	Length int `json:"length"`
}

type atlasManifest struct {
	Format int                  `json:"format"`
	Pages  []atlasPage          `json:"pages"`
	Icons  map[string]atlasRect `json:"icons"`
}

func main() {
	in := flag.String("in", "asset/abilities", "directory containing ability icon PNGs")
	out := flag.String("out", "public/asset/abilities", "directory for the generated atlas")
	maxPage := flag.Int("max-page", 2048, "maximum atlas page width and height")
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*in, "*.png"))
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatalf("no PNG icons found in %s", *in)
	}
	sort.Strings(files)

	icons := make([]image.Image, len(files))

	var cellWidth, cellHeight int

	for i, file := range files {
		icons[i] = decodePNG(file)

		bounds := icons[i].Bounds()
		if bounds.Dx() > cellWidth {
			cellWidth = bounds.Dx()
		}
		if bounds.Dy() > cellHeight {
			cellHeight = bounds.Dy()
		}
	}

	columns := *maxPage / cellWidth
	rows := *maxPage / cellHeight

	if columns == 0 || rows == 0 {
		log.Fatalf("icon size %dx%d does not fit in a %dx%d page", cellWidth, cellHeight, *maxPage, *maxPage)
	}

	perPage := columns * rows
	pageCount := (len(files) + perPage - 1) / perPage

	manifest := atlasManifest{
		Format: atlasFormat,
		Icons:  make(map[string]atlasRect, len(files)),
	}
	pageBlobs := make([][]byte, 0, pageCount)

	for page := range pageCount {
		start := page * perPage
		end := min(start+perPage, len(files))
		count := end - start
		pageRows := (count + columns - 1) / columns

		sheet := image.NewNRGBA(image.Rect(0, 0, columns*cellWidth, pageRows*cellHeight))

		for i, icon := range icons[start:end] {
			x := (i % columns) * cellWidth
			y := (i / columns) * cellHeight
			bounds := icon.Bounds()
			draw.Draw(sheet, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), icon, bounds.Min, draw.Src)

			manifest.Icons[filepath.Base(files[start+i])] = atlasRect{
				Page: page,
				X:    x,
				Y:    y,
				W:    bounds.Dx(),
				H:    bounds.Dy(),
			}
		}

		var blob bytes.Buffer

		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		if err := encoder.Encode(&blob, sheet); err != nil {
			log.Fatal(err)
		}

		manifest.Pages = append(manifest.Pages, atlasPage{
			Width:  sheet.Bounds().Dx(),
			Height: sheet.Bounds().Dy(),
		})
		pageBlobs = append(pageBlobs, blob.Bytes())
	}

	offset := 0
	for i := range manifest.Pages {
		manifest.Pages[i].Offset = offset
		manifest.Pages[i].Length = len(pageBlobs[i])
		offset += len(pageBlobs[i])
	}

	header, err := json.Marshal(manifest)
	if err != nil {
		log.Fatal(err)
	}

	var file bytes.Buffer

	file.WriteString(atlasMagic)

	if err := binary.Write(&file, binary.BigEndian, uint32(len(header))); err != nil {
		log.Fatal(err)
	}

	file.Write(header)

	for _, blob := range pageBlobs {
		file.Write(blob)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(*out, atlasFileName)
	if err := os.WriteFile(path, file.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}

	for i, page := range manifest.Pages {
		log.Printf("page %d: %dx%d, %d bytes", i, page.Width, page.Height, page.Length)
	}
	log.Printf("packed %d icons into %s (%.1f KB), cell %dx%d",
		len(files), path, float64(file.Len())/1024, cellWidth, cellHeight)
}

func decodePNG(file string) image.Image {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatalf("decode %s: %v", file, err)
	}

	return img
}
