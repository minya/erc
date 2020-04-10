package erclib

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/minya/goutils/web"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var ercPrivareOfficeURL = "https://lk.erc-ekb.ru/client/private_office/private_office.htp"

//ErcClient to query data from erc.ur.ru
type ErcClient struct {
	login      string
	password   string
	httpClient *http.Client
}

//NewErcClientWithCredentials creates new erc client
func NewErcClientWithCredentials(ercLogin string, ercPassword string) ErcClient {
	ercClient := ErcClient{
		login:    ercLogin,
		password: ercPassword,
	}
	return ercClient
}

// GetBalanceInfo gets balance
func (ercClient ErcClient) GetBalanceInfo(accNumber string, date time.Time) (BalanceInfo, error) {
	client, err := ercClient.getAuthContext()
	if nil != err {
		return BalanceInfo{}, fmt.Errorf("Authentication error")
	}

	dataURL := ercPrivareOfficeURL + "?ls=" + accNumber
	strDateTo := date.Format("02-01-2006")
	strDateFrom := date.AddDate(0, -1, 0).Format("02-01-2006")
	dataReq := url.Values{}
	dataReq.Set("show", "3")
	dataReq.Set("s_Date", strDateFrom)
	dataReq.Set("e_Date", strDateTo)

	respData, errPost := client.PostForm(dataURL, dataReq)
	if nil != errPost {
		return BalanceInfo{}, errPost
	}

	tr := transform.NewReader(respData.Body, charmap.Windows1251.NewDecoder())
	bytes, _ := ioutil.ReadAll(tr)
	html := string(bytes)

	//fmt.Printf("%v", html)
	return parseBalance(html)
}

// GetReceipt receives receipt for account
func (ercClient ErcClient) GetReceipt(accNumber string) ([]byte, error) {
	client, err := ercClient.getAuthContext()
	var bytesEmpty []byte
	if nil != err {
		return bytesEmpty, fmt.Errorf("Authentication error")
	}

	getReceiptURL := fmt.Sprintf(
		"https://lk.erc-ekb.ru/erc/client/private_office/private_office.htp?receipt=%v&quitance",
		accNumber)

	receiptResponse, err := client.Get(getReceiptURL)
	if err != nil || receiptResponse == nil || receiptResponse.StatusCode != 200 {
		return bytesEmpty, fmt.Errorf("Unable to fetch receipt")
	}
	body, err := ioutil.ReadAll(receiptResponse.Body)
	if err != nil {
		return bytesEmpty, fmt.Errorf("Unable to read receipt's response")
	}
	return body, nil
}

//GetAccounts asd
func (ercClient ErcClient) GetAccounts() ([]Account, error) {
	client, err := ercClient.getAuthContext()
	if nil != err {
		return []Account{}, fmt.Errorf("Authentication error")
	}
	accountsURL := "https://lk.erc-ekb.ru/client/private_office/private_office.htp?ls"
	response, err := client.Get(accountsURL)
	if err != nil || response == nil || response.StatusCode != 200 {
		return []Account{}, fmt.Errorf("Unable to fetch receipt")
	}
	tr := transform.NewReader(response.Body, charmap.Windows1251.NewDecoder())
	bytes, _ := ioutil.ReadAll(tr)
	html := string(bytes)
	//fmt.Printf("%v", html)
	return parseAccounts(html)
}

func (ercClient *ErcClient) getAuthContext() (*http.Client, error) {
	if nil != ercClient.httpClient {
		return ercClient.httpClient, nil
	}

	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport:     transport,
		Jar:           jar,
		CheckRedirect: checkRedirect,
	}
	data := url.Values{}
	data.Set("smth", "")
	data.Set("username", ercClient.login)
	data.Set("password", ercClient.password)

	respLogin, errLogin := client.PostForm(ercPrivareOfficeURL, data)

	if nil == respLogin || respLogin.StatusCode != 302 {
		log.Printf("Login error: %v \n", errLogin)
		if nil != respLogin {
			return nil, fmt.Errorf("Error: Login response code: %v", respLogin.StatusCode)
		}
		return nil, fmt.Errorf("Error: unable to log in")
	}

	ercClient.httpClient = &client
	return ercClient.httpClient, nil
}

func parseBalance(html string) (BalanceInfo, error) {
	reRow, errCompile := regexp.Compile("(<td>(.+?)</td>\\s+?<td class='sum'><b>(.?|.+?)</b></td>)")
	var result BalanceInfo
	if errCompile != nil {
		return result, errCompile
	}

	match := reRow.FindAllStringSubmatch(html, -1)

	if len(match) < 11 {
		return result, fmt.Errorf("No match found")
	}

	result.Month = match[1][3]

	result.Credit.Total, _ = strconv.ParseFloat(match[2][3], 64)
	result.Credit.CompanyPart, _ = strconv.ParseFloat(match[3][3], 64)
	result.Credit.RepairPart, _ = strconv.ParseFloat(match[4][3], 64)

	if len(match) > 5 {
		result.Debit.Total, _ = strconv.ParseFloat(match[5][3], 64)
		result.Debit.CompanyPart, _ = strconv.ParseFloat(match[6][3], 64)
		result.Debit.RepairPart, _ = strconv.ParseFloat(match[7][3], 64)

		result.AtTheEnd.CompanyPart, _ = strconv.ParseFloat(match[9][3], 64)
		result.AtTheEnd.RepairPart, _ = strconv.ParseFloat(match[10][3], 64)
		result.AtTheEnd.Total = result.AtTheEnd.CompanyPart + result.AtTheEnd.RepairPart
	}

	return result, nil
}

func parseAccounts(html string) ([]Account, error) {
	reRow, errCompile := regexp.Compile("<td>\\s+?<a\\shref=\"\\/client\\/private_office\\/private_office.htp\\?ls=(\\d+)\">\\d+<\\/a>\\s+<\\/td>\\s+<td>(.+?)<\\/td>")
	result := []Account{}
	if errCompile != nil {
		return result, errCompile
	}

	match := reRow.FindAllStringSubmatch(html, -1)

	if len(match) < 1 {
		return result, fmt.Errorf("No match found")
	}
	for _, group := range match {
		result = append(result, Account{Number: group[1], Address: group[2]})
	}
	return result, nil
}

func checkRedirect(r *http.Request, rr []*http.Request) error {
	log.Printf("Check redirect")
	return errors.New("Don't redirect")
}

// BalanceInfo struct describes balance
type BalanceInfo struct {
	Month    string
	Credit   Details
	Debit    Details
	AtTheEnd Details
}

// Details struct
type Details struct {
	Total       float64
	CompanyPart float64
	RepairPart  float64
}

// Account struct
type Account struct {
	Number  string
	Address string
}
