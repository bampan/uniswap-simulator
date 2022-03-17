package result

type Snapshot struct {
	Timestamp int    `json:"timestamp"`
	Amount0   string `json:"amount0"`
	Amount1   string `json:"amount1"`
	AmountUSD string `json:"amountUSD"`
	Price     string `json:"price"`
}
type Save struct {
	StartAmount string      `json:"start_amount"`
	StartTime   int         `json:"start_time"`
	EndTime     int         `json:"end_time"`
	Results     []RunResult `json:"results"`
}

type RunResult struct {
	UpdateInterval int `json:"update_interval"`
	//ParameterA     int `json:"parameterA"`
	//ParameterB     int `json:"parameterB"`
	MultiplierX10           int     `json:"multiplierX10"`
	HistoryWindow           int     `json:"history_window"`
	Return                  float64 `json:"return_on_investment"`
	EndAmount               string  `json:"end_amount"`
	MaxDrawdown             float64 `json:"max_draw_down"`
	StandardDeviationHourly float64 `json:"standard_deviation_hourly"`
	StandardDeviationDaily  float64 `json:"standard_deviation_daily"`
	StandardDeviationWeekly float64 `json:"standard_deviation_weekly"`
	DownwardDeviationHourly float64 `json:"downward_deviation_hourly"`
	DownwardDeviationDaily  float64 `json:"downward_deviation_daily"`
	DownwardDeviationWeekly float64 `json:"downward_deviation_weekly"`
	VaR95Hourly             float64 `json:"VaR95_hourly"`
	VaR95Daily              float64 `json:"VaR95_daily"`
	VaR95Weekly             float64 `json:"VaR95_weekly"`
}
