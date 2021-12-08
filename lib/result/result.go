package result

type Snapshot struct {
	Timestamp int    `json:"timestamp"`
	Amount0   string `json:"amount0"`
	Amount1   string `json:"amount1"`
	AmountUSD string `json:"amountUSD"`
	Price     string `json:"Price"`
}
type Result struct {
	StartTime      int        `json:"start_time"`
	EndTime        int        `json:"end_time"`
	UpdateInterval int        `json:"update_interval"`
	AmountStart    string     `json:"amount_start"`
	AmountEnd      string     `json:"amount_end"`
	Snapshots      []Snapshot `json:"snapshots"`
}
