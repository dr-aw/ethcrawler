package output

import (
	"fmt"
	"math/big"
	"os"

	// "strings"

	"ethcrawler/pkg/models"

	"github.com/xuri/excelize/v2"
)

// SaveToTextFile saves formatted transfers to a text file
func SaveToTextFile(transfers []models.FormattedTransfer, filename string) error {
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

// SaveToExcel saves formatted transfers to an Excel file
func SaveToExcel(transfers []models.FormattedTransfer, filename string) error {
	fmt.Printf("Creating Excel file with %d transactions...\n", len(transfers))

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing Excel file:", err)
		}
	}()

	// Delete default Sheet1
	f.DeleteSheet("Sheet1")

	// Create a new sheet
	sheetName := "USDT Transactions"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("error creating sheet: %v", err)
	}
	f.SetActiveSheet(index)

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

	// Create number formats
	usdtStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: "#,##0.000000",
	})
	if err != nil {
		return fmt.Errorf("error creating USDT style: %v", err)
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
	if err := f.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle); err != nil {
		return fmt.Errorf("error applying header style: %v", err)
	}

	// For large datasets, use streaming mode to reduce memory usage
	streamWriter, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("error creating stream writer: %v", err)
	}

	// Add data using streaming mode
	for i, tx := range transfers {
		row := i + 2 // +2 because Excel rows are 1-indexed and we have a header row

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

		// Write row data
		err := streamWriter.SetRow(fmt.Sprintf("A%d", row), []interface{}{
			tx.Date,
			tx.From,
			tx.To,
			valueWei,
			excelize.Cell{Value: valueUSDT, StyleID: usdtStyle},
			tx.Hash,
		})

		if err != nil {
			return fmt.Errorf("error setting row %d: %v", row, err)
		}

		// Report progress for large datasets
		if i > 0 && i%5000 == 0 {
			fmt.Printf("Processed %d of %d transactions...\n", i, len(transfers))
		}
	}

	// Finalize the stream writer
	if err := streamWriter.Flush(); err != nil {
		return fmt.Errorf("error flushing stream writer: %v", err)
	}

	// Set column widths
	columnWidths := []float64{20, 45, 45, 20, 15, 70}
	for i, width := range columnWidths {
		colName := string(rune('A' + i))
		if err := f.SetColWidth(sheetName, colName, colName, width); err != nil {
			return fmt.Errorf("error setting column width: %v", err)
		}
	}

	// Add filters to the header row
	if err := f.AutoFilter(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", ""); err != nil {
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
