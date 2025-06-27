package dto

import (
	"strings"
)

type Transaction struct {
	BlockNumber uint64
	Hash        string
	From        string
	To          string
	Value       string
	Contract    string
	Timestamp   uint64
}

func (tx *Transaction) ExtractTokenInfo(trace BlockTrace) bool {
	var params []string

	method := trace.Action.Input[:10]
	input := trace.Action.Input
	length := len(input)
	switch length {
	case 138:
		params = []string{
			input[10:74], 
			input[74:],
		}
	case 202:
		params = []string{
			input[10:74], 
			input[74:138], 
			input[138:],
		}
	default:
		return false
	}

	switch method {
	case "0xa9059cbb": // transfer
		tx.From = trace.Action.From
		tx.To = convertToAddress(params[0])
		tx.Value = params[1]
		tx.Contract = trace.Action.To

	case "0x23b872dd": // transferFrom
		tx.From = convertToAddress(params[0])
		tx.To = convertToAddress(params[1])
		tx.Value = params[2]
		tx.Contract = trace.Action.To

	case "0x6ea056a9": // sweep
		tx.From = trace.Action.To
		tx.To = trace.Action.From
		tx.Value = params[1]
		tx.Contract = convertToAddress(params[0])

	case "0x40c10f19": // mint
		tx.From = "0x0000000000000000000000000000000000000000"
		tx.To = convertToAddress(params[0])
		tx.Value = params[1]
		tx.Contract = tx.To
	default:
		return false
	}

	return true
}

func convertToAddress(str string) string {
	return "0x" + strings.ToLower(str[24:])
}