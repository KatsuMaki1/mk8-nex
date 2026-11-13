package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/lib/pq"
)

func GetObjectInfoByDataID(dataID types.UInt64) (datastore_types.DataStoreMetaInfo, *nex.Error) {
	var pqErr *pq.Error

	// Same NULL-safe explicit column list as GetObjectInfosByDataStoreSearchParam
	// (SELECT * scans straight into non-nullable Go types and crashes on a
	// NULL refer_data_id - see that file for the full story).
	rows, err := Postgres.Query(`
		SELECT
			data_id,
			upload_completed,
			deleted,
			COALESCE(owner, 0),
			COALESCE(size, 0),
			COALESCE(name, ''),
			COALESCE(data_type, 0),
			meta_binary,
			COALESCE(permission, 0),
			permission_recipients,
			COALESCE(delete_permission, 0),
			delete_permission_recipients,
			COALESCE(flag, 0),
			COALESCE(period, 0),
			COALESCE(refer_data_id, 0),
			tags,
			COALESCE(persistence_slot_id, 0),
			extra_data,
			access_password,
			update_password,
			COALESCE(creation_date, to_timestamp(0)),
			COALESCE(update_date, to_timestamp(0))
		FROM datastore.objects WHERE data_id = $1 AND upload_completed = TRUE AND deleted = FALSE
	`,
		int64(dataID),
	)
	if errors.Is(err, sql.ErrNoRows) || (errors.As(err, &pqErr) && pqErr.SQLState() == "42P01") {
		return datastore_types.DataStoreMetaInfo{}, nex.NewError(nex.ResultCodes.DataStore.NotFound, "not found")
	}
	if err != nil {
		fmt.Printf("Error selecting datastore object: %v\n", err)
		return datastore_types.DataStoreMetaInfo{}, nex.NewError(nex.ResultCodes.DataStore.Unknown, "unknown")
	}
	defer rows.Close()

	results, err := parseObjectList(rows)
	if err != nil || len(results) == 0 {
		return datastore_types.DataStoreMetaInfo{}, nex.NewError(nex.ResultCodes.DataStore.NotFound, "not found")
	}

	return results[0], nil
}

// The NEX DataStoreMetaInfo response does not expose the persistence slot.
// Keep this server-side lookup separate from the protocol response structure.
func GetObjectPersistenceSlot(dataID types.UInt64) (uint16, error) {
	var slot int64
	err := Postgres.QueryRow(`SELECT COALESCE(persistence_slot_id, 0)
		FROM datastore.objects WHERE data_id=$1 AND upload_completed=TRUE AND deleted=FALSE`, int64(dataID)).Scan(&slot)
	if err != nil {
		return 0, err
	}
	if slot < 0 || slot > 65535 {
		return 0, fmt.Errorf("invalid persistence slot %d", slot)
	}
	return uint16(slot), nil
}
