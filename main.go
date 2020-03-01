package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minya/erc/erclib"
	"github.com/minya/goutils/config"
)

var logPath string
var settings cSettings

func main() {
	log.Printf("Start\n")

	if len(os.Args) == 1 {
		usage()
	}
	request := os.Args[1]
	if request != "balance" && request != "receipt" && request != "accounts" {
		usage()
	}
	ercClient := erclib.NewErcClientWithCredentials(settings.ErcLogin, settings.ErcPassword)
	if request == "accounts" {
		accounts, _ := ercClient.GetAccounts()
		for _, acc := range accounts {
			fmt.Printf("%v\t%v\n", acc.Number, acc.Address)
		}
		return
	}

	var accountNumber string
	if len(os.Args) > 2 {
		accountNumber = os.Args[2]
	} else {
		accountNumber = settings.AccountNumber
	}

	if request == "balance" {
		balance, _ := ercClient.GetBalanceInfo(accountNumber, time.Now())
		fmt.Printf("%v\n", balance)
		return
	}

	if request == "receipt" {
		receipt, _ := ercClient.GetReceipt(settings.AccountNumber)
		os.Stdout.Write(receipt)
		return
	}

	log.Fatal("Impossible")
}

func init() {
	const (
		defaultLogPath = "erc.log"
	)
	flag.StringVar(&logPath, "logpath", defaultLogPath, "Path to write logs")
	flag.Parse()
	setUpLogger()
	var settingsErr error
	settings, settingsErr = readSettings()
	if nil != settingsErr {
		log.Fatalf("read settings: %v \n", settingsErr)
	}
}

func setUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func readSettings() (cSettings, error) {
	var settings cSettings
	err := config.UnmarshalJson(&settings, "~/.erc/settings.json")
	return settings, err
}

func usage() {
	fmt.Printf("Usage: erc balance|receipt|accounts\n")
	os.Exit(-1)
}

type cSettings struct {
	ErcLogin      string
	ErcPassword   string
	AccountNumber string
}
