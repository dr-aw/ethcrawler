package output

import (
	"fmt"
	"math/big"
	"os"

	// "strings"

	"ethcrawler/pkg/models"

	"github.com/xuri/excelize/v2"
)

// GenerateFileName generates a filename with the address prefix
func GenerateFileName(address string, fileType string) string {
	// Use the first 10 characters of the address (including 0x)
	shortAddress := address
	if len(address) > 10 {
		shortAddress = address[:10]
	}

	return fmt.Sprintf("usdt_transactions_%s.%s", shortAddress, fileType)
}

// SaveToTextFile saves formatted transfers to a text file with address in filename
func SaveToTextFile(transfers []models.FormattedTransfer, address string) (string, error) {
	filename := GenerateFileName(address, "txt")
	err := saveToTextFileImpl(transfers, filename)
	return filename, err
}

// Internal implementation function for text file saving
func saveToTextFileImpl(transfers []models.FormattedTransfer, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer f.Close()

	for _, tx := range transfers {
		line := fmt.Sprintf("%s | FROM: %s | TO: %s | VALUE: %s | HASH: %s\n",
			tx.Date, tx.From, tx.To, tx.Value, tx.Hash)
		_, err := f.WriteString(line)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}

// SaveToTextFileWithName saves formatted transfers to a text file with specific filename
func SaveToTextFileWithName(transfers []models.FormattedTransfer, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer f.Close()

	for _, tx := range transfers {
		line := fmt.Sprintf("%s | FROM: %s | TO: %s | VALUE: %s | HASH: %s\n",
			tx.Date, tx.From, tx.To, tx.Value, tx.Hash)
		_, err := f.WriteString(line)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}

// SaveToExcel saves formatted transfers to an Excel file with address in filename
func SaveToExcel(transfers []models.FormattedTransfer, address string) (string, error) {
	filename := GenerateFileName(address, "xlsx")
	err := saveToExcelImpl(transfers, filename)
	return filename, err
}

// Internal implementation function for Excel file saving
func saveToExcelImpl(transfers []models.FormattedTransfer, filename string) error {
	fmt.Printf("Creating Excel file with %d transactions...\n", len(transfers))

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing Excel file:", err)
		}
	}()

	// Create a new sheet
	sheetName := "USDT Transactions"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("error creating sheet: %v", err)
	}

	// Delete default Sheet1
	f.DeleteSheet("Sheet1")

	// Set the active sheet
	f.SetActiveSheet(1)

	// Set headers style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DDEBF7"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating header style: %v", err)
	}

	// Add header
	headers := []string{"Date", "From", "To", "Value (Wei)", "Value (USDT)", "Hash"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return fmt.Errorf("error setting header value: %v", err)
		}
	}

	// Apply header style to header row
	lastCol := string(rune('A' + len(headers) - 1))
	if err := f.SetCellStyle(sheetName, "A1", lastCol+"1", headerStyle); err != nil {
		return fmt.Errorf("error applying header style: %v", err)
	}

	// Add data in batches to avoid memory issues
	const batchSize = 5000
	for i := 0; i < len(transfers); i += batchSize {
		end := i + batchSize
		if end > len(transfers) {
			end = len(transfers)
		}

		fmt.Printf("Processing transactions %d-%d of %d...\n", i+1, end, len(transfers))

		// Process this batch
		for j := i; j < end; j++ {
			tx := transfers[j]
			row := j + 2 // +2 because Excel rows are 1-indexed and we have a header row

			// Format USDT value (convert from wei, 1 USDT = 10^6 wei for USDT)
			valueWei := tx.Value
			valueUSDT := 0.0

			// Use big.Float for more accurate calculation
			if val, ok := new(big.Float).SetString(valueWei); ok {
				divisor := new(big.Float).SetFloat64(1e6)
				val.Quo(val, divisor)

				// Convert to float64 for display
				valueUSDT, _ = val.Float64()
			}

			// Add row data
			cells := []interface{}{
				tx.Date,
				tx.From,
				tx.To,
				valueWei,
				valueUSDT,
				tx.Hash,
			}

			for k, value := range cells {
				cell := fmt.Sprintf("%c%d", 'A'+k, row)
				if err := f.SetCellValue(sheetName, cell, value); err != nil {
					return fmt.Errorf("error setting cell value at %s: %v", cell, err)
				}
			}
		}
	}

	// Set column widths
	columnWidths := []float64{20, 45, 45, 20, 15, 70}
	for i, width := range columnWidths {
		colName := string(rune('A' + i))
		if err := f.SetColWidth(sheetName, colName, colName, width); err != nil {
			return fmt.Errorf("error setting column width: %v", err)
		}
	}

	// Add filter
	filterRange := fmt.Sprintf("A1:%s1", lastCol)
	if err := f.AutoFilter(sheetName, filterRange, []excelize.AutoFilterOptions{}); err != nil {
		return fmt.Errorf("error adding filter: %v", err)
	}

	// Freeze the header row
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return fmt.Errorf("error freezing header row: %v", err)
	}

	// Save the Excel file
	fmt.Println("Saving Excel file...")
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("error saving Excel file: %v", err)
	}

	return nil
}

// SaveToExcelWithName saves formatted transfers to an Excel file with specific filename
func SaveToExcelWithName(transfers []models.FormattedTransfer, filename string) error {
	fmt.Printf("Creating Excel file with %d transactions...\n", len(transfers))

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing Excel file:", err)
		}
	}()

	// Create a new sheet
	sheetName := "USDT Transactions"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("error creating sheet: %v", err)
	}

	// Delete default Sheet1
	f.DeleteSheet("Sheet1")

	// Set the active sheet
	f.SetActiveSheet(1)

	// Set headers style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DDEBF7"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating header style: %v", err)
	}

	// Add header
	headers := []string{"Date", "From", "To", "Value (Wei)", "Value (USDT)", "Hash"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return fmt.Errorf("error setting header value: %v", err)
		}
	}

	// Apply header style to header row
	lastCol := string(rune('A' + len(headers) - 1))
	if err := f.SetCellStyle(sheetName, "A1", lastCol+"1", headerStyle); err != nil {
		return fmt.Errorf("error applying header style: %v", err)
	}

	// Add data in batches to avoid memory issues
	const batchSize = 5000
	for i := 0; i < len(transfers); i += batchSize {
		end := i + batchSize
		if end > len(transfers) {
			end = len(transfers)
		}

		fmt.Printf("Processing transactions %d-%d of %d...\n", i+1, end, len(transfers))

		// Process this batch
		for j := i; j < end; j++ {
			tx := transfers[j]
			row := j + 2 // +2 because Excel rows are 1-indexed and we have a header row

			// Format USDT value (convert from wei, 1 USDT = 10^6 wei for USDT)
			valueWei := tx.Value
			valueUSDT := 0.0

			// Use big.Float for more accurate calculation
			if val, ok := new(big.Float).SetString(valueWei); ok {
				divisor := new(big.Float).SetFloat64(1e6)
				val.Quo(val, divisor)

				// Convert to float64 for display
				valueUSDT, _ = val.Float64()
			}

			// Add row data
			cells := []interface{}{
				tx.Date,
				tx.From,
				tx.To,
				valueWei,
				valueUSDT,
				tx.Hash,
			}

			for k, value := range cells {
				cell := fmt.Sprintf("%c%d", 'A'+k, row)
				if err := f.SetCellValue(sheetName, cell, value); err != nil {
					return fmt.Errorf("error setting cell value at %s: %v", cell, err)
				}
			}
		}
	}

	// Set column widths
	columnWidths := []float64{20, 45, 45, 20, 15, 70}
	for i, width := range columnWidths {
		colName := string(rune('A' + i))
		if err := f.SetColWidth(sheetName, colName, colName, width); err != nil {
			return fmt.Errorf("error setting column width: %v", err)
		}
	}

	// Add filter
	filterRange := fmt.Sprintf("A1:%s1", lastCol)
	if err := f.AutoFilter(sheetName, filterRange, []excelize.AutoFilterOptions{}); err != nil {
		return fmt.Errorf("error adding filter: %v", err)
	}

	// Freeze the header row
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return fmt.Errorf("error freezing header row: %v", err)
	}

	// Save the Excel file
	fmt.Println("Saving Excel file...")
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("error saving Excel file: %v", err)
	}

	return nil
}
