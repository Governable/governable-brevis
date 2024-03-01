package age

import (
	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/brevis-network/brevis-sdk/test"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestCircuit(t *testing.T) {
	app, err := sdk.NewBrevisApp()
	check(err)

	contractAddress := common.HexToAddress("0xc00e94cb662c3520282e6f5717214004a7f26888")
	
	app.AddStorage(sdk.StorageData{
		BlockNum: big.NewInt(17800140),
		Address: contractAddress,
		Key: common.HexToHash("0xc2679997147cc711ecb6f1a090ddd97a89dfba7e3a04a3fb325563573f6fed21"),
	})

	guest := &AppCircuit{}
	guestAssignment := &AppCircuit{}

	circuitInput, err := app.BuildCircuitInput(guest)
	check(err)

	test.ProverSucceeded(t, guest, guestAssignment, circuitInput)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
