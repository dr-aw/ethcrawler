// Needed .env file with ETHERSCAN_API_KEY=XXXXXXXXXXXXXXX 
// and USDT_CONTRACT=0xdac17f958d2ee523a2206206994597c13d831ec7

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"ethcrawler/pkg/etherscan"
	"ethcrawler/pkg/output"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line arguments
	address := flag.String("a", "", "Ethereum address")
	outputFormat := flag.String("format", "both", "Output format: text, excel, or both")
	flag.Parse()

	// Validate the Ethereum address
	if *address == "" {
		log.Fatalf("%sEthereum address is necessary%s\n", etherscan.ColorRed, etherscan.ColorReset)
	}

	if len(*address) != 42 || (*address)[:2] != "0x" {
		log.Fatalf("%sAddress has to start from 0x and contain 40 hex-symbols%s\n", 
			etherscan.ColorRed, etherscan.ColorReset)
	}

	fmt.Printf("%sFetching transactions for address: %s%s\n", 
		etherscan.ColorGreen, *address, etherscan.ColorReset)

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("%sError loading .env file%s\n", etherscan.ColorRed, etherscan.ColorReset)
	}

	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	contract := os.Getenv("USDT_CONTRACT")

	// Create a new Etherscan client
	client := etherscan.NewClient(apiKey, contract)

	// Get the token transfers
	transfers, err := client.GetTokenTransfers(*address)
	if err != nil {
		log.Fatalf("%sError fetching transfers: %v%s\n", 
			etherscan.ColorRed, err, etherscan.ColorReset)
	}

	// Format the transfers
	formattedTransfers, err := etherscan.FormatTransfers(transfers)
	if err != nil {
		log.Fatalf("%sError formatting transfers: %v%s\n", 
			etherscan.ColorRed, err, etherscan.ColorReset)
	}

	// Save the transfers in the requested format(s)
	if *outputFormat == "text" || *outputFormat == "both" {
		err = output.SaveToTextFile(formattedTransfers, "usdt_transactions.txt")
		if err != nil {
			log.Fatalf("%sError saving text file: %v%s\n", 
				etherscan.ColorRed, err, etherscan.ColorReset)
		}
		fmt.Printf("%sTransactions saved to `usdt_transactions.txt`%s\n", 
			etherscan.ColorGreen, etherscan.ColorReset)
	}

	if *outputFormat == "excel" || *outputFormat == "both" {
		err = output.SaveToExcel(formattedTransfers, "usdt_transactions.xlsx")
		if err != nil {
			log.Fatalf("%sError saving Excel file: %v%s\n", 
				etherscan.ColorRed, err, etherscan.ColorReset)
		}
		fmt.Printf("%sTransactions saved to `usdt_transactions.xlsx`%s\n", 
			etherscan.ColorGreen, etherscan.ColorReset)
	}
}
