// Needed .env file with ETHERSCAN_API_KEY=XXXXXXXXXXXXXXX
// and USDT_CONTRACT=0xdac17f958d2ee523a2206206994597c13d831ec7

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ethcrawler/pkg/etherscan"
	"ethcrawler/pkg/output"

	"github.com/joho/godotenv"
)

// Значения по умолчанию
const (
	DefaultUsdtContract = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	EnvFileName         = ".env"
	ConfFileName        = "ethcrawler.conf"
)

func main() {
	// Parse command line arguments
	addressFlag := flag.String("a", "", "Ethereum address")
	outputFormat := flag.String("format", "both", "Output format: text, excel, or both")
	configFile := flag.String("config", "", "Path to config file (.env or .conf)")
	flag.Parse()

	// Приветствие
	fmt.Printf("%sEthCrawler - USDT Transaction Tool%s\n\n",
		etherscan.ColorGreen, etherscan.ColorReset)

	// Интерактивный режим, если адрес не указан через аргументы
	address := *addressFlag
	if address == "" {
		address = promptForEthereumAddress()
	}

	// Проверка валидности адреса
	if len(address) != 42 || address[:2] != "0x" {
		fmt.Printf("%sAddress has to start from 0x and contain 40 hex-symbols%s\n",
			etherscan.ColorRed, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	fmt.Printf("%sFetching transactions for address: %s%s\n",
		etherscan.ColorGreen, address, etherscan.ColorReset)

	// Load environment variables and handle first run setup
	apiKey, contract := setupConfiguration(*configFile)

	// Create a new Etherscan client
	client := etherscan.NewClient(apiKey, contract)

	// Get the token transfers
	transfers, err := client.GetTokenTransfers(address)
	if err != nil {
		fmt.Printf("%sError fetching transfers: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	// Format the transfers
	formattedTransfers, err := etherscan.FormatTransfers(transfers)
	if err != nil {
		fmt.Printf("%sError formatting transfers: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	// Save the transfers in the requested format(s)
	if *outputFormat == "text" || *outputFormat == "both" {
		filename, err := output.SaveToTextFile(formattedTransfers, address)
		if err != nil {
			fmt.Printf("%sError saving text file: %v%s\n",
				etherscan.ColorRed, err, etherscan.ColorReset)
		} else {
			fmt.Printf("%sTransactions saved to `%s`%s\n",
				etherscan.ColorGreen, filename, etherscan.ColorReset)
		}
	}

	if *outputFormat == "excel" || *outputFormat == "both" {
		filename, err := output.SaveToExcel(formattedTransfers, address)
		if err != nil {
			fmt.Printf("%sError saving Excel file: %v%s\n",
				etherscan.ColorRed, err, etherscan.ColorReset)
		} else {
			fmt.Printf("%sTransactions saved to `%s`%s\n",
				etherscan.ColorGreen, filename, etherscan.ColorReset)
		}
	}

	// Финальное сообщение и пауза перед выходом
	fmt.Printf("\n%sOperation completed. Files saved in the same directory as the program.%s\n",
		etherscan.ColorGreen, etherscan.ColorReset)

	waitForEnter()
}

// promptForEthereumAddress запрашивает Ethereum адрес у пользователя
func promptForEthereumAddress() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%sPlease enter an Ethereum address (starting with 0x): %s",
			etherscan.ColorGreen, etherscan.ColorReset)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%sError reading input: %v%s\n",
				etherscan.ColorRed, err, etherscan.ColorReset)
			continue
		}

		// Убираем переводы строк и пробелы
		address := strings.TrimSpace(input)

		// Проверка формата адреса (0x + 40 hex символов)
		if isValidEthereumAddress(address) {
			return address
		}

		fmt.Printf("%sInvalid Ethereum address format. Address should start with 0x followed by 40 hex characters.%s\n\n",
			etherscan.ColorRed, etherscan.ColorReset)
	}
}

// isValidEthereumAddress проверяет валидность Ethereum адреса
func isValidEthereumAddress(address string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(address)
}

// waitForEnter ожидает нажатия Enter
func waitForEnter() {
	fmt.Printf("\n%sPress Enter to exit...%s",
		etherscan.ColorYellow, etherscan.ColorReset)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// setupConfiguration загружает или создает конфигурационный файл и возвращает настройки
func setupConfiguration(customConfigPath string) (string, string) {
	var configPath string

	// Если указан пользовательский путь к конфигу, используем его
	if customConfigPath != "" {
		configPath = customConfigPath
		if !fileExists(configPath) {
			fmt.Printf("%sSpecified config file not found: %s%s\n",
				etherscan.ColorRed, configPath, etherscan.ColorReset)
			waitForEnter()
			os.Exit(1)
		}
	} else {
		// Иначе ищем конфигурационные файлы в стандартных местах
		configPath = findConfigFile()
	}

	// Если файл существует, загружаем из него данные
	if configPath != "" {
		apiKey, contract := loadConfigFile(configPath)
		return apiKey, contract
	}

	// Если не нашли подходящий файл, создаем новый .conf по умолчанию
	fmt.Printf("%sConfiguration file not found. Setting up for first use.%s\n",
		etherscan.ColorYellow, etherscan.ColorReset)

	configPath = getDefaultConfigPath()
	apiKey := promptForAPIKey()
	contract := DefaultUsdtContract

	// Создаем конфигурационный файл
	saveToConfigFile(configPath, apiKey, contract)

	return apiKey, contract
}

// findConfigFile ищет конфигурационные файлы в стандартных местах
func findConfigFile() string {
	// Пути для поиска по приоритету
	searchPaths := []string{
		getConfigPath(ConfFileName), // Сначала ищем .conf файл рядом с exe
		getConfigPath(EnvFileName),  // Затем .env файл рядом с exe
		ConfFileName,                // Затем .conf в текущей директории
		EnvFileName,                 // Затем .env в текущей директории
	}

	for _, path := range searchPaths {
		if fileExists(path) {
			fmt.Printf("%sFound configuration file: %s%s\n",
				etherscan.ColorGreen, path, etherscan.ColorReset)
			return path
		}
	}

	return ""
}

// getConfigPath возвращает путь к файлу конфигурации рядом с исполняемым файлом
func getConfigPath(fileName string) string {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		// В случае ошибки используем текущую директорию
		return fileName
	}

	// Используем директорию, в которой находится исполняемый файл
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, fileName)
}

// getDefaultConfigPath возвращает путь к конфигурационному файлу по умолчанию
func getDefaultConfigPath() string {
	return getConfigPath(ConfFileName)
}

// loadConfigFile загружает данные из конфигурационного файла
func loadConfigFile(path string) (string, string) {
	// Проверяем расширение файла
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".env" {
		return loadEnvFile(path)
	} else {
		return loadConfFile(path)
	}
}

// loadEnvFile загружает данные из .env файла
func loadEnvFile(path string) (string, string) {
	err := godotenv.Load(path)
	if err != nil {
		fmt.Printf("%sError loading .env file: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	contract := os.Getenv("USDT_CONTRACT")

	// Проверка наличия API ключа
	if apiKey == "" {
		fmt.Printf("%sAPI key not found in %s%s\n",
			etherscan.ColorYellow, path, etherscan.ColorReset)
		apiKey = promptForAPIKey()
		saveToEnvFile(path, apiKey, contract)
	}

	// Если контракт не указан, используем значение по умолчанию
	if contract == "" {
		contract = DefaultUsdtContract
		saveToEnvFile(path, apiKey, contract)
	}

	return apiKey, contract
}

// loadConfFile загружает данные из .conf файла
func loadConfFile(path string) (string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%sError reading conf file: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")

	apiKey := ""
	contract := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "ETHERSCAN_API_KEY" {
			apiKey = value
		} else if key == "USDT_CONTRACT" {
			contract = value
		}
	}

	// Проверка наличия API ключа
	if apiKey == "" {
		fmt.Printf("%sAPI key not found in %s%s\n",
			etherscan.ColorYellow, path, etherscan.ColorReset)
		apiKey = promptForAPIKey()
		saveToConfFile(path, apiKey, contract)
	}

	// Если контракт не указан, используем значение по умолчанию
	if contract == "" {
		contract = DefaultUsdtContract
		saveToConfFile(path, apiKey, contract)
	}

	return apiKey, contract
}

// promptForAPIKey запрашивает API ключ у пользователя
func promptForAPIKey() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%sPlease enter your Etherscan API key: %s",
		etherscan.ColorGreen, etherscan.ColorReset)

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("%sError reading input: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	// Убираем переводы строк и пробелы
	apiKey := strings.TrimSpace(input)

	if apiKey == "" {
		fmt.Printf("%sAPI key cannot be empty. Please try again.%s\n",
			etherscan.ColorRed, etherscan.ColorReset)
		return promptForAPIKey()
	}

	return apiKey
}

// saveToConfigFile сохраняет настройки в конфигурационный файл
func saveToConfigFile(path, apiKey, contract string) {
	// Проверяем расширение файла
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".env" {
		saveToEnvFile(path, apiKey, contract)
	} else {
		saveToConfFile(path, apiKey, contract)
	}
}

// saveToEnvFile сохраняет настройки в .env файл
func saveToEnvFile(path, apiKey, contract string) {
	content := fmt.Sprintf("ETHERSCAN_API_KEY=%s\nUSDT_CONTRACT=%s\n",
		apiKey, contract)

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("%sError saving configuration: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	fmt.Printf("%sConfiguration saved to %s%s\n",
		etherscan.ColorGreen, path, etherscan.ColorReset)
}

// saveToConfFile сохраняет настройки в .conf файл
func saveToConfFile(path, apiKey, contract string) {
	content := fmt.Sprintf("# EthCrawler configuration file\n\n# Etherscan API key\nETHERSCAN_API_KEY=%s\n\n# USDT contract address\nUSDT_CONTRACT=%s\n",
		apiKey, contract)

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("%sError saving configuration: %v%s\n",
			etherscan.ColorRed, err, etherscan.ColorReset)
		waitForEnter()
		os.Exit(1)
	}

	fmt.Printf("%sConfiguration saved to %s%s\n",
		etherscan.ColorGreen, path, etherscan.ColorReset)
}

// fileExists проверяет существование файла
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
