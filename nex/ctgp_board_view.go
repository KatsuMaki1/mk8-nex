package nex

import (
	"sync"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/types"
	common_datastore "github.com/PretendoNetwork/nex-protocols-common-go/v2/datastore"
	common_ranking "github.com/PretendoNetwork/nex-protocols-common-go/v2/ranking"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"

	"github.com/KatsuMaki1/mk8-nex/ctgp"
)

// CTGP ghost download fallback. Uploads reach the custom-track board slot (128+), but when a player opens a ghost from
// a custom-track leaderboard the client asks for the donor's native slot (<128) and finds nothing (empty "Rabid" ghost,
// race ends after the countdown). The client always queried that board just before, so remember it per requester.
type viewedBoard struct {
	category uint32
	at       time.Time
}

var (
	viewedMu     sync.Mutex
	viewedBoards = map[uint64]viewedBoard{}
)

func init() {
	common_ranking.OnGetRanking = func(pid uint64, category uint32) {
		viewedMu.Lock()
		defer viewedMu.Unlock()
		if _, ok := ctgp.Boards.ByCategory(category); ok {
			viewedBoards[pid] = viewedBoard{category, time.Now()}
		} else {
			delete(viewedBoards, pid) // a native board: native slots are correct
		}
	}
	common_datastore.RemapPersistenceTarget = func(requester uint64, target datastore_types.DataStorePersistenceTarget) (datastore_types.DataStorePersistenceTarget, bool) {
		if uint16(target.PersistenceSlotID) >= 128 {
			return target, false
		}
		viewedMu.Lock()
		view, ok := viewedBoards[requester]
		viewedMu.Unlock()
		if !ok || time.Since(view.at) > 10*time.Minute {
			return target, false
		}
		board, ok := ctgp.Boards.ByCategory(view.category)
		if !ok {
			return target, false
		}
		remapped := target.Copy().(datastore_types.DataStorePersistenceTarget)
		remapped.PersistenceSlotID = types.NewUInt16(board.PersistenceSlot)
		return remapped, true
	}
}
