package main

import (
	"flag"
	"fmt"
	"github.com/brevis-network/brevis-quickstart/age"
	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"path/filepath"
)

var mode = flag.String("mode", "", "compile or prove")
var outDir = flag.String("out", "$HOME/circuitOut/myBrevisApp", "compilation output dir")
var srsDir = flag.String("srs", "$HOME/kzgsrs", "where to cache kzg srs")
var slotNum = flag.String("slot", "", "slot to lookup")
var contractAddress = common.HexToAddress("0xc00e94cb662c3520282e6f5717214004a7f26888")

func main() {
	flag.Parse()
	switch *mode {
	case "compile":
		compile()
	case "prove":
		prove()
	default:
		panic(fmt.Errorf("unsupported mode %s", *mode))
	}
}

func compile() {
	// This first part is copied from app/circuit_test.go. We added the source data, then we generated the circuit input.
	app, err := sdk.NewBrevisApp()
	check(err)
	
	app.AddStorage(sdk.StorageData{
		BlockNum: big.NewInt(17800140),
		Address: contractAddress,
		Key: common.HexToHash("0xc2679997147cc711ecb6f1a090ddd97a89dfba7e3a04a3fb325563573f6fed21"),
		Value: common.HexToHash("0x000000000000000000000000000000000000000000000001397f97d8b255ec3a"),
	})
	appCircuit := &age.AppCircuit{}

	circuitInput, err := app.BuildCircuitInput(appCircuit)
	check(err)

	// The compilation output is the description of the circuit's constraint system.
	// You should use sdk.WriteTo to serialize and save your circuit so that it can
	// be used in the proving step later.
	compiledCircuit, err := sdk.Compile(appCircuit, circuitInput)
	check(err)
	err = sdk.WriteTo(compiledCircuit, filepath.Join(*outDir, "compiledCircuit"))
	check(err)

	// Setup is a one-time effort per circuit. A cache dir can be provided to output
	// external dependencies. Once you have the verifying key you should also save
	// its hash in your contract so that when a proof via Brevis is submitted
	// on-chain you can verify that Brevis indeed used your verifying key to verify
	// your circuit computations
	pk, vk, err := sdk.Setup(compiledCircuit, *srsDir)
	check(err)
	err = sdk.WriteTo(pk, filepath.Join(*outDir, "pk"))
	check(err)
	err = sdk.WriteTo(vk, filepath.Join(*outDir, "vk"))
	check(err)

	fmt.Println("compilation/setup complete")
}

func prove() {
	if len(*slotNum) == 0 {
		panic("-slot is required")
	}

	// Loading the previous compile result into memory
	fmt.Println(">> Reading circuit, pk, and vk from disk")
	compiledCircuit, err := sdk.ReadCircuitFrom(filepath.Join(*outDir, "compiledCircuit"))
	check(err)
	pk, err := sdk.ReadPkFrom(filepath.Join(*outDir, "pk"))
	check(err)
	vk, err := sdk.ReadVkFrom(filepath.Join(*outDir, "vk"))
	check(err)

	// Query the user specified tx
	app, err := sdk.NewBrevisApp()
	check(err)

	fmt.Println(common.HexToHash(*slotNum))

	app.AddStorage(sdk.StorageData{
		BlockNum: big.NewInt(17800140), //17800140
		Address: contractAddress,
		Key: common.HexToHash(*slotNum),
		Value: common.HexToHash("0x000000000000000000000000000000000000000000000001397f97d8b255ec3a"),
	})

	appCircuit := &age.AppCircuit{}
	appCircuitAssignment := &age.AppCircuit{}

	// Prove
	fmt.Println(">> Proving the transaction using my circuit")
	circuitInput, err := app.BuildCircuitInput(appCircuit)
	check(err)
	witness, publicWitness, err := sdk.NewFullWitness(appCircuitAssignment, circuitInput)
	check(err)
	proof, err := sdk.Prove(compiledCircuit, pk, witness)
	check(err)
	err = sdk.WriteTo(proof, filepath.Join(*outDir, "proof-"+*slotNum))
	check(err)

	// Test verifying the proof we just generated
	err = sdk.Verify(vk, publicWitness, proof)
	check(err)

	fmt.Println(">> Initiating Brevis request")
	appContract := common.HexToAddress("0x73090023b8D731c4e87B3Ce9Ac4A9F4837b4C1bd")
	refundee := common.HexToAddress("0x164Ef8f77e1C88Fb2C724D3755488bE4a3ba4342")

	calldata, requestId, feeValue, err := app.PrepareRequest(vk, 1, 11155111, refundee, appContract)
	check(err)
	fmt.Printf("calldata %x\n", calldata)
	fmt.Printf("feeValue %d\n", feeValue)
	fmt.Printf("requestId %s\n", requestId)

	// Submit proof to Brevis
	fmt.Println(">> Submitting my proof to Brevis")
	err = app.SubmitProof(proof)
	check(err)

	// // Poll Brevis gateway for query status till the final proof is submitted
	// // on-chain by Brevis and your contract is called
	// fmt.Println(">> Waiting for final proof generation and submission")
	// submitTx, err := app.WaitFinalProofSubmitted(context.Background())
	// check(err)
	// fmt.Printf(">> Final proof submitted: tx hash %s\n", submitTx)

	// // [Don't forget to make the transaction that pays the fee by calling Brevis.sendRequest]
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
