package ctgp

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed wire_catalog.json
var wireCatalogJSON []byte

// A wire slot identifies a track REVISION and ruleset, never a donor/cup position.
// Historical slots are retained in the append-only registry, even after pack edits.
type WireTrack struct {
	ID        string `json:"id"`
	Revision  string `json:"revision"`
	BaseSlot  uint16 `json:"base_slot"`
	Directory string `json:"directory"`
	Name      string `json:"name"`
}

type WireBoard struct {
	Track           WireTrack
	CC              uint16
	Pink            bool
	PersistenceSlot uint16
	Category        uint32
}

type WireCatalog struct{ slots map[uint16]WireBoard }

func LoadWireCatalog(data []byte) (*WireCatalog, error) {
	var manifest struct {
		Protocol       int         `json:"protocol"`
		Base           int         `json:"base"`
		Limit          int         `json:"limit"`
		CategoryOffset int         `json:"category_offset"`
		Variants       []string    `json:"variants"`
		Tracks         []WireTrack `json:"tracks"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest.Protocol != 1 || manifest.Base != 128 || manifest.Limit != 1024 || manifest.CategoryOffset != 16 {
		return nil, fmt.Errorf("unsupported CTGP wire catalog")
	}
	expected := []string{"150-native", "200-native", "150-pink-v1", "200-pink-v1"}
	if len(manifest.Variants) != len(expected) {
		return nil, fmt.Errorf("invalid variants")
	}
	for i := range expected {
		if manifest.Variants[i] != expected[i] {
			return nil, fmt.Errorf("invalid variant order")
		}
	}
	catalog := &WireCatalog{slots: make(map[uint16]WireBoard)}
	identities := make(map[string]bool)
	for _, track := range manifest.Tracks {
		key := track.ID + "/" + track.Revision
		if track.ID == "" || len(track.Revision) != 64 || identities[key] || track.BaseSlot < 128 || track.BaseSlot%4 != 0 || track.BaseSlot > 1020 {
			return nil, fmt.Errorf("invalid CTGP track %q", track.ID)
		}
		identities[key] = true
		for variant := uint16(0); variant < 4; variant++ {
			slot := track.BaseSlot + variant
			if _, exists := catalog.slots[slot]; exists {
				return nil, fmt.Errorf("duplicate slot %d", slot)
			}
			cc := uint16(150)
			if variant&1 != 0 {
				cc = 200
			}
			catalog.slots[slot] = WireBoard{track, cc, variant&2 != 0, slot, uint32(slot) + 16}
		}
	}
	return catalog, nil
}

func (catalog *WireCatalog) BySlot(slot uint16) (WireBoard, bool) {
	board, ok := catalog.slots[slot]
	return board, ok
}

func (catalog *WireCatalog) ByCategory(category uint32) (WireBoard, bool) {
	if category < 144 || category >= 1040 {
		return WireBoard{}, false
	}
	return catalog.BySlot(uint16(category - 16))
}

var Boards = func() *WireCatalog {
	catalog, err := LoadWireCatalog(wireCatalogJSON)
	if err != nil {
		panic(err)
	}
	return catalog
}()
