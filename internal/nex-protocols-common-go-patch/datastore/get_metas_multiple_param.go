package datastore

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
)

// RemapPersistenceTarget is tried when a persistence-target lookup finds nothing (MK8 CTGP ghosts: the client asks the donor's
// native slot although the ghost was uploaded to the custom-track board slot).
var RemapPersistenceTarget func(requester uint64, target datastore_types.DataStorePersistenceTarget) (datastore_types.DataStorePersistenceTarget, bool)

func (commonProtocol *CommonProtocol) getMetasMultipleParam(err error, packet nex.PacketInterface, callID uint32, params types.List[datastore_types.DataStoreGetMetaParam]) (*nex.RMCMessage, *nex.Error) {
	if commonProtocol.GetObjectInfoByPersistenceTargetWithPassword == nil {
		common_globals.Logger.Warning("GetObjectInfoByPersistenceTargetWithPassword not defined")
		return nil, nex.NewError(nex.ResultCodes.Core.NotImplemented, "change_error")
	}

	if commonProtocol.GetObjectInfoByDataIDWithPassword == nil {
		common_globals.Logger.Warning("GetObjectInfoByDataIDWithPassword not defined")
		return nil, nex.NewError(nex.ResultCodes.Core.NotImplemented, "change_error")
	}

	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, "change_error")
	}

	connection := packet.Sender()
	endpoint := connection.Endpoint()

	pMetaInfo := types.NewList[datastore_types.DataStoreMetaInfo]()
	pResults := types.NewList[types.QResult]()

	for _, param := range params {
		var objectInfo datastore_types.DataStoreMetaInfo
		var errCode *nex.Error

		// * Real server ignores PersistenceTarget if DataID is set
		if param.DataID == 0 {
			objectInfo, errCode = commonProtocol.GetObjectInfoByPersistenceTargetWithPassword(param.PersistenceTarget, param.AccessPassword)
			if errCode != nil && RemapPersistenceTarget != nil {
				if remapped, ok := RemapPersistenceTarget(uint64(connection.PID()), param.PersistenceTarget); ok {
					if info, remapErr := commonProtocol.GetObjectInfoByPersistenceTargetWithPassword(remapped, param.AccessPassword); remapErr == nil {
						common_globals.Logger.Infof("[MK8 GetMetasMulti] owner=%d persistence %d -> %d (viewed CTGP board)", uint64(param.PersistenceTarget.OwnerID), uint16(param.PersistenceTarget.PersistenceSlotID), uint16(remapped.PersistenceSlotID))
						objectInfo, errCode = info, nil
					}
				}
			}
		} else {
			objectInfo, errCode = commonProtocol.GetObjectInfoByDataIDWithPassword(param.DataID, param.AccessPassword)
		}

		common_globals.Logger.Infof("[MK8 GetMetasMulti] dataID=%d owner=%d persistence=%d resultOption=%d found=%v err=%v", uint64(param.DataID), uint64(param.PersistenceTarget.OwnerID), uint16(param.PersistenceTarget.PersistenceSlotID), uint8(param.ResultOption), errCode == nil, errCode)
		if errCode != nil {
			objectInfo = datastore_types.NewDataStoreMetaInfo()

			pResults = append(pResults, types.NewQResultError(errCode.ResultCode))
		} else {
			errCode = commonProtocol.VerifyObjectPermission(objectInfo.OwnerID, connection.PID(), objectInfo.Permission)
			if errCode != nil {
				objectInfo = datastore_types.NewDataStoreMetaInfo()

				pResults = append(pResults, types.NewQResultError(errCode.ResultCode))
			} else {
				pResults = append(pResults, types.NewQResultSuccess(nex.ResultCodes.DataStore.Unknown))
			}

			objectInfo.FilterPropertiesByResultOption(param.ResultOption)
		}

		pMetaInfo = append(pMetaInfo, objectInfo)
	}

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	pMetaInfo.WriteTo(rmcResponseStream)
	pResults.WriteTo(rmcResponseStream)

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = datastore.ProtocolID
	rmcResponse.MethodID = datastore.MethodGetMetasMultipleParam
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterGetMetasMultipleParam != nil {
		go commonProtocol.OnAfterGetMetasMultipleParam(packet, params)
	}

	return rmcResponse, nil
}
