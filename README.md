# TriArb Ops

In this Repo I'm collecting scripts meant to trade triangular arbitrage opportunities on various crypto currency exchanges. The code is written in golang in order to run fast and compete with the other arbitrators out there.
I've hosted both scripts near the corresponding exchange server, but could't make any profits from it. Both, Binance and Gate seem to be overpopulated with Arbitrators - and sadly only the fastest wins. 
Nevertheless - I think this code is a quite good base to start a serious trading operation. If you can afford good hardware and a server located right next to the exchanges, these scripts could (maybe) make you a ton of money.

Note that the code was not meant to be published - thus it's messy and maybe hard to read. Nevertheless you should understand what it's doing before you're using it.


**Use at your own risk. I'm not responsible for any losses and/or  damage caused by this code**

**This code should only be used if you are familiar with go and understand what it is doing - it may be buggy and you could loose your money**




# How it works

The script tracks the prices of each altcoin on the USDT **and**  the BTC market. If it detects a difference, greater than the cumulated trading fees, it buys the altcoin on the cheaper market and sells on the more expensive. Below both possibilities are visualized:

![Trinangular Arbitrage](https://github.com/georgk10/BinanceTriArb/raw/master/TriArb.PNG)

Trying to be fast, a goroutine is started for each order. If we would wait each trade to finish before making the other ones, we would be way to slow compared to other players in the game. 
That's why you need to have at least a predefined amount (in USDT) of each currency you plan to trade.
That said, you're also exposed to the risk of a crash of each of these currencies - choose them wisely.

Here's a quick example:<br/>
You have 10 USDT, XX ETH (equals to 10 USDT) & YY BTC (equals to 10 USDT).<br/>
Now the script detects an arbitrage opportunity with earning of 0.1% and sends 3 order (almost concurrently).<br/>
After these trades happened, you'll still have XX ETH and YY BTC, but 10.01 USDT - WOW it's magic!!!<br/>
        
## Why??

Just because...

No just kidding. If you're wondering why I'm publishing this - it's simly because it's not profitable to run. I just couldn't beat the other guys in terms of speed, so I decided to publish the code. Now I'm focusing my work on smaller exchanges, where it's actually possible to make some money without too many expenses. 
Maybe anyone of you guys can use these scripts to make some money - otherwise the hours I spent on this would be a complete waste of time...

## Dependencies

This code depends on (for gates.io)<br/>
[github.com/gorilla/websocket](https://github.com/gorilla/websocket) <br/>
[github.com/buger/jsonparser](https://github.com/buger/jsonparser)<br/>

And (for binance.com)<br/>
[github.com/adshao/go-binance](https://github.com/adshao/go-binance)<br/>


Just type <br/>
>go get github.com/gorilla/websocket<br/>
>go get github.com/buger/jsonparser<br/>
>go get github.com/adshao/go-binance<br/>

## Setup

1. Download this repo   
2. Install dependencies   
3. Open the codebase   
4. Read and understand the code!!!
5. Edit the variables marked with "Custom Variables"
6. Make sure you have at least "min_amount" usdt of **each** currency you want to trade
7. Build & run the script (go build scriptname.go)


##

I would love to know if you could make some profits out of this. For any questions, inquiries or suggestions, contact me at:
 georgk98@gmail.com
