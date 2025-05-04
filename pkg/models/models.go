package models

import (
	"encoding/json"
	"strconv"
	"time"
)

// EtherscanResponse represents the response from Etherscan API
type EtherscanResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Result  json.RawMessage`json:"result"`
}

// ERC20Transfer represents a single ERC20 token transfer
type ERC20Transfer struct {
	TimeStamp string `json:"timeStamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Hash      string `json:"hash"`
}

// FormattedTransfer adds a formatted timestamp for display
type FormattedTransfer struct {
	Date      string
	From      string
	To        string
	Value     string
	Hash      string
	TimeStamp int64 // Original timestamp as int for sorting
}

// TimeStampToDate converts Unix timestamp string to a formatted date string
func TimeStampToDate(ts string) (string, int64, error) {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "", 0, err
	}
	t := time.Unix(sec, 0)
	return t.Format("2006-01-02 15:04:05"), sec, nil
} 