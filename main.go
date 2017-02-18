package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/minya/goutils/web"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	//"strings"
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

	SetUpLogger()
	log.Printf("Start\n")
	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport:     transport,
		Jar:           jar,
		CheckRedirect: checkRedirect,
	}

	user, _ := user.Current()
	settingsPath := path.Join(user.HomeDir, ".erc/settings.json")

	var settings Settings
	settingsBin, settingsErr := ioutil.ReadFile(settingsPath)
	if nil != settingsErr {
		log.Fatalf("read settings: %v \n", settingsErr)
	}

	json.Unmarshal(settingsBin, &settings)
	data := url.Values{}
	data.Set("smth", "")
	data.Set("username", settings.ErcLogin)
	data.Set("password", settings.ErcPassword)

	ercPrivareOfficeUrl := "https://www.erc.ur.ru/client/private_office/private_office.htp"
	respLogin, errLogin := client.PostForm(ercPrivareOfficeUrl, data)
	//fmt.Printf("Code: %v\n", respLogin.StatusCode)
	if nil == respLogin || respLogin.StatusCode != 302 {
		log.Fatalf("Login error: %v \n", errLogin)
		if nil != respLogin {
			log.Fatalf("Login response code: %v \n", respLogin.StatusCode)
		}
	}

	dataUrl := ercPrivareOfficeUrl + "?ls=" + settings.AccountNumber
	dataReq := url.Values{}
	dataReq.Set("show", "3")
	dataReq.Set("s_Date", "18-02-2017")
	dataReq.Set("e_Date", "18-02-2017")

	respData, _ := client.PostForm(dataUrl, dataReq)

	tr := transform.NewReader(respData.Body, charmap.Windows1251.NewDecoder())
	bytes, _ := ioutil.ReadAll(tr)
	html := string(bytes)

	//fmt.Printf("%v", html)

	balance := parseBalance(html)
	fmt.Printf("%v\n", balance)
}

func checkRedirect(r *http.Request, rr []*http.Request) error {
	log.Printf("Check redirect")
	return errors.New("Don't redirect")
}

func parseBalance(html string) BalanceInfo {
	reRow, _ := regexp.Compile("(<td>(.+?)</td>\\s+?<td class='sum'><b>(.?|.+?)</b></td>)")
	match := reRow.FindAllStringSubmatch(html, -1)

	var result BalanceInfo
	result.Month = match[1][3]

	result.Credit.Total, _ = strconv.ParseFloat(match[2][3], 64)
	result.Credit.CompanyPart, _ = strconv.ParseFloat(match[3][3], 64)
	result.Credit.RepairPart, _ = strconv.ParseFloat(match[4][3], 64)

	result.Debit.Total, _ = strconv.ParseFloat(match[5][3], 64)
	result.Debit.CompanyPart, _ = strconv.ParseFloat(match[6][3], 64)
	result.Debit.RepairPart, _ = strconv.ParseFloat(match[7][3], 64)

	result.AtTheEnd.CompanyPart, _ = strconv.ParseFloat(match[9][3], 64)
	result.AtTheEnd.RepairPart, _ = strconv.ParseFloat(match[10][3], 64)
	result.AtTheEnd.Total = result.AtTheEnd.CompanyPart + result.AtTheEnd.RepairPart

	return result
}

func SetUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

type Settings struct {
	ErcLogin      string
	ErcPassword   string
	AccountNumber string
}

type BalanceInfo struct {
	Month    string
	Credit   Details
	Debit    Details
	AtTheEnd Details
}

type Details struct {
	Total       float64
	CompanyPart float64
	RepairPart  float64
}
