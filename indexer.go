package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/panjf2000/ants/v2"
)

/*
	9 block (2sec) > wait 18 sec > 9 block (2sec) > process 18 block (2,332,800 RPs)
	websocket subscribe 1sec new block (2,592,000 RPs)

	collect 18 blocks
	process them and check if re-org hapenned
		- re-org occur
			- pause any confirmation queues until we find the ancestor
			- check ancestor block if exist on 18 blocks array
			- if not exists check database for ancestor and use binary search
			- trigger an event to all queues with ancestor block number to check if each queue block should be dropped or not
				- dropped ? send an event to backend that block no longer valid via webhook
				- not ? continue if passed confirmation number send a confirmed block webhook
		- re-org does not occur
			- save to database
		- check if any transactions we have
			- found
				- check if block already exceed number of confirmation blocks ex. block 2/18 have a transaction
				- if not add it to a confirmation queue

	3 checks when processing blocks
		- re-org detected?
		- waiting blocks threshold confirmation
		- processing new blocks transactions

	1. read last block and wait until the difference between last block and current block on chain 18 block and start get each block
		- if db return nil a call to last block will occur (1 rpc)
		- 18 sec pass we may have extra rpc call to last_block

	2. read 9 blocks and wait 18 sec and then read 9 blocks and start processing 18 blocks (may require more rpc calls if no new blocks added on the chain)
	3. read 18 blocks and wait 1 sec between each block (may require more rpc calls if no new blocks added on the chain)

	run a job every 24h to clear all data for the previous day

	addresses > address - chain id
	tokens > smart_contract_address - chain id - decimals - symbol
	blocks > block_number - block_hash - block_parent_hash - chain id
	indexer_states > last_block_number - last_block_hash - chain id

	Indexer
		* track and collect blocks > 1 thread
		* re-org handling > 1 thread
		* blocks processing > 1 or n threads ?
		* webhooks events

	{
		"bsc" : { last_block_number: 12345678, last_block_hash: "0x1234567890abcdef", chain: "bsc" },
		"eth" : { last_block_number: 12345678, last_block_hash: "0x1234567890abcdef", chain: "eth" }
	}

	re-org steps:
		1. detect re-org parentHash != hash
		2.
		2. backtrack chain until we find the case where parentHash == hash and while backtracking put all orphaned blocks (new) on array from the chain not the database
			1. check blocks array
			2. second check database if we have that hash (we should at least have the ancestor on the database if we save blocks stored)
		3. broadcast

		func (tx *Transaction) IsTokenTransfer() bool {

	if tx.Input == "0x" || tx.Input == "0x00" {
		return false
	}

	if len(tx.Input) < 10 {
		return false
	}

	switch tx.Input[:10] {
	case "0xa9059cbb": transfer
		return true
	case "0x23b872dd": transferFrom
		return true
	case "0x6ea056a9": swap
		return true
	case "0x40c10f19": mint(address,uint256)
		return true
	default:
		return false
	}

	BNB
	ETH
	Base
	POLYGON
	ARBITRUM
	AVALANCHE
	Optimisim

	Solana
	Bitcoin
	Tron

	Dogecoin
	Litecoin
}
*/

type chain struct {
	id string
	name string
}


func main() {
	// rpcClient, err := rpc.Dial("https://base-mainnet.g.alchemy.com/v2/1MTFMyOaGNUQQco1Urk9oWVHqPZ27ddw")
	// if err != nil {
	// 	panic(fmt.Errorf("failed to connect to RPC: %v", err))
	// }
	// defer rpcClient.Close()

	// client, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
	// if err != nil {
	// 	panic(fmt.Errorf("failed to connect to Ethereum client: %v", err))
	// }
	// defer client.Close()
	// // b, err := client.BlockNumber(context.Background())
	// // if err != nil {
	// // 	panic(fmt.Errorf("failed to get block: %v", err))
	// // }
	// blockNumber := new(big.Int).SetUint64(22746691)
	// logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
	// 	FromBlock: blockNumber,
	// 	ToBlock: blockNumber,
	// 	Addresses: []common.Address{},
	// 	Topics: [][]common.Hash{
	// 		{
	// 			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// writeLogsToFile(logs, blockNumber, "get_logs")
	pool, _ := ants.NewPool(2)
	resultCh := make(chan int, 5)
	var wg sync.WaitGroup
	for i:=0; i<5; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()
			fmt.Println("Running goroutine number ", i+1)
			time.Sleep(4*time.Second)
			resultCh <- i+1
		})
		fmt.Println("Executing next goroutine")
		// if err != nil {
		// 	panic(err)
		// }
		// time.Sleep(1*time.Second)
	}
	wg.Wait()
	close(resultCh)
	for res := range resultCh {
		fmt.Println(res)
	}
	// var blockTraces []dto.BlockTrace
	// err = client.Client().CallContext(context.Background(), &blockTraces, "trace_block", blockNumber)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(blockTraces[0].Action.To)
	
	// var result []polygon.BlockTrace
	// if err := json.Unmarshal([]byte(raw), &result); err != nil {
	// 	panic(err)
	// }
	// for _, tx := range result {
	// 	if tx.Error != "" || tx.Type != "call" || tx.Action.Value == "" {
	// 		continue
	// 	}
	// 	value := new(big.Int).SetUint64(00045577)
	// 	if value.Sign() > 0 {
	// 		fmt.Println(tx.Action.Value)
	// 		break
	// 	}
	// }
	// b, err := client.BlockNumber(context.Background())
	// if err != nil {
	// 	panic(fmt.Errorf("failed to get block: %v", err))
	// }
	// blockNumber := new(big.Int).SetUint64(b)

	// var traces []interface{}
	// err = rpcClient.CallContext(context.Background(), &traces, "trace_block", blockNumber)
	// if err != nil {
	// 	panic(fmt.Errorf("failed to get traces: %v", err))
	// }

	// // 1. Write all traces to file
	// writeTracesToFile(traces, blockNumber, "all")

	// // 2. Filter traces with callType == "call"
	// callTraces := filterTracesBySpecificCallType(traces, "call")
	// writeTracesToFile(callTraces, blockNumber, "call_only")

	// // 3. Filter traces with callType == "call" AND value > 0
	// callWithValueTraces := filterCallTracesWithValue(callTraces)
	// writeTracesToFile(callWithValueTraces, blockNumber, "call_with_value")

	// fmt.Printf("Block %d stats:\n", blockNumber)
	// // fmt.Printf("Block %d transactions:\n", len(b.Transactions()))
	// fmt.Printf("- Total traces: %d\n", len(traces))
	// fmt.Printf("- Call traces: %d\n", len(callTraces))
	// fmt.Printf("- Call traces with value > 0: %d\n", len(callWithValueTraces))
	
}

// func writeTracesToFile(traces []interface{}, blockNumber *big.Int, suffix string) {
// 	jsonData, err := json.MarshalIndent(traces, "", "  ")
// 	if err != nil {
// 		panic(fmt.Errorf("failed to marshal JSON: %v", err))
// 	}

// 	filename := fmt.Sprintf("traces_block_%d_%s.json", blockNumber, suffix)
// 	err = os.WriteFile(filename, jsonData, 0644)
// 	if err != nil {
// 		panic(fmt.Errorf("failed to write file: %v", err))
// 	}
// }

func writeLogsToFile(logs []types.Log, blockNumber *big.Int, suffix string) {
	jsonData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal JSON: %v", err))
	}

	filename := fmt.Sprintf("logs_block_%d_%s.json", blockNumber, suffix)
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write file: %v", err))
	}
}

// func filterTracesBySpecificCallType(traces []interface{}, callType string) []interface{} {
// 	var filtered []interface{}

// 	for _, trace := range traces {
// 		traceMap, ok := trace.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		action, ok := traceMap["action"].(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		if ct, exists := action["callType"]; exists {
// 			if ctStr, ok := ct.(string); ok && ctStr == callType {
// 				filtered = append(filtered, trace)
// 			}
// 		}
// 	}

// 	return filtered
// }

// func filterCallTracesWithValue(traces []interface{}) []interface{} {
// 	var filtered []interface{}
// 	zero := big.NewInt(0)

// 	for _, trace := range traces {
// 		traceMap, ok := trace.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		action, ok := traceMap["action"].(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		// Check if value exists and is > 0
// 		if val, exists := action["value"]; exists {
// 			valueStr, ok := val.(string)
// 			if !ok {
// 				continue
// 			}

// 			value := new(big.Int)
// 			value, ok = value.SetString(valueStr[2:], 16) // Remove 0x prefix and parse as hex
// 			if !ok {
// 				continue
// 			}

// 			if value.Cmp(zero) > 0 {
// 				filtered = append(filtered, trace)
// 			}
// 		}
// 	}

// 	return filtered
// }

// func main() {
// 	rpcClient, err := rpc.Dial("https://fullnode.avalanche.shkeeper.io:9960/ext/bc/C/rpc")
// 	if err != nil {
// 		panic(fmt.Errorf("failed to connect to RPC: %v", err))
// 	}
// 	defer rpcClient.Close()

// 	client, err := ethclient.Dial("https://fullnode.avalanche.shkeeper.io:9960/ext/bc/C/rpc")
// 	if err != nil {
// 		panic(fmt.Errorf("failed to connect to Ethereum client: %v", err))
// 	}
// 	defer client.Close()

// 	b, err := client.BlockByNumber(context.Background(), nil)
// 	if err != nil {
// 		panic(fmt.Errorf("failed to get block: %v", err))
// 	}

// 	// Trace config - customize as needed
// 	traceConfig := map[string]interface{}{
// 		"tracer": "callTracer",
// 		"timeout": "10s",
// 	}

// 	var traces []interface{}
// 	err = rpcClient.CallContext(context.Background(), &traces, "debug_traceBlockByNumber", b.Number(), traceConfig)
// 	if err != nil {
// 		panic(fmt.Errorf("failed to get traces: %v", err))
// 	}

// 	// 1. Write all traces to file
// 	writeTracesToFile(traces, b.Number(), "all")

// 	// 2. Filter traces with callType == "call"
// 	callTraces := filterTracesBySpecificCallType(traces, "CALL") // Note uppercase CALL for debug_traceBlockByNumber
// 	writeTracesToFile(callTraces, b.Number(), "call_only")

// 	// 3. Filter traces with callType == "call" AND value > 0
// 	callWithValueTraces := filterCallTracesWithValue(callTraces)
// 	writeTracesToFile(callWithValueTraces, b.Number(), "call_with_value")

// 	fmt.Printf("Block %d stats:\n", b.Number())
// 	fmt.Printf("- Total traces: %d\n", len(traces))
// 	fmt.Printf("- Call traces: %d\n", len(callTraces))
// 	fmt.Printf("- Call traces with value > 0: %d\n", len(callWithValueTraces))
// }

// func writeTracesToFile(traces []interface{}, blockNumber *big.Int, suffix string) {
// 	jsonData, err := json.MarshalIndent(traces, "", "  ")
// 	if err != nil {
// 		panic(fmt.Errorf("failed to marshal JSON: %v", err))
// 	}

// 	filename := fmt.Sprintf("traces_block_%d_%s.json", blockNumber, suffix)
// 	err = os.WriteFile(filename, jsonData, 0644)
// 	if err != nil {
// 		panic(fmt.Errorf("failed to write file: %v", err))
// 	}
// }

// func filterTracesBySpecificCallType(traces []interface{}, callType string) []interface{} {
// 	var filtered []interface{}

// 	for _, trace := range traces {
// 		traceMap, ok := trace.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		if ct, exists := traceMap["type"]; exists { // Note "type" instead of "callType" for debug_traceBlockByNumber
// 			if ctStr, ok := ct.(string); ok && ctStr == callType {
// 				filtered = append(filtered, trace)
// 			}
// 		}
// 	}

// 	return filtered
// }

// func filterCallTracesWithValue(traces []interface{}) []interface{} {
// 	var filtered []interface{}
// 	zero := big.NewInt(0)

// 	for _, trace := range traces {
// 		traceMap, ok := trace.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		// Check if value exists and is > 0
// 		if val, exists := traceMap["value"]; exists { // Directly in trace for debug_traceBlockByNumber
// 			valueStr, ok := val.(string)
// 			if !ok {
// 				continue
// 			}

// 			value := new(big.Int)
// 			value, ok = value.SetString(valueStr[2:], 16) // Remove 0x prefix and parse as hex
// 			if !ok {
// 				continue
// 			}

// 			if value.Cmp(zero) > 0 {
// 				filtered = append(filtered, trace)
// 			}
// 		}
// 	}

// 	return filtered
// }

// func TrackAndBatchBlocks(ctx context.Context, client * ethclient.Client) {
// 	for {
		
// 		// Sleep for a while before checking again
// 		select {
// 			case <-ctx.Done():
// 				fmt.Println("Stopping Indexer")
// 				return
// 			default:
// 				last_block_number, err := client.BlockNumber(ctx)
// 				if err != nil {
// 					panic(fmt.Errorf("failed to get last block number: %v", err.Error()))
// 				}
				
// 				if latest_block_number == 0 {
// 					latest_block_number = last_block_number
// 				} else {
// 					threshold := last_block_number - latest_block_number
// 					fmt.Printf("Latest Block: %d, Last Block: %d, Threshold: %d\n", latest_block_number, last_block_number, threshold)
// 					if threshold >= 18 { // process 18 blocks
// 						fmt.Printf("Processing blocks from %d to %d\n", latest_block_number, last_block_number)
// 						// Here you would fetch and process the blocks
// 						latest_block_number = last_block_number
// 					}
// 				}
				
// 				time.Sleep(10 * time.Second) // Adjust the sleep duration as needed
// 		}
// 	}
// }