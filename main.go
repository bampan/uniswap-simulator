package main

import (
	"encoding/json"
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
	sqrtmath "uniswap-simulator/lib/sqrtprice_math"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

func main() {
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

	startTime := transactions[0].Timestamp + 60*60*24*30
	updateInterval := 60 * 60 * 2

	var wg sync.WaitGroup
	start := time.Now()

	step := 10
	upper_a := 40
	upper_b := 10

	len_a := upper_a / step
	len_b := upper_b / step

	results := make([]result.Result, len_a*len_b)

	for b := step; b <= upper_b; b += step {
		for a := step; a <= upper_a; a += step {
			i := (b/step-1)*len_a + a/step - 1
			strategy := strat.NewTwoIntervalAroundPriceStrategy(startAmount0, startAmount1, pool, a, b)
			execution := executor.CreateExecution(strategy, startTime, updateInterval, transactions)
			wg.Add(1)
			go runAndAppend(&wg, execution, a, b, i, results)
		}
		wg.Wait()
	}

	saveFile(results)

	t := time.Now()
	fmt.Println("Time: ", t.Sub(start))
	fmt.Println("Done")
}

func saveFile(results []result.Result) {
	filename := "2_hours.json"
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
	err = encoder.Encode(results)
	if err != nil {
		return
	}

}

func runAndAppend(wg *sync.WaitGroup, excecution *executor.Execution, a, b, i int, results []result.Result) {
	defer wg.Done()
	excecution.Run()
	res := createResult(excecution, a, b)
	results[i] = res
}

func createResult(excecution *executor.Execution, a, b int) result.Result {

	length := len(excecution.Amounts0)

	startTime := excecution.Timestamps[0]
	endTime := excecution.Timestamps[length-1]
	updateInterval := excecution.UpdateInterval

	amount0Start := excecution.Amounts0[0]
	amount1Start := excecution.Amounts1[0]
	x96Start := excecution.SqrtPricesX96[0]
	priceStart := sqrtmath.GetPrice(x96Start)
	amountEthConvertedStart := new(big.Int).Div(amount1Start.ToBig(), priceStart)
	amountUSDStart := new(big.Int).Add(amount0Start.ToBig(), amountEthConvertedStart)
	amountStart := amountUSDStart.String()

	amount0End := excecution.Amounts0[length-1]
	amount1End := excecution.Amounts1[length-1]
	x96End := excecution.SqrtPricesX96[length-1]
	priceEnd := sqrtmath.GetPrice(x96End)
	amountEthConvertedEnd := new(big.Int).Div(amount1End.ToBig(), priceEnd)
	amountUSDEnd := new(big.Int).Add(amount0End.ToBig(), amountEthConvertedEnd)
	amountEnd := amountUSDEnd.String()

	r := result.Result{
		StartTime:      startTime,
		EndTime:        endTime,
		UpdateInterval: updateInterval,
		AmountStart:    amountStart,
		AmountEnd:      amountEnd,
		ParamterA:      a,
		ParamterB:      b,
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
