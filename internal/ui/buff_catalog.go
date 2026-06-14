package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Xinrea/ffreplay/internal/model"
)

// BuffCatalogEntry describes a buff that can be picked in editor UIs.
type BuffCatalogEntry struct {
	ID   int64
	Name string
	Icon string
}

var (
	buffCatalogOnce sync.Once
	buffCatalog     []BuffCatalogEntry
	buffCatalogByID map[int64]*BuffCatalogEntry
)

// BuffCatalog returns sorted buff catalog entries from local ability icons.
func BuffCatalog() []BuffCatalogEntry {
	loadBuffCatalog()
	return buffCatalog
}

// BuffCatalogLookup finds a catalog entry by buff id.
func BuffCatalogLookup(id int64) *BuffCatalogEntry {
	loadBuffCatalog()
	if e, ok := buffCatalogByID[id]; ok {
		return e
	}
	return nil
}

func loadBuffCatalog() {
	buffCatalogOnce.Do(func() {
		byID := map[int64]*BuffCatalogEntry{}
		paths, err := filepath.Glob("asset/abilities/*.png")
		if err != nil {
			paths = nil
		}

		for _, path := range paths {
			base := strings.TrimSuffix(filepath.Base(path), ".png")
			id := parseBuffIconID(base)
			if id <= 0 {
				continue
			}

			name := ""
			icon := base
			if info := model.GetBuffInfo(id); info != nil {
				if info.Name != "" {
					name = info.Name
				}
				if info.Icon != "" {
					icon = info.Icon
				}
			}
			if name == "" {
				name = base
			}

			byID[id] = &BuffCatalogEntry{ID: id, Name: name, Icon: icon}
		}

		buffCatalog = make([]BuffCatalogEntry, 0, len(byID))
		for _, entry := range byID {
			buffCatalog = append(buffCatalog, *entry)
		}
		sort.Slice(buffCatalog, func(i, j int) bool {
			if buffCatalog[i].Name != buffCatalog[j].Name {
				return buffCatalog[i].Name < buffCatalog[j].Name
			}
			return buffCatalog[i].ID < buffCatalog[j].ID
		})

		buffCatalogByID = make(map[int64]*BuffCatalogEntry, len(buffCatalog))
		for i := range buffCatalog {
			buffCatalogByID[buffCatalog[i].ID] = &buffCatalog[i]
		}
	})
}

func parseBuffIconID(iconBase string) int64 {
	if dash := strings.LastIndex(iconBase, "-"); dash >= 0 && dash+1 < len(iconBase) {
		if id, err := strconv.ParseInt(iconBase[dash+1:], 10, 64); err == nil {
			return id
		}
	}
	if id, err := strconv.ParseInt(iconBase, 10, 64); err == nil {
		return id
	}
	return 0
}

// ensureBuffCatalogFromAssets allows tests/runtime to warm the catalog without reading the whole tree.
func ensureBuffCatalogFromAssets() {
	if _, err := os.Stat("asset/abilities"); err != nil {
		return
	}
	loadBuffCatalog()
}
