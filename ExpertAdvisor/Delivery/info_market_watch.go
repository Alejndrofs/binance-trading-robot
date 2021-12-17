package expert

import (
	"Include/go-binance"
	"Include/go-binance/delivery"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type MarketWatch struct {
	Expert   *ExpertAdvisorDelivery
	SocketOn bool

	InitServiceTime time.Time

	StopServices bool

	//delivery.WsAggTradeEvent
	//delivery.WsBookTickerEvent
	delivery.WsContinuousKlineEvent
	//delivery.WsDepthEvent
	delivery.WsIndexPriceEvent
	//delivery.WsKlineEvent
	delivery.WsLiquidationOrderEvent
	delivery.WsMarkPriceEvent
	delivery.WsMarkPriceKlineEvent
	delivery.WsMarketTickerEvent
	delivery.WsPairMarkPriceEvent

	SymbolTrade map[string]*binance.WsTradeEvent
	StopTrade   map[string]bool

	DSymbolAggTrade map[string]*delivery.WsAggTradeEvent
	DStopAggTrade   map[string]bool

	BSymbolAggTrade map[string]*binance.WsAggTradeEvent
	BStopAggTrade   map[string]bool

	DSymbolBookTicker  map[string]*delivery.WsBookTickerEvent
	DStopBookTicker    map[string]bool
	DStopAllBookTicker bool

	BSymbolBookTicker  map[string]*binance.WsBookTickerEvent
	BStopBookTicker    map[string]bool
	BStopAllBookTicker bool

	// delivery.WsContinuousKlineEvent

	AllMarketsStat    binance.WsAllMarketsStatEvent
	SymbolMarketStat  map[string]*binance.WsMarketStatEvent
	StopMarketStat    map[string]bool
	StopAllMarketStat bool

	dSymbolDepth          map[string]*delivery.WsDepthEvent
	dSymbolDepthListenKey string

	BSymbolDepth          map[string]*binance.WsDepthEvent
	bSymbolDepthListenKey string

	StopDepth      map[string]bool
	StopDepth100Ms map[string]bool

	// delivery.WsIndexPriceEvent

	SymbolPartialDepth    map[string]*binance.WsPartialDepthEvent
	StopPartialDepth      map[string]bool
	StopPartialDepth100Ms map[string]bool

	DSymbolKline map[string]map[string]*delivery.WsKlineEvent
	DStopKline   map[string]bool

	BSymbolKline map[string]map[string]*binance.WsKlineEvent
	BStopKline   map[string]bool
}

func NewMarketWatchCrypto(expert *ExpertAdvisorDelivery) *MarketWatch {
	MW := &MarketWatch{
		Expert:                expert,
		SymbolTrade:           make(map[string]*binance.WsTradeEvent),
		StopTrade:             make(map[string]bool),
		SymbolAggTrade:        make(map[string]*delivery.WsAggTradeEvent),
		StopAggTrade:          make(map[string]bool),
		SymbolBookTicker:      make(map[string]*delivery.WsBookTickerEvent),
		StopBookTicker:        make(map[string]bool),
		SymbolMarketStat:      make(map[string]*binance.WsMarketStatEvent),
		StopMarketStat:        make(map[string]bool),
		SymbolDepth:           make(map[string]*delivery.WsDepthEvent),
		StopDepth:             make(map[string]bool),
		StopDepth100Ms:        make(map[string]bool),
		SymbolPartialDepth:    make(map[string]*binance.WsPartialDepthEvent),
		StopPartialDepth:      make(map[string]bool),
		StopPartialDepth100Ms: make(map[string]bool),
		SymbolKline:           make(map[string]map[string]*delivery.WsKlineEvent),
		StopKline:             make(map[string]bool),
	}
	go MW.WsAllMarketStat()
	go MW.RunSockets()
	return MW
}

func (ea *MarketWatch) RunSockets() {
	for {
		if !ea.SocketOn {
			go ea.WsAllMarketStat()
		}
		ea.Expert.UpdateSocketStatus("SOCKET STATUS OFF")
		ea.SocketOn = false
		if time.Since(ea.InitServiceTime) > time.Hour*12 {
			ea.StopServices = true
			time.Sleep(time.Second * 10)
			ea.StopServices = false
			go ea.WsAllMarketStat()
		}
		time.Sleep(time.Second * 3)
	}
}

func (ea *MarketWatch) GetLastPrice(symbol *SymbolInfo) (float64, error) {
	var intentos uint64 = 10
	for intentos > 0 {
		for i := range ea.AllMarketsStat {
			if ea.AllMarketsStat[i].Symbol == symbol.SymbolName && !symbol.Synthetic {
				price, err := strconv.ParseFloat(ea.AllMarketsStat[i].LastPrice, 64)
				if err != nil {
					break
				}
				return price, err
			}
		}
		time.Sleep(time.Microsecond * 250)
		intentos--
	}

	if symbol.Synthetic {
		return 0, nil
	}

	return 0, errors.New("get last price: symbol not found: " + symbol.SymbolName)
}

func (mw *MarketWatch) WsTrade(symbol string) {
	var stop bool
	mw.StopTrade[symbol] = false
	wsTradeHandler := func(event *binance.WsTradeEvent) {
		mw.SymbolTrade[symbol] = event
		if mw.StopTrade[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println("error trade", symbol)
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsTradeServe(symbol, wsTradeHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopTrade, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsAggregateTrade(symbol string) {
	var stop bool
	mw.StopAggTrade[symbol] = false
	wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
		mw.SymbolAggTrade[symbol] = event
		if mw.StopAggTrade[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsAggTradeServe(symbol, wsAggTradeHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopAggTrade, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsBookTicker(symbol string) {
	var stop bool
	mw.StopBookTicker[symbol] = false
	wsBookTickerHandler := func(event *binance.WsBookTickerEvent) {
		mw.SymbolBookTicker[symbol] = event
		if mw.StopBookTicker[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsBookTickerServe(symbol, wsBookTickerHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop || mw.StopAllBookTicker {
				delete(mw.StopBookTicker, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsAllMarketStat() {
	mw.InitServiceTime = time.Now()
	wsMarketStatHandler := func(event binance.WsAllMarketsStatEvent) {
		mw.SocketOn = true
		mw.Expert.UpdateSocketStatus("SOCKET STATUS ON")
		if len(event) > len(mw.AllMarketsStat) {
			mw.AllMarketsStat = event
		} else {
			for i := range mw.AllMarketsStat {
				for j := range event {
					if mw.AllMarketsStat[i].Symbol == event[j].Symbol {
						mw.AllMarketsStat[i] = event[j]
						break
					}
				}
			}
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsAllMarketsStatServe(wsMarketStatHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || mw.StopAllMarketStat {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsMarketStat(symbol string) {
	var stop bool
	mw.StopMarketStat[symbol] = false
	wsMarketStatHandler := func(event *binance.WsMarketStatEvent) {
		mw.SymbolMarketStat[symbol] = event
		if mw.StopMarketStat[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsMarketStatServe(symbol, wsMarketStatHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop || mw.StopAllMarketStat {
				delete(mw.StopMarketStat, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsDepth(symbol string) {
	var stop bool
	mw.StopDepth[symbol] = false
	wsDepthHandler := func(event *binance.WsDepthEvent) {
		mw.SymbolDepth[event.Symbol] = event
		if mw.StopDepth[event.Symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}

	doneC, stopC, err := binance.WsDepthServe(symbol, wsDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopDepth, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsDepth100Ms(symbol string) {
	var stop bool
	mw.StopDepth100Ms[symbol] = false
	wsDepthHandler := func(event *binance.WsDepthEvent) {
		mw.SymbolDepth[symbol] = event
		if mw.StopDepth[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsDepthServe100Ms(symbol, wsDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopDepth100Ms, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsPartialDepth(symbol string, levels string) {
	var stop bool
	mw.StopPartialDepth[symbol] = false
	wsPartialDepthHandler := func(event *binance.WsPartialDepthEvent) {
		mw.SymbolPartialDepth[symbol] = event
		if mw.StopDepth[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsPartialDepthServe(symbol, levels, wsPartialDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopPartialDepth, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsPartialDepth100Ms(symbol string, levels string) {
	var stop bool
	mw.StopPartialDepth100Ms[symbol] = false
	wsPartialDepthHandler := func(event *binance.WsPartialDepthEvent) {
		mw.SymbolPartialDepth[symbol] = event
		if mw.StopDepth[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsPartialDepthServe100Ms(symbol, levels, wsPartialDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopPartialDepth100Ms, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

func (mw *MarketWatch) WsKline(symbol string, timeframe string) {
	var stop bool
	mw.StopKline[symbol] = false
	wsKlineHandler := func(event *binance.WsKlineEvent) {
		if mw.SymbolKline[symbol] == nil {
			mw.SymbolKline[symbol] = make(map[string]*binance.WsKlineEvent)
		}
		mw.SymbolKline[symbol][timeframe] = event
		if mw.StopKline[symbol] {
			stop = true
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsKlineServe(symbol, string(timeframe), wsKlineHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices || stop {
				delete(mw.StopKline, symbol)
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}
