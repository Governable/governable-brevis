package age

import (
	"github.com/brevis-network/brevis-sdk/sdk"
)

type AppCircuit struct{}

func (c *AppCircuit) Allocate() (maxReceipts, maxStorage, maxTransactions int) {
	// Our app is only ever going to use one storage data at a time so
	// we can simply limit the max number of data for storage to 1 and
	// 0 for all others
	return 0, 1, 0
}

func (c *AppCircuit) Define(api *sdk.CircuitAPI, in sdk.DataInput) error {
	slots := sdk.NewDataStream(api, in.StorageSlots)
	slot := slots.Get(0)

	// Output variables can be later accessed in our app contract
	api.OutputUint(64, slot.BlockNum)
	api.OutputAddress(slot.Contract)
	api.OutputBytes32(slot.Key)
	api.OutputBytes32(slot.Value)

	return nil
}
