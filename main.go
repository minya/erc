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

func init() {
	const (
		defaultLogPath = "erc.log"
	)
	flag.StringVar(&logPath, "logpath", defaultLogPath, "Path to write logs")
}

func main() {

	flag.Parse()

	setUpLogger()

	log.Printf("Start\n")

	settings, settingsErr := readSettings()
	if nil != settingsErr {
		log.Fatalf("read settings: %v \n", settingsErr)
	}

	if len(os.Args) == 1 || (os.Args[1] != "balance" && os.Args[1] != "receipt") {
		fmt.Printf("Usage: erc balance|receipt\n")
		os.Exit(-1)
	}
	method := os.Args[1]
	if method == "balance" {
		balance, _ := erclib.GetBalanceInfo(
			settings.ErcLogin, settings.ErcPassword, settings.AccountNumber, time.Now())
		fmt.Printf("%v\n", balance)
	} else {
		receipt, _ := erclib.GetReceipt(
			settings.ErcLogin, settings.ErcPassword, settings.AccountNumber)
		os.Stdout.Write(receipt)
	}
}

func setUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func readSettings() (Settings, error) {
	var settings Settings
	err := config.UnmarshalJson(&settings, "~/.erc/settings.json")
	return settings, err
}

type Settings struct {
	ErcLogin      string
	ErcPassword   string
	AccountNumber string
}
