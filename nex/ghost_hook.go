package nex

import (
	"context"
	"fmt"
	"io"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/minio/minio-go/v7"

	"github.com/KatsuMaki1/mk8-nex/ctgp"
	"github.com/KatsuMaki1/mk8-nex/database"
	"github.com/KatsuMaki1/mk8-nex/globals"
)

// ghostHook validates every completed DataStore upload (auto-approve): a ghost that fails the integrity check is removed from S3 and the
// database. CTG0 is a native Wii U save wrapper, not a CTGP identifier. A reserved
// persistence slot declares a CTGP board; it does not attest the client binary.
func ghostHook(client *minio.Client, bucket, keyBase string, deleteObject func(types.UInt64) *nex.Error) func(packet nex.PacketInterface, param datastore_types.DataStoreCompletePostParam) {
	return func(packet nex.PacketInterface, param datastore_types.DataStoreCompletePostParam) {
		if !bool(param.IsSuccess) {
			return
		}
		dataID := uint64(param.DataID)
		pid := uint64(packet.Sender().PID())
		key := fmt.Sprintf("%s/%d.bin", keyBase, dataID)
		obj, err := client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
		if err != nil {
			globals.Logger.Errorf("[MK8 ghost] data_id=%d pid=%d: cannot read object: %v", dataID, pid, err)
			return
		}
		defer obj.Close()
		data, err := io.ReadAll(io.LimitReader(obj, 4<<20))
		if err != nil {
			globals.Logger.Errorf("[MK8 ghost] data_id=%d pid=%d: read failed: %v", dataID, pid, err)
			return
		}
		format := "native-replay"
		if len(data) >= 4 && string(data[:4]) == "CTG0" {
			format = "native-save-container"
		}
		slot, slotError := database.GetObjectPersistenceSlot(param.DataID)
		if slotError != nil {
			globals.Logger.Warningf("[MK8 ghost] data_id=%d: cannot resolve persistence slot: %v", dataID, slotError)
		}
		board, custom := ctgp.Boards.BySlot(slot)
		globals.Logger.Infof("[MK8 ghost] data_id=%d pid=%d size=%d format=%s persistence=%d ctgp_board=%t",
			dataID, pid, len(data), format, slot, custom)
		if custom {
			globals.Logger.Infof("[CTGP ghost] data_id=%d track=%s cc=%d pink=%t category=%d", dataID, board.Track.ID, board.CC, board.Pink, board.Category)
		}
		if verr := ctgp.ValidateGhost("", data); verr != nil {
			globals.Logger.Warningf("[MK8 ghost] data_id=%d REJECTED: %v", dataID, verr)
			_ = client.RemoveObject(context.Background(), bucket, key, minio.RemoveObjectOptions{})
			deleteObject(param.DataID)
		}
	}
}
