# USDT Transaction Crawler (Go + Etherscan)

This tool fetches all **outgoing USDT (ERC-20)** transactions from a specified Ethereum address using the Etherscan API. It saves the results to a `.txt` file for further analysis.

## 🔧 Setup

1. Install **Go 1.18+**
2. Clone the repository:
```bash
git clone https://github.com/yourname/usdt-crawler.git
cd usdt-crawler
```
3. Create a .env file:
```bash
ETHERSCAN_API_KEY=your_etherscan_api_key
USDT_CONTRACT=0xdAC17F958D2ee523a2206206994597C13D831ec7
```
4. Install dependencies:
```bash
go mod tidy
```

## 🚀 Usage
```bash
go run main.go -a <EthereumAddress>
```
## 📦 Features
- Fetches all USDT transactions for a given Ethereum address
- Filters for outgoing transactions only
- Outputs to a human-readable .txt file
- Uses .env for configuration

## 🛠️ Planned
- Pagination support (for >10,000 transactions)
- Export to SQLite / PostgreSQL
- Analysis of recurring payments (salary-like behavior)
- Web dashboard and visualizations

----------
