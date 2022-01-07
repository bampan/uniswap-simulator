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
	"strconv"
	"sync"
	"time"
	"uniswap-simulator/lib/executor"
	ppool "uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/result"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

func main() {
	// Parse flags
	updateIntervalPtr := flag.Int("n", 2, "updateInterval in hours")
	filenamePtr := flag.String("file", "2_hours.json", "filename")
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
	startAmount1, _ := ui.FromBig(startAmount1big)
	startAmount := "2000000" // HardCoded is the easy way to do it

	startTime := transactions[0].Timestamp + 60*60*24*30
	snapshotInterval := 60 * 60 // daily

	var wg sync.WaitGroup
	start := time.Now()

	step := 10
	upperA := 10
	lenA := upperA / step
	results := make([]result.RunResult, lenA)

	for a := step; a <= upperA; a += step {
		i := a/step - 1
		strategy := strat.NewIntervalAroundPriceStrategy(startAmount0, startAmount1, pool, a)
		execution := executor.CreateExecution(strategy, startTime, updateInterval, snapshotInterval, transactions)
		wg.Add(1)
		go runAndAppend(&wg, execution, a, i, results)
	}
	wg.Wait()

	transLen := len(transactions)
	saveFile(results, filename, startAmount, updateInterval, transactions[0].Timestamp, transactions[transLen-1].Timestamp)

	t := time.Now()
	fmt.Println("Time: ", t.Sub(start))
	fmt.Println("Done")
}

func calculateStd(prices []*ui.Int) float64 {
	returns := make([]float64, 0, len(prices))
	for i := 1; i < len(prices); i++ {
		S1 := prices[i-1].ToBig()
		S2 := prices[i].ToBig()
		S1F, _ := new(big.Float).SetInt(S1).Float64()
		S2F, _ := new(big.Float).SetInt(S2).Float64()
		ret := math.Log(S2F / S1F)
		returns = append(returns, ret)
	}
	average := 0.0
	length := len(returns)
	for _, amount := range returns {
		average += amount
	}
	average = average / float64(length)
	variance := 0.0
	for _, amount := range returns {
		diff := amount - average
		squared := diff * diff
		variance += squared
	}
	// Erwartungstreuer Schaetzer
	variance = variance / float64(length-1)
	// Standardabweichung
	standardDeviation := math.Sqrt(variance)
	return standardDeviation
}

//goland:noinspection SpellCheckingInspection
func runAndAppend(wg *sync.WaitGroup, execution *executor.Execution, a, i int, results []result.RunResult) {
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

	stdHourly := calculateStd(pricesHourly)
	stdDaily := calculateStd(pricesDaily)
	stdWeekly := calculateStd(pricesWeekly)
	res := createResult(execution, a, stdHourly, stdDaily, stdWeekly)
	results[i] = res
}

func createResult(execution *executor.Execution, a int, stdHourly, stdDaily, stdWeekly float64) result.RunResult {

	length := len(execution.AmountUSDSnapshots)

	amountUSDEnd := execution.AmountUSDSnapshots[length-1]
	amountEnd := amountUSDEnd.ToBig().String()

	r := result.RunResult{
		EndAmount:               amountEnd,
		ParameterA:              a,
		StandardDeviationHourly: stdHourly,
		StandardDeviationDaily:  stdDaily,
		StandardDeviationWeekly: stdWeekly,
	}

	return r

}

func saveFile(results []result.RunResult, filename, startAmount string, updateInterval, startTime, endTime int) {

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
		UpdateInterval: updateInterval,
		StartAmount:    startAmount,
		StartTime:      startTime,
		EndTime:        endTime,
		Results:        results,
	}
	err = encoder.Encode(toSave)
	if err != nil {
		return
	}

}

func getTransactions() []ent.Transaction {
	filename := "transactions.json"
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
