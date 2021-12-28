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
	updateInterval := *flag.Int("n", 2, "updateInterval in hours")
	updateInterval = updateInterval * 60 * 60
	filename := *flag.String("file", "2_hours.json", "filename")
	flag.Parse()

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

	step := 10
	upperA := 40000
	lenA := upperA / step
	results := make([]result.RunResult, lenA)

	for a := step; a <= upperA; a += step {
		i := a/step - 1
		strategy := strat.NewConstantIntervalStrategy(startAmount0, startAmount1, pool, a)
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

func runAndAppend(wg *sync.WaitGroup, excecution *executor.Execution, a, i int, results []result.RunResult) {
	defer wg.Done()
	excecution.Run()

	average := new(ui.Int)
	for _, amount := range excecution.AmountUSDSnapshots {
		average.Add(average, amount)
	}
	length := len(excecution.AmountUSDSnapshots)
	average.Div(average, ui.NewInt(uint64(length)))
	varianceHourly := new(ui.Int)
	varianceDaily := new(ui.Int)
	for i := 0; i < length; i++ {
		diff := new(ui.Int).Sub(excecution.AmountUSDSnapshots[i], average)
		diffSquared := new(ui.Int).Mul(diff, diff)
		varianceHourly.Add(varianceHourly, diffSquared)
		if i%24 == 0 {
			varianceDaily.Add(varianceDaily, diffSquared)
		}
	}
	varianceHourly.Div(varianceHourly, ui.NewInt(uint64(length-1)))
	lengthDaily := length / 24
	varianceDaily.Div(varianceDaily, ui.NewInt(uint64(lengthDaily-1)))

	res := createResult(excecution, a, varianceHourly, varianceDaily)
	results[i] = res
}

func saveFile(results []result.RunResult, filename, startAmount string, updateInterval, startTime, endTime int) {

	filepath := path.Join("results", filename)
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

func createResult(execution *executor.Execution, a int, varianceHourly, varianceDaily *ui.Int) result.RunResult {

	length := len(execution.AmountUSDSnapshots)

	amountUSDEnd := execution.AmountUSDSnapshots[length-1]
	amountEnd := amountUSDEnd.ToBig().String()

	r := result.RunResult{
		EndAmount:      amountEnd,
		ParameterA:     a,
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
