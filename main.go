package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	snapshotInterval := 60 * 60 // Should be 3600

	var wg sync.WaitGroup
	start := time.Now()

	//durations := []int{2, 6, 24, 7 * 24, 30 * 24, 100 * 24, 200 * 24}
	durations := []int{200 * 24}
	for j := 0; j < len(durations); j++ {
		durations[j] = durations[j] * 60 * 60
	}
	amountHistorySnapshots := 100
	mulUpperBound := IntPow(2, 16)
	results := make([]result.RunResult, len(durations)*(mulUpperBound-1))

	for durIndex, duration := range durations {
		// Addiitonal forloop to reduce memory usage
		mul := 1
		for {
			if mul == mulUpperBound {
				break
			}
			for j := 0; j < 1000; j, mul = j+1, mul+1 {
				if mul == mulUpperBound {
					break
				}
				// Prices Snapshot for moving average
				// Interval in which the snapshots are taken
				i := durIndex*mulUpperBound + (mul - 1)
				priceHistoryInterval := duration / amountHistorySnapshots
				strategy := strat.NewVolatilitySizedIntervalStrategy(startAmount0, startAmount1, pool, amountHistorySnapshots, mul)
				execution := executor.CreateExecution(strategy, startTime, updateInterval, snapshotInterval, priceHistoryInterval, transactions)
				wg.Add(1)
				go runAndAppend(&wg, execution, i, mul, duration, results)
			}
			wg.Wait()
		}
	}

	transLen := len(transactions)
	saveFile(results, filename, startAmount, updateInterval, transactions[0].Timestamp, transactions[transLen-1].Timestamp)

	t := time.Now()
	fmt.Println("Time: ", t.Sub(start))
	fmt.Println("Done")
}

//goland:noinspection SpellCheckingInspection
func runAndAppend(wg *sync.WaitGroup, execution *executor.Execution, i, mul, duration int, results []result.RunResult) {
	defer wg.Done()
	execution.Run()

	averageHourly := new(ui.Int)
	averageDaily := new(ui.Int)
	length := len(execution.AmountUSDSnapshots)
	lengthDaily := length / 24
	for i, amount := range execution.AmountUSDSnapshots {
		averageHourly.Add(averageHourly, amount)
		if i%24 == 0 {
			averageDaily.Add(averageDaily, amount)
		}
	}

	averageHourly.Div(averageHourly, ui.NewInt(uint64(length)))
	averageDaily.Div(averageDaily, ui.NewInt(uint64(lengthDaily)))
	varianceHourly := new(ui.Int)
	varianceDaily := new(ui.Int)
	for i := 0; i < length; i++ {
		diffHourly := new(ui.Int).Sub(execution.AmountUSDSnapshots[i], averageHourly)
		diffHourlySquared := new(ui.Int).Mul(diffHourly, diffHourly)
		varianceHourly.Add(varianceHourly, diffHourlySquared)
		if i%24 == 0 {
			diffDaily := new(ui.Int).Sub(execution.AmountUSDSnapshots[i], averageDaily)
			diffDailySquared := new(ui.Int).Mul(diffDaily, diffDaily)
			varianceDaily.Add(varianceDaily, diffDailySquared)
		}
	}
	// n-1 gives a unbiased estimator
	varianceHourly.Div(varianceHourly, ui.NewInt(uint64(length-1)))
	varianceDaily.Div(varianceDaily, ui.NewInt(uint64(lengthDaily-1)))

	res := createResult(execution, duration, mul, varianceHourly, varianceDaily)
	results[i] = res
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

	tosafe := result.Save{
		UpdateInterval: updateInterval,
		StartAmount:    startAmount,
		StartTime:      startTime,
		EndTime:        endTime,
		Results:        results,
	}
	err = encoder.Encode(tosafe)
	if err != nil {
		return
	}

}

func createResult(execution *executor.Execution, duration, mul int, varianceHourly, varianceDaily *ui.Int) result.RunResult {

	length := len(execution.AmountUSDSnapshots)

	amountUSDEnd := execution.AmountUSDSnapshots[length-1]
	amountEnd := amountUSDEnd.ToBig().String()

	r := result.RunResult{
		EndAmount:      amountEnd,
		HistoryWindow:  duration,
		MultiplierX8:   mul,
		VarianceHourly: varianceHourly.ToBig().String(),
		VarianceDaily:  varianceDaily.ToBig().String(),
	}

	return r

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
