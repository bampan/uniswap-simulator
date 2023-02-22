package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ftchann/uniswap-simulator/lib/executor"
	ppool "github.com/ftchann/uniswap-simulator/lib/pool"
	"github.com/ftchann/uniswap-simulator/lib/result"
	strat "github.com/ftchann/uniswap-simulator/lib/strategy"
	ent "github.com/ftchann/uniswap-simulator/lib/transaction"

	ui "github.com/holiman/uint256"
)

// Rc Aave 6 month average APY
// may - november
// 2.94, 1.57, 2.82, 0.33, 2.45, 1.06, 7.18
// Average is 3.93
// https://loanscan.io/borrow/historical?interval=1y
var Rc = 0.0262
var startAmount string

func main() {
	// Parse flags
	updateIntervalPtr := flag.Int("n", 24, "updateInterval in hours")
	filenamePtr := flag.String("file", "1_day.json", "filename")
	flag.Parse()
	filename := *filenamePtr
	updateInterval := *updateIntervalPtr
	// Log flags
	fmt.Println("updateInterval in hours:", updateInterval)
	fmt.Println("filename:", filename)

	updateInterval = updateInterval * 60 * 60
	transactions := getTransactions()
	fmt.Println("Amount of Transactions: ", len(transactions))
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96big, _ := new(big.Int).SetString("1350174849792634181862360983626536", 10)
	sqrtX96, _ := ui.FromBig(sqrtX96big)

	pool := ppool.NewPool(token0, token1, fee, sqrtX96)

	startAmount0 := ui.NewInt(1_000_000) // 1 USDC
	// From the Price One month in
	startAmount1big, _ := new(big.Int).SetString("366874042000000", 10) // 366874042000000 wei ~= 1 USD worth of ETH
	// out_of_sample
	//startAmount1big, _ = new(big.Int).SetString("231104725000000", 10) // 231103715000000 wei ~= 1 USD worth of ETH
	startAmount1, _ := ui.FromBig(startAmount1big)
	startAmount = "2000000" // HardCoded is the easy way to do it

	startTime := transactions[0].Timestamp + 60*60*24*30
	// For out of Sample
	//startTime = 1637411195

	fmt.Println("Start Time: ", startTime)
	snapshotInterval := 60 * 60 // hourly

	var wg sync.WaitGroup
	start := time.Now()

	amountHistorySnapshots := 100
	mulUpperBound := IntPow(2, 16)
	results := make([]result.RunResult, mulUpperBound-1)
	duration := 24 * 60 * 60
	mul := 1
	for mul < mulUpperBound {
		for j := 0; j < 5000 && mul < mulUpperBound; j, mul = j+1, mul+1 {
			i := mul - 1
			priceHistoryInterval := duration / amountHistorySnapshots
			strategy := strat.NewBollingerBandsFillUpStrategy(startAmount0, startAmount1, pool, amountHistorySnapshots, mul)
			execution := executor.CreateExecution(strategy, startTime, 90000000000000, updateInterval, snapshotInterval, priceHistoryInterval, transactions)
			wg.Add(1)
			go runAndAppend(&wg, execution, i, mul, duration, updateInterval, results)
		}
		wg.Wait()
	}

	transLen := len(transactions)
	saveFile(results, filename, transactions[0].Timestamp, transactions[transLen-1].Timestamp)

	t := time.Now()
	fmt.Println("Time: ", t.Sub(start))
	fmt.Println("Done")
}
func getReturns(prices []*ui.Int) []float64 {
	returns := make([]float64, 0, len(prices))
	for i := 1; i < len(prices); i++ {
		sIMinus1 := prices[i-1].ToBig()
		sI := prices[i].ToBig()
		sIMinus1F, _ := new(big.Float).SetInt(sIMinus1).Float64()
		sIF, _ := new(big.Float).SetInt(sI).Float64()
		ret := math.Log(sIF / sIMinus1F)
		returns = append(returns, ret)
	}
	return returns
}

func calculateMaximumDrawdown(prices []*ui.Int) float64 {
	if len(prices) <= 1 {
		return 0
	}
	pricesFloat := make([]float64, 0, len(prices))
	for _, price := range prices {
		floatNumber, _ := new(big.Float).SetInt(price.ToBig()).Float64()
		pricesFloat = append(pricesFloat, floatNumber)
	}
	maxPrice := pricesFloat[0]
	maxDrawdown := 0.0
	for _, price := range pricesFloat {
		maxPrice = math.Max(maxPrice, price)
		drawdown := (price - maxPrice) / maxPrice
		maxDrawdown = math.Min(maxDrawdown, drawdown)
	}
	return maxDrawdown
}

// MAR is Zero
func calculateDownDeviation(prices []*ui.Int) float64 {
	if len(prices) <= 2 {
		return 0
	}
	returns := getReturns(prices)
	sum := 0.0
	for _, r := range returns {
		if r < 0 {
			sum += r * r
		}
	}
	n := len(returns)
	sum = sum / float64(n)
	downwardDeviation := math.Sqrt(sum)
	return downwardDeviation

}

// VaR 95%
func calculateVar(prices []*ui.Int) float64 {
	if len(prices) <= 2 {
		return 0
	}
	returns := getReturns(prices)
	sort.Float64s(returns)
	idx := len(returns) / 20
	var95 := returns[idx]
	return var95
}

func calculateStd(prices []*ui.Int) float64 {
	if len(prices) <= 2 {
		return 0
	}
	returns := getReturns(prices)
	n := len(returns)
	sum1 := 0.0
	for _, r := range returns {
		sum1 += r * r
	}
	sum1 = sum1 / float64(n-1)
	sum2 := 0.0
	for _, r := range returns {
		sum2 += r
	}
	sum2 = sum2 * sum2
	quotient := n * (n - 1)
	sum2 = sum2 / float64(quotient)

	variance := sum1 - sum2
	standardDeviation := math.Sqrt(variance)

	return standardDeviation
}

//goland:noinspection SpellCheckingInspection
func runAndAppend(wg *sync.WaitGroup, execution *executor.Execution, i, mul, duration, updateInterval int, results []result.RunResult) {
	defer wg.Done()
	execution.Run()

	pricesHourly := make([]*ui.Int, 0, len(execution.AmountUSDSnapshots))
	pricesDaily := make([]*ui.Int, 0, len(execution.AmountUSDSnapshots)/24)
	pricesWeekly := make([]*ui.Int, 0, len(execution.AmountUSDSnapshots)/24/7)
	for i := 0; i < len(execution.AmountUSDSnapshots); i++ {
		pricesHourly = append(pricesHourly, execution.AmountUSDSnapshots[i].Clone())
		if i%24 == 0 {
			pricesDaily = append(pricesDaily, execution.AmountUSDSnapshots[i].Clone())
		}
		if i%(24*7) == 0 {
			pricesWeekly = append(pricesWeekly, execution.AmountUSDSnapshots[i].Clone())
		}
	}

	maxDrawDown := calculateMaximumDrawdown(pricesHourly)

	stdHourly := calculateStd(pricesHourly)
	stdDaily := calculateStd(pricesDaily)
	stdWeekly := calculateStd(pricesWeekly)
	dDHourly := calculateDownDeviation(pricesHourly)
	dDDaily := calculateDownDeviation(pricesDaily)
	dDWeekly := calculateDownDeviation(pricesWeekly)
	varHourly := calculateVar(pricesHourly)
	varDaily := calculateVar(pricesDaily)
	varWeekly := calculateVar(pricesWeekly)
	res := createResult(execution, mul, duration, updateInterval, stdHourly, stdDaily, stdWeekly, dDHourly, dDDaily, dDWeekly, maxDrawDown, varHourly, varDaily, varWeekly)
	results[i] = res
}

func createResult(execution *executor.Execution, mul, duration, updateInterval int, stdHourly, stdDaily, stdWeekly, dDHourly, dDDaily, dDWeekly, maxDrawDown, var95Hourly, var95Daily, var95Weekly float64) result.RunResult {

	length := len(execution.AmountUSDSnapshots)

	amountUSDEnd := execution.AmountUSDSnapshots[length-1]
	amountEnd := amountUSDEnd.ToBig().String()

	amountEndFloat, _ := new(big.Float).SetInt(amountUSDEnd.ToBig()).Float64()
	amountStartFloat, _ := new(big.Float).SetInt(execution.AmountUSDSnapshots[0].ToBig()).Float64()
	amountDiff := amountEndFloat - amountStartFloat
	roi := amountDiff / amountStartFloat

	r := result.RunResult{
		EndAmount:     amountEnd,
		MultiplierX10: mul,
		//ParameterA: a,
		HistoryWindow:           duration,
		UpdateInterval:          updateInterval,
		Return:                  roi,
		StandardDeviationHourly: stdHourly,
		StandardDeviationDaily:  stdDaily,
		StandardDeviationWeekly: stdWeekly,
		DownwardDeviationHourly: dDHourly,
		DownwardDeviationDaily:  dDDaily,
		DownwardDeviationWeekly: dDWeekly,
		MaxDrawdown:             maxDrawDown,
		VaR95Hourly:             var95Hourly,
		VaR95Daily:              var95Daily,
		VaR95Weekly:             var95Weekly,
	}

	return r

}

func saveFile(results []result.RunResult, filename string, startTime, endTime int) {

	filepath := path.Join("results", filename)
	err := os.Mkdir("results", os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory: ", err)
	}
	fmt.Println("Saving to: ", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	toSave := result.Save{
		StartAmount: startAmount,
		StartTime:   startTime,
		EndTime:     endTime,
		Results:     results,
	}
	err = encoder.Encode(toSave)
	if err != nil {
		return
	}

}

func getTransactions() []ent.Transaction {
	filename := "transactions_insample.json"
	filepath := path.Join("data", filename)
	file, err := os.Open(filepath)
	check(err)
	value, err := ioutil.ReadAll(file)
	check(err)
	var transactionsInput []ent.TransactionInput
	err = json.Unmarshal([]byte(value), &transactionsInput)
	check(err)
	var transactions []ent.Transaction
	for _, transIn := range transactionsInput {
		useX96, _ := strconv.ParseBool(transIn.UseX96)
		trans := ent.Transaction{
			transIn.Type,
			stringToUint256(transIn.Amount),
			stringToUint256(transIn.Amount0),
			stringToUint256(transIn.Amount1),
			transIn.ID,
			stringToUint256(transIn.SqrtPriceX96),
			transIn.Tick,
			transIn.TickLower,
			transIn.TickUpper,
			transIn.Timestamp,
			useX96,
		}
		transactions = append(transactions, trans)
	}
	return transactions
}

func stringToUint256(amount string) *ui.Int {
	bigint := new(big.Int)
	bigint.SetString(amount, 10)
	uint256, _ := ui.FromBig(bigint)
	return uint256
}

// IntPow calculates n to the mth power. Since the result is an int, it is assumed that m is a positive power
func IntPow(n, m int) int {
	// IDK why go doesn't have it...
	if m == 0 {
		return 1
	}
	result := n
	for i := 2; i <= m; i++ {
		result *= n
	}
	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
