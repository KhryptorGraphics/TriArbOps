package main


import (
	"github.com/gorilla/websocket"
	"github.com/buger/jsonparser"
	"fmt"
    "strconv"
    "sync"
    "time"
	"crypto/hmac"
	"crypto/sha512"
	"net/http"
	"io/ioutil"
	"strings"
)


// Custom Variables - change these
var min_amount = 5.0 // Minimum amount per trade  (in USDT) (Shouldn't be less than 5)
var max_amount = 25.0 // Maximum amount per trade (in USDT) (This is also the amount you need to have in EACH currency you trade)
var fee = 0.006 // ~ cumulated fee for 3 trades
var curs = [...]string{"GRIN", "BEAM", "XMR"}//"BCH", "ETH", "ETC", "QTUM", "LTC", "DASH", "ZEC", "BTM", "EOS", "SNT", "OMG", "PAY", "ZRX", "XMR", "XRP", "DOGE", "BAT", "BTG", "LRC", "STORJ", "AE", "XTZ", "XLM", "MOBI", "OCN", "ZPT", "JNT", "GXS", "RUFF", "TNC", "DDD", "MDT", "GTC", "QLC", "DBC", "BTF", "ADA", "LSK", "WAVES", "BIFI", "QASH", "POWR", "BCD", "SBTC", "GOD", "BCX", "INK", "NEO", "GAS", "IOTA", "NAS", "OAX", "BTS", "GT", "ATOM", "XEM", "BU", "BCHSV", "DCR", "BCN", "XMC", "GRIN", "BEAM", "LYM", "LEO", "HC", "XVG", "NANO", "MXC"}
var KEY  = ""; // gate.io api key
var SECRET = "";  // gate.io api secret

// System variables - don't change these
var precision = make(map[string]float64)
var data = make(map[string]float64)
var mutex = &sync.Mutex{}
var block = 0


func order(market, side, amount, price string){
    if side == "buy"{
        buy(market, price, amount)
    }
    if side == "sell" {
        sell(market, price, amount)
    }
}

func buy(currencyPair string, rate string, amount string){
	var method string = "POST"
	var url string = "https://api.gateio.co/api2/1/private/buy"
	var param string = "orderType=ioc&currencyPair=" + currencyPair + "&rate=" + rate + "&amount=" + amount
	var ret []byte = httpDo(method,url,param)
	fmt.Println(string(ret))
}

// Place order sell
func sell(currencyPair string, rate string, amount string){
	var method string = "POST"
	var url string = "https://api.gateio.co/api2/1/private/sell"
	var param string = "currencyPair=" + currencyPair + "&rate=" + rate + "&amount=" + amount
	var ret []byte = httpDo(method,url,param)
	fmt.Println(string(ret))
}

func getSign( params string) string {
    key := []byte(SECRET)
    mac := hmac.New(sha512.New, key)
    mac.Write([]byte(params))
    return fmt.Sprintf("%x", mac.Sum(nil))
}
	
/**
*  http request
*/	
func httpDo(method string,url string, param string) []byte {
    client := &http.Client{}
 
    req, err := http.NewRequest(method, url, strings.NewReader(param))
    if err != nil {
        // handle error
    }
    var sign string = getSign(param)
 
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("key", KEY)
    req.Header.Set("sign", sign)
 
    resp, err := client.Do(req)
 
    defer resp.Body.Close()
 
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        // handle error
    }
 	
 	return body;
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func handle_data(message []byte){
    raw_market, _, _, _  := jsonparser.Get(message, "params", "[2]")
    market := string(raw_market)
    
    jsonparser.ArrayEach(message, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        raw_price, _, _, _ := jsonparser.Get(value, "[0]")
        raw_size, _, _, _ := jsonparser.Get(value, "[1]")
        price, _ := strconv.ParseFloat(string(raw_price), 64)
        size, _ := strconv.ParseFloat(string(raw_size), 64)
        if size != 0.0{
            mutex.Lock()
            data[market+"ap"] = price
            data[market+"as"] = size
            mutex.Unlock()
        } 
        
    }, "params", "[1]", "asks")
    
    jsonparser.ArrayEach(message, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        raw_price, _, _, _ := jsonparser.Get(value, "[0]")
        raw_size, _, _, _ := jsonparser.Get(value, "[1]")
        price, _ := strconv.ParseFloat(string(raw_price), 64)
        size, _ := strconv.ParseFloat(string(raw_size), 64)
        if size != 0.0{
            mutex.Lock()
            data[market+"bp"] = price
            data[market+"bs"] = size           
            mutex.Unlock()
        }        
    }, "params", "[1]", "bids")
}

func marketinfo() []byte {
	var method string = "GET"
	var url string = "http://data.gateio.co/api2/1/marketinfo"
	var param string = ""
	var ret []byte = httpDo(method,url,param)
	return ret
}

func get_rates(cur string) {
    price := 0.0
	amount := 0.0
    mutex.Lock()
    r1 := (1/data[cur+"_USDTap"]*data[cur+"_BTCbp"]*data["BTC_USDTbp"])-1 
    am11 := (data[cur+"_USDTas"]*data[cur+"_USDTbp"])
    am12 := (data[cur+"_BTCbs"]*data[cur+"_BTCbp"])*data["BTC_USDTbp"]
    am13 := (data["BTC_USDTbs"]*data["BTC_USDTbp"])
    mutex.Unlock()
    euro_available1 := min(min(am11, am12), am13)
    if 1.0 > r1 && r1 > fee && euro_available1 > min_amount && block == 0{
        block = 1
        t := time.Now()
        mutex.Lock()
        price = data[cur+"_USDTap"]
        amount = floor(min(euro_available1, max_amount)/price, precision[cur+"_USDT"])
        mutex.Unlock()
        go order(cur+"_USDT", "buy", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))   
        mutex.Lock()
        price = data[cur+"_BTCbp"]
        mutex.Unlock()
        amount = floor(amount*0.999, precision[cur+"_BTC"])
        go order(cur+"_BTC", "sell", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))
        mutex.Lock()
        price = data["BTC_USDTbp"]
        amount = floor((amount*0.999)*data[cur+"_BTCbp"], precision["BTC_USDT"])
        mutex.Unlock()
        go order("BTC_USDT", "sell", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))  
        fmt.Println("Time:", time.Now().Sub(t))
        fmt.Println("USDT --->", cur, "---> BTC ---> USDT", r1-fee, euro_available1)
        time.Sleep(time.Second*10)
        block = 0
    }
                 
    mutex.Lock()
    r2 := (1/data["BTC_USDTap"]/data[cur+"_BTCap"]*data[cur+"_USDTbp"])-1 
    am21 := (data["BTC_USDTas"]*data["BTC_USDTbp"])
    am22 := (data[cur+"_BTCas"]*data[cur+"_BTCbp"])*data[cur+"_USDTbp"]
    am23 := (data[cur+"_USDTbs"]*data[cur+"_USDTbp"])
    mutex.Unlock()
    euro_available2 := min(min(am21, am22), am23)
    if 1.0 > r2 && r2 > fee && euro_available2 > min_amount && block == 0{
        block = 1
        t := time.Now() 
        mutex.Lock()
        price = data["BTC_USDTap"]
        amount = floor(min(euro_available2, max_amount)/price, precision["BTC_USDT"])
        mutex.Unlock()
        go order("BTC_USDT", "buy", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))
        mutex.Lock()
        price = data[cur+"_BTCap"]
        amount = floor((amount*0.999)/price, precision[cur+"_BTC"])
        mutex.Unlock()
        go order(cur+"_BTC", "buy", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))
        mutex.Lock()
        price = data[cur+"_USDTbp"]
        mutex.Unlock()
        amount = floor(amount*0.999, precision[cur+"_USDT"])
        go order(cur+"_USDT", "sell", fmt.Sprintf("%f", amount), fmt.Sprintf("%f", price))
        fmt.Println("Time:", time.Now().Sub(t))
        fmt.Println("USDT ---> BTC --->", cur,"---> USDT", r2-fee, euro_available2)  
        time.Sleep(time.Second*10)
        block = 0    
    }  
}

func floor(amount, precision float64) float64{
	return float64(int(amount*(1.0/precision)))/(1.0/precision)
}

func main() {
    fmt.Println("Starting Gate.io TriArb 0.1")
    fmt.Println("Init Data Structures")
    //init the data structs
    mutex.Lock()
    jsonparser.ArrayEach(marketinfo(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        jsonparser.ObjectEach(value, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
            market := strings.ToUpper(string(key))
            x, _, _, _ := jsonparser.Get(value, "min_amount")
            precision[market], _ = strconv.ParseFloat(string(x), 64)
            return nil
        }) 
    },"pairs")    
	data["BTC_USDTap"] = 0.0
	data["BTC_USDTas"] = 0.0
	data["BTC_USDTbp"] = 0.0
	data["BTC_USDTbs"] = 0.0
    for _, cur := range curs {
		data[cur+"_USDTbp"] =  0.0
		data[cur+"_USDTbs"] =  0.0
		data[cur+"_USDTap"] =  0.0
		data[cur+"_USDTas"] =  0.0
		data[cur+"_BTCbp"] = 0.0
		data[cur+"_BTCbs"] =  0.0
		data[cur+"_BTCap"] =  0.0
		data[cur+"_BTCas"] =  0.0
    }
    mutex.Unlock()

    fmt.Println("Establishing Connection")
	c, _, err := websocket.DefaultDialer.Dial("wss://ws.gate.io/v3/", nil)
	if err != nil {
		fmt.Println("dial:", err)
	}
	defer c.Close() // Close Connection when main function ends
	
	payloads := "[\"BTC_USDT\", 1, \"0\"],"
	for _, cur := range curs{
        payloads += "[\""+cur+"_USDT\", 1, \"0\"],"
        payloads += "[\""+cur+"_BTC\", 1, \"0\"],"
	}
	c.WriteMessage(websocket.TextMessage, []byte("{\"id\":234234, \"method\":\"depth.subscribe\", \"params\":["+payloads[:len(payloads)-1]+"]}")) // Subscribe
    if err != nil {
        fmt.Println("write:", err)
        return
    }
    fmt.Println("Ready")
    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            fmt.Println("read:", err)
            return
        }
        go handle_data(message)
        for _, cur := range curs{
            go get_rates(cur)
        }
    }
}
