package main

import (
    "github.com/adshao/go-binance"
    "fmt"
    "strconv"
    "context"
    "sync"
    "time"
)

// Custom Variables - change these
var curs = [...]string{"WAN", "STORM", "ANKR"}//"ETH", "LTC", "BNB", "NEO", "BCC", "QTUM", "OMG", "ZRX", "FUN", "IOTA", "LINK", "MTL", "EOS", "ETC", "ZEC", "DASH", "TRX", "XRP", "ENJ", "VEN", "NULS", "XMR", "BAT", "ADA", "XLM", "WAVES", "GTO", "ICX", "IOST", "NANO", "ZIL", "ONT", "STORM", "WAN", "TUSD", "CVC", "THETA", "NPXS", "KEY", "MFT", "DENT", "HOT", "VET", "DOCK", "PAX", "USDC", "MITH", "BCHABC", "BCHSV", "BTT", "ONG", "FET", "CELR", "MATIC", "ATOM", "TFUEL", "ONE", "FTM", "ALGO", "ERD", "DOGE", "DUSK", "ANKR", "WIN", "COS", "COCOS", "TOMO", "PERL"}
var min_amount = 15.0 // Minimum amount per trade (in USDT) (Shouldn't be less than 15)
var max_amount = 25.0 // Maximum amount per trade (in USDT) (This is also the amount you need to have in EACH currency you trade)
var fee = 0.003 // ~ cumulated fee for 3 trades
var apiKey = "" // binance api key
var	secretKey = "" // binance secret key

// System variables - don't change these
var data = make(map[string]float64)
var precision map[string]float64
var client = binance.NewClient(apiKey, secretKey)
var fail = 0
var mutex = &sync.Mutex{}
var block = 0

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func get_max(bids []binance.Bid) (float64, float64){
    max_bid := float64(0);
    qty := float64(0);
    for _, element := range bids {
        thisprice, _ := strconv.ParseFloat(element.Price, 64)
        thisqty, _ := strconv.ParseFloat(element.Quantity, 64)
        if max_bid < thisprice{
            max_bid = thisprice
            qty = thisqty
        }
    }
    return max_bid, qty
}

func order(market, side, amount string){
    if side == "buy"{
        client.NewCreateOrderService().Symbol(market).
            Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
            Quantity(amount).
            Do(context.Background())
    }
    if side == "sell" {
        client.NewCreateOrderService().Symbol(market).
            Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
            Quantity(amount).
            Do(context.Background())        
    }
}

func get_min(asks []binance.Ask) (float64, float64){
    min_ask := float64(999999999999);
    qty := float64(0);
    for _, element := range asks {
        thisprice, _ := strconv.ParseFloat(element.Price, 64)
        thisqty, _ := strconv.ParseFloat(element.Quantity, 64)
        if min_ask > thisprice{
            min_ask = thisprice
            qty = thisqty
        }
    }
    return min_ask, qty
}

func minqrt() (map[string]float64) {
	var symbolinfos map[string]float64
	symbolinfos = make(map[string]float64)
	info,err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil{
		return nil
	}
	for _,each:= range info.Symbols{
		symbolinfos[each.Symbol], _ = strconv.ParseFloat(each.Filters[2]["minQty"].(string), 64)
	}
	return symbolinfos
}

func floor(amount, precision float64) float64{
	return float64(int(amount*(1.0/precision)))/(1.0/precision)
}

func get_rates(cur string) {
	amount := 0.0
    mutex.Lock()
    r1 := (1/data[cur+"USDTap"]*data[cur+"BTCbp"]*data["BTCUSDTbp"])-1 
    am11 := (data[cur+"USDTas"]*data[cur+"USDTbp"])
    am12 := (data[cur+"BTCbs"]*data[cur+"BTCbp"])*data["BTCUSDTbp"]
    am13 := (data["BTCUSDTbs"]*data["BTCUSDTbp"])
    mutex.Unlock()
    euro_available1 := min(min(am11, am12), am13)
    if 1.0 > r1 && r1 > fee && euro_available1 > min_amount && block == 0{
        block = 1
        t := time.Now()
        mutex.Lock()
        amount = floor(min(euro_available1, max_amount)/data[cur+"USDTap"], precision[cur+"USDT"])
        mutex.Unlock()
        go order(cur+"USDT", "buy", fmt.Sprintf("%f", amount))
        // mutex lock not needed
        amount = floor(amount*0.999, precision[cur+"BTC"])
        // mutex unlock not needed
        go order(cur+"BTC", "sell", fmt.Sprintf("%f", amount))
        mutex.Lock()
        amount = floor((amount*0.999)*data[cur+"BTCbp"], precision["BTCUSDT"])
        mutex.Unlock()
        go order("BTCUSDT", "sell", fmt.Sprintf("%f", amount))      
        fmt.Println("Time:", time.Now().Sub(t))
        fmt.Println("USDT --->", cur, "---> BTC ---> USDT", r1-fee, euro_available1)
    }
                 
    mutex.Lock()
    r2 := (1/data["BTCUSDTap"]/data[cur+"BTCap"]*data[cur+"USDTbp"])-1 
    am21 := (data["BTCUSDTas"]*data["BTCUSDTbp"])
    am22 := (data[cur+"BTCas"]*data[cur+"BTCbp"])*data[cur+"USDTbp"]
    am23 := (data[cur+"USDTbs"]*data[cur+"USDTbp"])
    mutex.Unlock()
    euro_available2 := min(min(am21, am22), am23)
    if 1.0 > r2 && r2 > fee && euro_available2 > min_amount && block == 0{
        block = 1
        t := time.Now() 
        mutex.Lock()
        amount = floor(min(euro_available2, max_amount)/data["BTCUSDTap"], precision["BTCUSDT"])
        mutex.Unlock()
        go order("BTCUSDT", "buy", fmt.Sprintf("%f", amount))
        mutex.Lock()
        amount = floor((amount*0.999)/data[cur+"BTCap"], precision[cur+"BTC"])
        mutex.Unlock()
        go order(cur+"BTC", "buy", fmt.Sprintf("%f", amount))
        // mutex lock not needed
        amount = floor(amount*0.999, precision[cur+"USDT"])
        // mutex unlock not needed
        go order(cur+"USDT", "sell", fmt.Sprintf("%f", amount))
        fmt.Println("Time:", time.Now().Sub(t))
        fmt.Println("USDT ---> BTC --->", cur,"---> USDT", r2-fee, euro_available2)      
    }  
}

func get_balance(cur string) string{
	bal := "0"
    res, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return bal
	}
	for _, balance := range res.Balances {
		if balance.Asset == cur{
			bal = balance.Free
		}
	}
    return bal;
}

func handle_data(message *binance.WsPartialDepthEvent) {
    bp, bs := get_max(message.Bids)
    ap, as := get_min(message.Asks)
    mutex.Lock()
    data[message.Symbol+"ap"] = ap
    data[message.Symbol+"as"] = as
    data[message.Symbol+"bp"] = bp
    data[message.Symbol+"bs"] = bs
    mutex.Unlock()
}

func main() {
	fmt.Println("Binance Arbitrator - 'BArbit 0.8'")
	
	// get precision
	precision = minqrt()
	
    //init the data struct
	data["BTCUSDTap"] = 0.0
	data["BTCUSDTas"] = 0.0
	data["BTCUSDTbp"] = 0.0
	data["BTCUSDTbs"] = 0.0
    for _, cur := range curs {
		data[cur+"USDTbp"] =  0.0
		data[cur+"USDTbs"] =  0.0
		data[cur+"USDTap"] =  0.0
		data[cur+"USDTas"] =  0.0
		data[cur+"BTCbp"] = 0.0
		data[cur+"BTCbs"] =  0.0
		data[cur+"BTCap"] =  0.0
		data[cur+"BTCas"] =  0.0
    }
    
    // Create Data Handlers & connect to websocket
    datahandler := func(event *binance.WsPartialDepthEvent) {
        go handle_data(event)
        for _, cur := range curs{
            go get_rates(cur)
        }
    }
    errHandler := func(err error) {
        fmt.Println(err)
    }
    
    curmap := make(map[string]string)
	curmap["BTCUSDT"] = "5"
    for _, cur := range curs {
		curmap[cur+"USDT"] = "5"
		curmap[cur+"BTC"] = "5"
    }
    binance.WsCombinedPartialDepthServe(curmap, datahandler, errHandler)
	
	// Main Loop
	sum := 1
	for sum > 0 {
        time.Sleep(time.Second)
		if block == 1 {
			time.Sleep(time.Second*5)
			block = 0			
		}		
	}
}
