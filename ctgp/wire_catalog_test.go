package ctgp

import "testing"

func TestWireBoardsSeparateAllRulesAndTracks(t *testing.T) {
	seen := make(map[uint32]bool)
	for slot := uint16(128); slot < 1024; slot += 4 {
		first, exists := Boards.BySlot(slot)
		if !exists {
			continue
		}
		for variant := uint16(0); variant < 4; variant++ {
			board, ok := Boards.BySlot(slot + variant)
			if !ok || board.Track.ID != first.Track.ID || board.Pink != (variant&2 != 0) || (board.CC == 200) != (variant&1 != 0) {
				t.Fatalf("incorrect board at %d", slot+variant)
			}
			if seen[board.Category] {
				t.Fatal("category collision")
			}
			seen[board.Category] = true
			decoded, ok := Boards.ByCategory(board.Category)
			if !ok || decoded != board {
				t.Fatal("category roundtrip")
			}
		}
	}
	if len(seen) != 688 {
		t.Fatalf("expected 172 revisions x 4, got %d", len(seen))
	}
	for _, category := range []uint32{0, 29, 55, 143, 1040, 0xffffffff} {
		if _, ok := Boards.ByCategory(category); ok {
			t.Fatalf("accepted native/invalid category %d", category)
		}
	}
}

func TestWireCatalogRejectsCollisions(t *testing.T) {
	bad := []byte(`{"protocol":1,"base":128,"limit":1024,"category_offset":16,"variants":[]}`)
	if _, err := LoadWireCatalog(bad); err == nil {
		t.Fatal("invalid catalog accepted")
	}
}
