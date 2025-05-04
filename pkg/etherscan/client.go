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
	pageSize := 5000 // Etherscan allows max 10000, but we'll use a smaller size to be safe
	page := 1

	for {
		fmt.Printf("%sDownloading page %d (transactions %d-%d)...%s\n",
			ColorYellow, page, ((page-1)*pageSize)+1, page*pageSize, ColorReset)

		url := fmt.Sprintf(
			"%s?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=asc&apikey=%s",
			c.BaseURL, c.Contract, address, page, pageSize, c.ApiKey,
		)

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
			break
		}

		// Increment page for next request
		page++

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
