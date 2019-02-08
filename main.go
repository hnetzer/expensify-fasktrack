package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type FastrackRecord struct {
	transactionDate string
	tollPlaza       string
	cost            string
}

type ExpensifyTransaction struct {
	Created  string  `json:"created"`
	Currensy string  `json:"currensy"`
	Merchant string  `json:"merchant"`
	Amount   float64 `json:"amount"`
}

type ExpensifyCredentials struct {
	PartnerUserID     string `json:"partnerUserID"`
	PartnerUserSecret string `json:"partnerUserSecret"`
}

type ExpensifyInputSettings struct {
	Type            string                 `json:"type"`
	EmployeeEmail   string                 `json:"employeeEmail"`
	TransactionList []ExpensifyTransaction `json:"transactionList"`
}

type ExpensifyRequestJobDescription struct {
	Type          string                 `json:"type"`
	Credentials   ExpensifyCredentials   `json:"credentials"`
	InputSettings ExpensifyInputSettings `json:"inputSettings"`
}

func NewExpensifyTransacation(r FastrackRecord) ExpensifyTransaction {
	// convert string to float for cost

	cost, err := strconv.ParseFloat(strings.TrimPrefix(r.cost, "$"), 32)
	if err != nil {
		log.Fatal(err)
	}

	return ExpensifyTransaction{
		Created:  r.transactionDate,
		Currensy: "USD",
		Merchant: r.tollPlaza,
		Amount:   cost,
	}
}

func NewExpensifyRequestJobDescription(transactions []ExpensifyTransaction) ExpensifyRequestJobDescription {

	credentials := ExpensifyCredentials{
		PartnerUserID:     os.Getenv("PartnerUserID"),
		PartnerUserSecret: os.Getenv("PartnerUserSecret"),
	}

	inputSettings := ExpensifyInputSettings{
		Type:            "create",
		EmployeeEmail:   "henry@able.co",
		TransactionList: transactions,
	}

	return ExpensifyRequestJobDescription{
		Type:          "create",
		Credentials:   credentials,
		InputSettings: inputSettings,
	}
}

func main() {
	fmt.Println("Starting the CSV reading program")

	// Open the file from Fastrack
	file, err := os.Open("fasttrack.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create the CSV reader
	r := csv.NewReader(file)

	var transactions []ExpensifyTransaction
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		ftRecord := FastrackRecord{
			transactionDate: record[1],
			tollPlaza:       record[4],
			cost:            record[9],
		}

		if ftRecord.cost != "-" {
			transaction := NewExpensifyTransacation(ftRecord)
			// fmt.Printf("%+v\n", transaction)
			transactions = append(transactions, transaction)
		}
	}

	requestJobDescription := NewExpensifyRequestJobDescription(transactions)

	// fmt.Printf("%+v\n", requestJobDescription)
	b, err := json.Marshal(requestJobDescription)
	if err != nil {
		log.Fatal(err)
	}

	requestBody := "requestJobDescription=" + string(b)
	fmt.Println(requestBody)

	expensifyEndpoint := "https://integrations.expensify.com/Integration-Server/ExpensifyIntegrations"

	resp, err := http.Post(expensifyEndpoint, "application/json", strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
