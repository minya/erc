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

// GetBalanceInfo gets balance
func GetBalanceInfo(ercLogin string, ercPassword string, accNumber string, date time.Time) (BalanceInfo, error) {
	client, err := getAuthContext(ercLogin, ercPassword)
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
	return parseBalance(html), nil
}

// GetReceipt receives receipt for account
func GetReceipt(ercLogin string, ercPassword string, accNumber string) ([]byte, error) {
	client, err := getAuthContext(ercLogin, ercPassword)
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

func getAuthContext(ercLogin string, ercPassword string) (*http.Client, error) {
	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport:     transport,
		Jar:           jar,
		CheckRedirect: checkRedirect,
	}
	data := url.Values{}
	data.Set("smth", "")
	data.Set("username", ercLogin)
	data.Set("password", ercPassword)

	respLogin, errLogin := client.PostForm(ercPrivareOfficeURL, data)

	if nil == respLogin || respLogin.StatusCode != 302 {
		log.Printf("Login error: %v \n", errLogin)
		if nil != respLogin {
			return nil, fmt.Errorf("Error: Login response code: %v ", respLogin.StatusCode)
		} else {
			return nil, fmt.Errorf("Error: unable to log in ")
		}
	}
	return &client, nil
}

func parseBalance(html string) BalanceInfo {
	reRow, _ := regexp.Compile("(<td>(.+?)</td>\\s+?<td class='sum'><b>(.?|.+?)</b></td>)")
	match := reRow.FindAllStringSubmatch(html, -1)

	var result BalanceInfo
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

	return result
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
