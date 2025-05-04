// Needed .env file with ETHERSCAN_API_KEY=XXXXXXXXXXXXXXX 
// and USDT_CONTRACT=0xdac17f958d2ee523a2206206994597c13d831ec7

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type EtherscanResponse struct {
	Status  string         	`json:"status"`
	Message string         	`json:"message"`
	Result  json.RawMessage	`json:"result"`
}

type ERC20Transfer struct {
	TimeStamp string `json:"timeStamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Hash      string `json:"hash"`
}

var colorReset 	= "\033[0m"
var colorRed 	= "\033[31m"
var colorGreen 	= "\033[32m"

func main() {
	address := flag.String("a", "", "Ethereum address")
	flag.Parse()

	if *address == "" {
		log.Fatalf("%sEthereum address is necessary%s\n", colorRed, colorReset)
	}

	if len(*address) != 42 || (*address)[:2] != "0x" {
		log.Fatalf("%sAddress has to start from 0x and contain 40 hex-symbols%s\n", colorRed, colorReset)
	}

	fmt.Printf("%sFetching transactions for address: %s%s\n", colorGreen, *address, colorReset)	
	

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("%sError loading .env file%s\n", colorRed, colorReset)
	}

	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	contract := os.Getenv("USDT_CONTRACT")


	url := fmt.Sprintf(
		"https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&page=1&offset=10000&sort=asc&apikey=%s",
		contract, *address, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%sError request: %v%s\n", colorRed, err, colorReset)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var raw EtherscanResponse
	err = json.Unmarshal(body, &raw)
	if err != nil {
		log.Fatalf("%sJSON error: %v%s\n", colorRed, err, colorReset)
	}

	if raw.Status != "1" {
		log.Fatalf("%sEtherscan error: %v%s", colorRed, raw.Message, colorReset)
	}

	var result []ERC20Transfer
	err = json.Unmarshal(raw.Result, &result)
	if err != nil {
		log.Fatalf("%sError parsing list of transactions: %v%s\n", colorRed, err, colorReset)
	}

	// Open file
	f, err := os.Create("usdt_transactions.txt")
	if err != nil {
		log.Fatalf("%sError creating file: %v%s", colorRed, err, colorReset)
	}
	defer f.Close()

	for _, tx := range result {
		timestamp, _ := timeStampToDate(tx.TimeStamp)
		line := fmt.Sprintf("%s | FROM: %s | TO: %s | VALUE: %s | HASH: %s\n",
			timestamp, tx.From, tx.To, tx.Value, tx.Hash)
		f.WriteString(line)
	}

	fmt.Printf("%sTransactions saved to `usdt_transactions.txt`%s\n", colorGreen, colorReset)
}

func timeStampToDate(ts string) (string, error) {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "", err
	}
	t := time.Unix(sec, 0)
	return t.Format("2006-01-02 15:04:05"), nil
}
