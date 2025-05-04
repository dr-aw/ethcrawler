# USDT Transaction Crawler (Go + Etherscan)

This tool fetches all **outgoing USDT (ERC-20)** transactions from a specified Ethereum address using the Etherscan API. It saves the results to a `.txt` file and/or Excel file for further analysis.

## 🔧 Setup

1. Install **Go 1.18+**
2. Clone the repository:
```bash
git clone https://github.com/yourname/usdt-crawler.git
cd usdt-crawler
```

3. Configuration options:
   - Create a `.env` file:
     ```
     ETHERSCAN_API_KEY=your_etherscan_api_key
     USDT_CONTRACT=0xdAC17F958D2ee523a2206206994597C13D831ec7
     ```
   - OR the program will create a `ethcrawler.conf` file on first run
   - OR specify a custom config file with `-config` flag

4. Install dependencies:
```bash
go mod tidy
```

5. Build the executable:
```bash
go build -o ethcrawler.exe
```

## 🚀 Usage

### Basic Usage
```bash
# Run with a specific Ethereum address
ethcrawler -a 0xYourEthereumAddress

# Run in interactive mode (will prompt for address)
ethcrawler
```

### Advanced Options
```bash
# Specify output format (text, excel, or both)
ethcrawler -a 0xYourEthereumAddress -format text
ethcrawler -a 0xYourEthereumAddress -format excel
ethcrawler -a 0xYourEthereumAddress -format both

# Use a custom configuration file
ethcrawler -a 0xYourEthereumAddress -config path/to/your/config.env
```

## 📦 Features

- Fetches all USDT transactions for a given Ethereum address
- Filters for outgoing transactions only
- Interactive mode for input if no address is provided
- Supports multiple configuration methods:
  - `.env` file
  - `ethcrawler.conf` file
  - Command-line specified config file
- First-run setup with API key prompting
- Multiple output formats:
  - Human-readable .txt file
  - Formatted Excel spreadsheet
- Validates Ethereum address format

## 🛠️ Planned

- Export to SQLite / PostgreSQL
- Analysis of recurring payments (salary-like behavior)
- Web dashboard and visualizations

----------
