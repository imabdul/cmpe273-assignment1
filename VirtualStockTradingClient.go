package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/bitly/go-simplejson"
)

//one check for all different errors
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		log.Fatal("error: ", err)
		os.Exit(2)
	}

}



func main() {

	// arguments check
	if len(os.Args) > 4 || len(os.Args) < 2 {
		fmt.Println("Invalid number of arguments passed!")
		return

	} else if len(os.Args) == 2 {

		_, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Invalid argument passed!")
			return
		}

		data, err := json.Marshal(map[string]interface{}{
			"method": "StockAccounts.Check",
			"id":     1,
			"params": []map[string]interface{}{map[string]interface{}{"TradeId": os.Args[1]}},
		})

		if err != nil {
			log.Fatalf("Marshal : %v", err)
		}

		resp, err := http.Post("http://127.0.0.1:8070/rpc", "application/json", strings.NewReader(string(data)))

		if err != nil {
			log.Fatalf("Post: %v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatalf("ReadAll: %v", err)
		}

		newjson, err := simplejson.NewJson(body)

		checkError(err)

		fmt.Print("Stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(stocks)

		fmt.Print("Current Market Value: ")
		currentMarketValue, _ := newjson.Get("result").Get("CurrentMarketValue").Float64()
		fmt.Print("$")
		fmt.Println(currentMarketValue)
		
		fmt.Print("Uninvested Amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)

	} else if len(os.Args) == 3 { 
		//check new shares purchase
		budget, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Println("Invalid budget argument passed!")
			return
		}

		data, err := json.Marshal(map[string]interface{}{
			"method": "StockAccounts.Buy",
			"id":     2,
			"params": []map[string]interface{}{map[string]interface{}{"StockSymbolAndPercentage": os.Args[1], "Budget": float32(budget)}},
		})

		if err != nil {
			log.Fatalf("Marshal : %v", err)
		}

		resp, err := http.Post("http://127.0.0.1:8070/rpc", "application/json", strings.NewReader(string(data)))

		if err != nil {
			log.Fatalf("Post: %v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatalf("ReadAll: %v", err)
		}

		newjson, err := simplejson.NewJson(body)

		checkError(err)

		fmt.Print("Trade ID: ")
		tradeid, _ := newjson.Get("result").Get("TradeId").Int()
		fmt.Println(tradeid)

		fmt.Print("Stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(*stocks)

		fmt.Print("Uninvested Amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)

	} else {
		fmt.Println("Unknown error.")
		return
	}

}

