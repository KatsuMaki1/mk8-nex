package ctgp

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// Native Wii U save container (CTG0): 0x48-byte header, then the recording.
// Header: "CTG0" at 0, big-endian total file size at 8, big-endian CRC-32 of the body (bytes 0x48..) at 0x38.
// This proves integrity (a truncated or corrupted upload is refused), NOT that the run is genuine.
const (
	ghostHeader  = 0x48
	ghostMinBody = 0x400
	ghostMaxSize = 2 << 20
)

// ValidateGhost checks the transport format. Both vanilla and CTGP use native
// recordings; saves have a CTG0 wrapper while NEX uploads may omit it.
// Neither format identifies/authenticates a modded client.
func ValidateGhost(_ string, data []byte) error {
	if len(data) == 0 || len(data) > ghostMaxSize {
		return errors.New("ghost size out of range")
	}
	container := len(data) >= 4 && string(data[:4]) == "CTG0"
	if !container {
		if len(data) < ghostMinBody {
			return errors.New("ghost too small")
		}
		return nil
	}
	if len(data) < ghostHeader+ghostMinBody {
		return errors.New("truncated CTG0 ghost")
	}
	if binary.BigEndian.Uint32(data[8:]) != uint32(len(data)) {
		return errors.New("CTG0 size mismatch")
	}
	if binary.BigEndian.Uint32(data[0x38:]) != crc32.ChecksumIEEE(data[ghostHeader:]) {
		return errors.New("CTG0 checksum mismatch")
	}
	return nil
}

const (
	ClientVanilla = "vanilla"
	ClientCTGP    = "ctgp-cafe"
)
