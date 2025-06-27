package dto

type BlockTrace struct {
	Action              TraceAction  `json:"action"`
	BlockHash           string       `json:"blockHash"`
	BlockNumber         uint64       `json:"blockNumber"`
	Result              *TraceResult `json:"result,omitempty"`
	Error               *string      `json:"error,omitempty"`
	Subtraces           int          `json:"subtraces"`
	TraceAddress        []int        `json:"traceAddress"`
	TransactionHash     string       `json:"transactionHash"`
	TransactionPosition int          `json:"transactionPosition"`
	Type                string       `json:"type"` // "call", "create", "suicide", "reward"
}

func (tx *BlockTrace) IsTokenTransfer() bool {
	input := tx.Action.Input
	if input == "0x" || input == "0x00" {
		return false
	}

	if len(input) < 10 {
		return false
	}

	switch input[:10] {
	case "0xa9059cbb":
		return true
	case "0x23b872dd":
		return true
	case "0x6ea056a9":
		return true
	case "0x40c10f19":
		return true
	default:
		return false
	}
}

type TraceAction struct {
	// common fields
	CallType string `json:"callType,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Value    string `json:"value,omitempty"`
	Input    string `json:"input,omitempty"`
	Gas      string `json:"gas,omitempty"`

	// Create-specific fields
	Init string `json:"init,omitempty"`

	// Suicide-specific fields
	Address       string `json:"address,omitempty"`
	RefundAddress string `json:"refundAddress,omitempty"`
	Balance       string `json:"balance,omitempty"`

	// Reward-specific fields
	Author     string `json:"author,omitempty"`
	RewardType string `json:"rewardType,omitempty"` // "block", "uncle"
}

type TraceResult struct {
	// Call result
	GasUsed string `json:"gasUsed,omitempty"`
	Output  string `json:"output,omitempty"`

	// Create result
	Address string `json:"address,omitempty"`
	Code    string `json:"code,omitempty"`
}