# Governable Brevis

## Usage:
- Set the token contract on L1 as `contractAddress` in `main.go`.

- Set ETH RPCs
- Compile using `go run main.go -mode=compile -out="$HOME/circuitOut/myBrevisApp"`
- Run using `go run main.go -mode=prove -slot=0x55ccb1b16b10b19d498a335426da71059f3255a84a320fe81c2a761e2cc095d0 -block 19341099`
    - Replace your slot with the derived slot for the users balance.
    - Update the block number to match your L1checkpointBlock