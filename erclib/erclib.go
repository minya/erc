package erclib

import (
	"errors"
	"fmt"
	"github.com/minya/goutils/web"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func GetBalanceInfo(ercLogin string, ercPassword string, accNumber string, date time.Time) (BalanceInfo, error) {
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

	ercPrivareOfficeUrl := "https://www.erc.ur.ru/client/private_office/private_office.htp"
	respLogin, errLogin := client.PostForm(ercPrivareOfficeUrl, data)

	if nil == respLogin || respLogin.StatusCode != 302 {
		log.Fatalf("Login error: %v \n", errLogin)
		if nil != respLogin {
			return BalanceInfo{}, fmt.Errorf("Error: Login response code: %v \n", respLogin.StatusCode)
		}
	}

	dataUrl := ercPrivareOfficeUrl + "?ls=" + accNumber
	strDate := date.Format("02-01-2006")
	dataReq := url.Values{}
	dataReq.Set("show", "3")
	dataReq.Set("s_Date", strDate)
	dataReq.Set("e_Date", strDate)

	respData, errPost := client.PostForm(dataUrl, dataReq)
	if nil != errPost {
		return BalanceInfo{}, errPost
	}

	tr := transform.NewReader(respData.Body, charmap.Windows1251.NewDecoder())
	bytes, _ := ioutil.ReadAll(tr)
	html := string(bytes)

	//fmt.Printf("%v", html)
	return parseBalance(html), nil
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

func checkRedirect(r *http.Request, rr []*http.Request) error {
	log.Printf("Check redirect")
	return errors.New("Don't redirect")
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
