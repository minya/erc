package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/minya/erc/erclib"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"time"
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

	balance, _ := erclib.GetBalanceInfo(
		settings.ErcLogin, settings.ErcPassword, settings.AccountNumber, time.Now())

	fmt.Printf("%v\n", balance)
}

func setUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func readSettings() (Settings, error) {
	user, _ := user.Current()
	settingsPath := path.Join(user.HomeDir, ".erc/settings.json")

	var settings Settings
	settingsBin, settingsErr := ioutil.ReadFile(settingsPath)
	if nil != settingsErr {
		return settings, settingsErr
	}

	json.Unmarshal(settingsBin, &settings)
	return settings, nil
}

type Settings struct {
	ErcLogin      string
	ErcPassword   string
	AccountNumber string
}
