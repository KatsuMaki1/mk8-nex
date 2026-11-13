package ctgp

import (
	"encoding/binary"
	"hash/crc32"
	"testing"
)

func container(n int) []byte {
	d := make([]byte, ghostHeader+n)
	copy(d, "CTG0")
	binary.BigEndian.PutUint32(d[8:], uint32(len(d)))
	binary.BigEndian.PutUint32(d[0x38:], crc32.ChecksumIEEE(d[ghostHeader:]))
	return d
}

func TestValidateGhost(t *testing.T) {
	if err := ValidateGhost(ClientCTGP, container(0x800)); err != nil {
		t.Fatal(err)
	}
	bad := container(0x800)
	bad[ghostHeader+5] ^= 1
	if ValidateGhost(ClientCTGP, bad) == nil {
		t.Fatal("corrupted body accepted")
	}
	if err := ValidateGhost(ClientCTGP, make([]byte, 0x800)); err != nil {
		t.Fatal("native CTGP NEX upload rejected:", err)
	}
	if err := ValidateGhost(ClientVanilla, container(0x800)); err != nil {
		t.Fatal("vanilla Wii U save container rejected:", err)
	}
	if err := ValidateGhost(ClientVanilla, make([]byte, 0x800)); err != nil {
		t.Fatal(err)
	}
}
