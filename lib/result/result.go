package result

type Snapshot struct {
	Timestamp int    `json:"timestamp"`
	Amount0   string `json:"amount0"`
	Amount1   string `json:"amount1"`
	AmountUSD string `json:"amountUSD"`
	Price     string `json:"price"`
}
type Save struct {
	UpdateInterval int         `json:"update_interval"`
	StartAmount    string      `json:"start_amount"`
	StartTime      int         `json:"start_time"`
	EndTime        int         `json:"end_time"`
	Results        []RunResult `json:"results"`
}

type RunResult struct {
	ParameterA int `json:"parameterA"`
	// ParameterB int `json:"parameterB"`
	EndAmount      string `json:"end_amount"`
	VarianceHourly string `json:"variance_hourly"` // o^2
	VarianceDaily  string `json:"variance_daily"`
	//Snapshots      []Snapshot `json:"snapshots"`
}
