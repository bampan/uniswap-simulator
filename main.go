package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"runtime"
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
	runtime.GOMAXPROCS(14)
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
	updateInterval := 60 * 60 * 24

	var wg sync.WaitGroup
	start := time.Now()
	for i := 10; i <= 40000; i += 10 {
		strategy := strat.NewIntervalAroundPriceStrategy(startAmount0, startAmount1, pool, i)
		execution := executor.CreateExecution(strategy, startTime, updateInterval, transactions)
		wg.Add(1)
		go runAndSave(&wg, execution, i)
	}
	wg.Wait()
	t := time.Now()
	fmt.Println("Time: ", t.Sub(start))
	fmt.Println("Done")
}

func runAndSave(wg *sync.WaitGroup, excecution *executor.Execution, i int) {
	defer wg.Done()
	excecution.Run()
	saveExectution(excecution, i)
}

func saveExectution(excecution *executor.Execution, intervalWidth int) {
	filename := fmt.Sprintf("cons_width_%d.json", intervalWidth)
	filepath := path.Join("results", "one_day", filename)
	file, _ := os.Create(filepath)

	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	length := len(excecution.Amounts0)
	snapshots := make([]result.Snapshot, length)

	for i := 0; i < length; i++ {
		amount0 := excecution.Amounts0[i]
		amount1 := excecution.Amounts1[i]
		x96 := excecution.SqrtPricesX96[i]
		price := sqrtmath.GetPrice(x96)
		amount_eth_converted := new(big.Int).Div(amount1.ToBig(), price)
		amountUSD := new(big.Int).Add(amount0.ToBig(), amount_eth_converted)
		snapshots[i] = result.Snapshot{
			Timestamp: excecution.Timestamps[i],
			Amount0:   amount0.ToBig().String(),
			Amount1:   amount1.ToBig().String(),
			Price:     price.String(),
			AmountUSD: amountUSD.String(),
		}
	}

	startTime := snapshots[0].Timestamp
	endTime := snapshots[length-1].Timestamp
	updateInterval := excecution.UpdateInterval
	amountStart := snapshots[0].AmountUSD
	amountEnd := snapshots[length-1].AmountUSD

	result := result.Result{
		StartTime:      startTime,
		EndTime:        endTime,
		UpdateInterval: updateInterval,
		AmountStart:    amountStart,
		AmountEnd:      amountEnd,
		Snapshots:      snapshots,
	}
	encoder.Encode(result)
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
