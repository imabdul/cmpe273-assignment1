package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/bitly/go-simplejson"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/justinas/alice"
)

type PurchaseRequest struct {
	Budget                   float32
	StockSymbolAndPercentage string
}

type PurchaseResponse struct {
	TradeId         int
	Stocks           []string
	UninvestedAmount float32
}

type CheckResponse struct {
	Stocks           []string
	CurrentMarketValue float32
	UninvestedAmount float32
	
}

type CheckRequest struct {
	TradeId string
}

type StockAccounts struct {
	stockPortfolio map[int](*Portfolio)
}

type Portfolio struct {
	stocks           map[string](*Share)
	uninvestedAmount float32
}

type Share struct {
	shareNum    int
	boughtPrice float32
}


//Stock Accounts variable declaration
var st StockAccounts

//Trade Id variable declaration
var tradeId int

//purchase shares, maintain stock account and provide summary of Stock Account 
func (st *StockAccounts) Buy(httpRq *http.Request, rq *PurchaseRequest, rsp *PurchaseResponse) error {

	//Trade Id gets increased by 1 in order to keep it unique.
	tradeId++
	rsp.TradeId = tradeId

	//account setup if not exists already
	if st.stockPortfolio == nil {

		st.stockPortfolio = make(map[int](*Portfolio))

		st.stockPortfolio[tradeId] = new(Portfolio)
		st.stockPortfolio[tradeId].stocks = make(map[string]*Share)

	}

	//Purchase arguments get parsed here
	
	symbolAndPercentages := strings.Split(rq.StockSymbolAndPercentage, ",")
	newBudget := float32(rq.Budget)
	
	var amtSpent float32

	for _, stk := range symbolAndPercentages {

		//Different share and there associated budget get parsed here

		splited := strings.Split(stk, ":")
		stkQuote := splited[0]
		percentage := splited[1]
		strPercentage := strings.TrimSuffix(percentage, "%")
		floatPercentage64, _ := strconv.ParseFloat(strPercentage, 32)
		floatPercentage := float32(floatPercentage64 / 100.00)
		currentPrice := checkQuote(stkQuote)

		shares := int(math.Floor(float64(newBudget * floatPercentage / currentPrice)))
		sharesFloat := float32(shares)
		amtSpent += sharesFloat * currentPrice

		// for every new trade id, setting up portfolio first
		
		if _, ok := st.stockPortfolio[tradeId]; !ok {

			newPortfolio := new(Portfolio)
			newPortfolio.stocks = make(map[string]*Share)
			st.stockPortfolio[tradeId] = newPortfolio
		}
		if _, ok := st.stockPortfolio[tradeId].stocks[stkQuote]; !ok {

			newShare := new(Share)
			newShare.boughtPrice = currentPrice
			newShare.shareNum = shares
			st.stockPortfolio[tradeId].stocks[stkQuote] = newShare
		} else {

			total := float32(sharesFloat*currentPrice) + float32(st.stockPortfolio[tradeId].stocks[stkQuote].shareNum)*st.stockPortfolio[tradeId].stocks[stkQuote].boughtPrice
			st.stockPortfolio[tradeId].stocks[stkQuote].boughtPrice = total / float32(shares+st.stockPortfolio[tradeId].stocks[stkQuote].shareNum)
			st.stockPortfolio[tradeId].stocks[stkQuote].shareNum += shares
		}

		stockBought := stkQuote + ":" + strconv.Itoa(shares) + ":$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)

		rsp.Stocks = append(rsp.Stocks, stockBought)
	}

	//Uninvested Amount gets calculated here
	amtLeftOver := newBudget - amtSpent
	rsp.UninvestedAmount = amtLeftOver
	st.stockPortfolio[tradeId].uninvestedAmount += amtLeftOver

	return nil
}

//Account check using trade Id
func (st *StockAccounts) Check(httpRq *http.Request, checkRq *CheckRequest, checkResp *CheckResponse) error {

	if st.stockPortfolio == nil {
		return errors.New("No account set up yet.")
	}

	//Argument being parsed to get the tradeId
	tradeId64, err := strconv.ParseInt(checkRq.TradeId, 10, 64)

	if err != nil {
		return errors.New("Invalid Trade ID. ")
	}
	tradeId := int(tradeId64)

	if pocket, ok := st.stockPortfolio[tradeId]; ok {

		var currentMarketVal float32
		for stockquote, sh := range pocket.stocks {
		
			//Present price
			currentPrice := checkQuote(stockquote)

			//Price up or down calculations
			var str string
			if sh.boughtPrice < currentPrice {
				str = "+$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else if sh.boughtPrice > currentPrice {
				str = "-$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else {
				str = "$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			}

			//object setup to send response
			entry := stockquote + ":" + strconv.Itoa(sh.shareNum) + ":" + str
			checkResp.Stocks = append(checkResp.Stocks, entry)
			currentMarketVal += float32(sh.shareNum) * currentPrice
		}

		//uninvested amount
		checkResp.UninvestedAmount = pocket.uninvestedAmount

		// total Market Value
		checkResp.CurrentMarketValue = currentMarketVal
	} else {
		return errors.New("This trade ID doesn't exists")
	}

	return nil
}

func main() {

	//stock account Initialization
	var st = (new(StockAccounts))

	//Trade Id random generator
	tradeId = rand.Intn(99999) + 1

	//start listening
	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(st, "")

	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	router.Handle("/rpc", chain.Then(server))
	log.Fatal(http.ListenAndServe(":8070", server))

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func checkQuote(stockName string) float32 {

	//only one stock is being queried using Yahoo API every time in order to simplify it.
	urlLeftPart := "https://query.yahooapis.com/v1/public/yql?q=select%20LastTradePriceOnly%20from%20yahoo.finance%0A.quotes%20where%20symbol%20%3D%20%22"
	urlRightPart := "%22%0A%09%09&format=json&env=http%3A%2F%2Fdatatables.org%2Falltables.env"

	//request to API
	resp, err := http.Get(urlLeftPart + stockName + urlRightPart)
	//fmt.Print("URL:" + urlLeftPart + stockName + urlRightPart)
	if err != nil {
		log.Fatal(err)
	}

	//body reading
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	//query success check
	if resp.StatusCode != 200 {
		log.Fatal("Query Failed!! Reason could be invalid stock quote or no network connection")
	}

	//body into NewJson object conversion
	newjson, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println(err)
	}

	//Present stock price
	price, _ := newjson.Get("query").Get("results").Get("quote").Get("LastTradePriceOnly").String()
	floatPrice, err := strconv.ParseFloat(price, 32)
	
	return float32(floatPrice)
}
