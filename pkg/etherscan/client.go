package etherscan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"ethcrawler/pkg/models"
)

// Colors for console output
var (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// Client represents an Etherscan API client
type Client struct {
	ApiKey   string
	Contract string
	BaseURL  string
}

// NewClient creates a new Etherscan API client
func NewClient(apiKey, contract string) *Client {
	return &Client{
		ApiKey:   apiKey,
		Contract: contract,
		BaseURL:  "https://api.etherscan.io/api",
	}
}

// GetTokenTransfers fetches ERC20 token transfers for a given address
func (c *Client) GetTokenTransfers(address string) ([]models.ERC20Transfer, error) {
	var allTransfers []models.ERC20Transfer

	// Etherscan API limitation: page * offset must be <= 10000
	// Using a dynamic pagination strategy to handle large datasets
	maxWindow := 10000
	initialPageSize := 5000

	// First fetch with large page size to get most results efficiently
	pageSize := initialPageSize
	page := 1
	startBlock := 0

	// Display counter for showing logical page numbers to user
	displayPage := 1

	for {
		// Check if we'd exceed the API limit with current page
		if page*pageSize > maxWindow {
			// We need to reset our pagination strategy
			// Save the last block we processed
			if len(allTransfers) > 0 {
				lastTransfer := allTransfers[len(allTransfers)-1]
				blockNum, err := models.StringToInt(lastTransfer.BlockNumber)
				if err != nil {
					return nil, fmt.Errorf("error converting block number: %v", err)
				}
				// Start from the next block
				startBlock = blockNum + 1
			}
			// Reset pagination for API, but keep display counter incrementing
			page = 1
		}

		// Calculate display transaction range based on total count so far
		displayStart := ((displayPage - 1) * pageSize) + 1
		displayEnd := displayPage * pageSize

		fmt.Printf("%sDownloading page %d (transactions %d-%d)%s",
			ColorYellow, displayPage, displayStart, displayEnd, ColorReset)

		if startBlock > 0 {
			fmt.Printf("%s from block %d%s\n", ColorYellow, startBlock, ColorReset)
		} else {
			fmt.Println(ColorReset)
		}

		// Build URL with startBlock parameter if needed
		var url string
		if startBlock > 0 {
			url = fmt.Sprintf(
				"%s?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=asc&startblock=%d&apikey=%s",
				c.BaseURL, c.Contract, address, page, pageSize, startBlock, c.ApiKey,
			)
		} else {
			url = fmt.Sprintf(
				"%s?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=asc&apikey=%s",
				c.BaseURL, c.Contract, address, page, pageSize, c.ApiKey,
			)
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error making request: %v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		var raw models.EtherscanResponse
		err = json.Unmarshal(body, &raw)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling response: %v", err)
		}

		if raw.Status != "1" {
			return nil, fmt.Errorf("Etherscan API error: %v", raw.Message)
		}

		var pageTransfers []models.ERC20Transfer
		err = json.Unmarshal(raw.Result, &pageTransfers)
		if err != nil {
			return nil, fmt.Errorf("error parsing list of transactions: %v", err)
		}

		// Add this page's transfers to the total
		allTransfers = append(allTransfers, pageTransfers...)

		// If we got fewer transfers than the page size, we've reached the end
		if len(pageTransfers) < pageSize {
			// If we were using a startBlock filter, we need to check if
			// we've actually reached the end or just need to reset again
			if startBlock > 0 && len(pageTransfers) == pageSize {
				// Continue with next block range
				lastTransfer := pageTransfers[len(pageTransfers)-1]
				blockNum, err := models.StringToInt(lastTransfer.BlockNumber)
				if err != nil {
					return nil, fmt.Errorf("error converting block number: %v", err)
				}
				startBlock = blockNum + 1
				page = 1
				// Do not reset displayPage, it continues to increase
				displayPage++
				continue
			}
			// We've truly reached the end
			break
		}

		// Increment page for next request
		page++
		displayPage++

		// Sleep to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("%sDownloaded %d transactions total%s\n",
		ColorGreen, len(allTransfers), ColorReset)

	return allTransfers, nil
}

// FormatTransfers converts raw transfers to formatted transfers
func FormatTransfers(transfers []models.ERC20Transfer) ([]models.FormattedTransfer, error) {
	var formatted []models.FormattedTransfer

	for _, tx := range transfers {
		date, timestamp, err := models.TimeStampToDate(tx.TimeStamp)
		if err != nil {
			return nil, fmt.Errorf("error formatting timestamp: %v", err)
		}

		formatted = append(formatted, models.FormattedTransfer{
			Date:      date,
			From:      tx.From,
			To:        tx.To,
			Value:     tx.Value,
			Hash:      tx.Hash,
			TimeStamp: timestamp,
		})
	}

	return formatted, nil
}
