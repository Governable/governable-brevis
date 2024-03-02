package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/brevis-network/brevis-quickstart/age"
	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"path/filepath"
)

var mode = flag.String("mode", "", "compile or prove")
var outDir = flag.String("out", "$HOME/circuitOut/myBrevisApp", "compilation output dir")
var srsDir = flag.String("srs", "$HOME/kzgsrs", "where to cache kzg srs")
var slotNum = flag.String("slot", "", "slot to lookup")
var blockNum = flag.Int64("block", 0, "block number to lookup")
var contractAddress = common.HexToAddress("0xc944e90c64b2c07662a292be6244bdf05cda44a7")

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
		BlockNum: big.NewInt(19341099),
		Address: contractAddress,
		Key: common.HexToHash("0x55ccb1b16b10b19d498a335426da71059f3255a84a320fe81c2a761e2cc095d0"),
		Value: common.HexToHash("0x0000000000000000000000000000000000000000000000252248deb6e6940000"),
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

	if *blockNum == 0 {
		panic("-block is required")
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

	ec, err := ethclient.Dial("<your-eth-rpc>")
	slotValue, err := ec.StorageAt(context.Background(), contractAddress, common.HexToHash(*slotNum), big.NewInt(*blockNum))
	check(err)

	app.AddStorage(sdk.StorageData{
		BlockNum: big.NewInt(*blockNum), //17800140
		Address: contractAddress,
		Key: common.HexToHash(*slotNum),
		Value: common.BytesToHash(slotValue),
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
	appContract := common.HexToAddress("0xb0DA53679B6e7aB6c7c21e92B02abFd18BF627EA")
	refundee := common.HexToAddress("0xb0DA53679B6e7aB6c7c21e92B02abFd18BF627EA")

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
