package futures

import (
	"Include/go-binance/futures"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type MarketWatchFutures struct {
	Expert						*ExpertAdvisorFutures
	SocketOn					bool
	InitServiceTime				time.Time
	StopServices      			bool

	AggTradeEvent 				map[string]*futures.WsAggTradeEvent
	MarkPriceEvent				map[string]*futures.WsMarkPriceEvent
	AllMarkPriceEvent			futures.WsAllMarkPriceEvent
	KlineEvent					map[string]*futures.WsKlineEvent
	MiniMarketTickerEvent		map[string]*futures.WsMiniMarketTickerEvent
	AllMiniMarketTickerEvent 	*futures.WsAllMiniMarketTickerEvent
	MarketTickerEvent			map[string]*futures.WsMarketTickerEvent	
	AllMarketTickerEvent		*futures.WsAllMarketTickerEvent
	BookTickerEvent				map[string]*futures.WsBookTickerEvent
	LiquidationOrderEvent		map[string]*futures.WsLiquidationOrderEvent
	DepthEvent					map[string]*futures.WsDepthEvent
	BLVTInfoEvent				map[string]*futures.WsBLVTInfoEvent
	BLVTKlineEvent				map[string]*futures.WsBLVTKlineEvent
	CompositeIndexEvent			map[string]*futures.WsCompositeIndexEvent
	UserDataEvent				*futures.WsUserDataEvent
}

func NewMarketWatchCrypto(expert *ExpertAdvisorFutures) *MarketWatchFutures {
	MW := &MarketWatchFutures{
		Expert: expert,
		SocketOn: false,
		InitServiceTime: time.Now(),
		StopServices: false,
		AggTradeEvent: make(map[string]*futures.WsAggTradeEvent),
		MarkPriceEvent: make(map[string]*futures.WsMarkPriceEvent),
		KlineEvent: make(map[string]*futures.WsKlineEvent),
		MiniMarketTickerEvent: make(map[string]*futures.WsMiniMarketTickerEvent),
		MarketTickerEvent: make(map[string]*futures.WsMarketTickerEvent),
		BookTickerEvent: make(map[string]*futures.WsBookTickerEvent),
		LiquidationOrderEvent: make(map[string]*futures.WsLiquidationOrderEvent),
		DepthEvent: make(map[string]*futures.WsDepthEvent),
		BLVTInfoEvent: make(map[string]*futures.WsBLVTInfoEvent),
		BLVTKlineEvent: make(map[string]*futures.WsBLVTKlineEvent),
		CompositeIndexEvent: make(map[string]*futures.WsCompositeIndexEvent),
	}
	go MW.WsAllMarkPrice()
	go MW.RunSockets()
	return MW
}

func (ea *MarketWatchFutures) RunSockets() {
	for{
		if !ea.SocketOn {
			go ea.WsAllMarkPrice()
		}
		ea.Expert.UpdateSocketStatus("SOCKET STATUS OFF")
		ea.SocketOn = false
		if time.Since(ea.InitServiceTime) > time.Hour*12 {
			ea.StopServices = true
			time.Sleep(time.Second*10)
			ea.StopServices = false
			go ea.WsAllMarkPrice()
		} 
		time.Sleep(time.Minute*1)
	}
}

func (ea *MarketWatchFutures) GetLastPrice(symbol *SymbolInfo) (float64,error) {
	var intentos uint64 = 10
	for intentos > 0 {
		for i := range ea.AllMarkPriceEvent {
			if ea.AllMarkPriceEvent[i].Symbol == symbol.SymbolName && !symbol.Synthetic {
				price, err := strconv.ParseFloat(ea.AllMarkPriceEvent[i].MarkPrice,64)
				if err != nil {
					break
				}
				return price, err 
			}
		}
		time.Sleep(time.Microsecond*250)
		intentos--
	}

	if symbol.Synthetic {
		return 1, nil
	}

	return 0, errors.New("get last price: symbol not found: "+symbol.SymbolName)
}

//+----------------------------------------------------------------------------+
//|
//+----------------------------------------------------------------------------+

// AggTradeEvent 				*futures.WsAggTradeEvent
func(mw *MarketWatchFutures) WsAggTrade (symbol string) {
	wsAggTradeHandler := func (event *futures.WsAggTradeEvent)  {
		mw.AggTradeEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error aggTrade")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsAggTradeServe(symbol, wsAggTradeHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}


// MarkPrice					*futures.WsMarkPriceEvent
func(mw *MarketWatchFutures) WsMarkPrice (symbol string) {
	wsMarkPriceEvent := func (event *futures.WsMarkPriceEvent)  {
		mw.MarkPriceEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error MarkPrice")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsMarkPriceServe(symbol, wsMarkPriceEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// AllMarkPrice					*futures.WsAllMarkPriceEvent
func(mw *MarketWatchFutures) WsAllMarkPrice () {
	wsAllMarkPriceEvent := func (event futures.WsAllMarkPriceEvent)  {
		mw.AllMarkPriceEvent = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMarkPrice")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsAllMarkPriceServe(wsAllMarkPriceEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// Kline					*futures.WsKlineEvent
func(mw *MarketWatchFutures) WsKline (symbol string, timeframe ENUM_TIMEFRAME) {
	wsKlineEvent := func (event *futures.WsKlineEvent)  {
		mw.KlineEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error Kline")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsKlineServe(symbol,string(timeframe),wsKlineEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// MiniMarketTicker					*futures.WsMiniMarketTicker
func(mw *MarketWatchFutures) WsMiniMarketTicker (symbol string) {
	wsMiniMarketTicker := func (event *futures.WsMiniMarketTickerEvent)  {
		mw.MiniMarketTickerEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error MiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsMiniMarketTickerServe(symbol,wsMiniMarketTicker, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// AllMiniMarketTicker					*futures.WsAllMiniMarketTicker
func(mw *MarketWatchFutures) WsAllMiniMarketTicker () {
	wsAllMiniMarketTicker := func (event futures.WsAllMiniMarketTickerEvent)  {
		mw.AllMiniMarketTickerEvent = &event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsAllMiniMarketTickerServe(wsAllMiniMarketTicker, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// MarketTickerEvent					*futures.WsMarketTickerEvent
func(mw *MarketWatchFutures) WsMarketTicker (symbol string) {
	wsMarketTickerEvent := func (event *futures.WsMarketTickerEvent)  {
		mw.MarketTickerEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsMarketTickerServe(symbol,wsMarketTickerEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// AllMarketTickerEvent					*futures.WsAllMarketTickerEvent
func(mw *MarketWatchFutures) WsAllMarketTicker () {
	wsAllMarketTickerEvent := func (event futures.WsAllMarketTickerEvent)  {
		mw.AllMarketTickerEvent = &event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsAllMarketTickerServe(wsAllMarketTickerEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

//BookTickerEvent 						*futures.WsBookTickerEvent
func(mw *MarketWatchFutures) WsBookTicker (symbol string) {
	wsBookTickerEvent := func (event *futures.WsBookTickerEvent)  {
		mw.BookTickerEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsBookTickerServe(symbol,wsBookTickerEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// LiquidationOrderEvent				*futures.WsLiquidationOrderEvent
func(mw *MarketWatchFutures) WsLiquidationOrderEvent (symbol string) {
	wsLiquidationOrderEvent := func (event *futures.WsLiquidationOrderEvent)  {
		mw.LiquidationOrderEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsLiquidationOrderServe(symbol,wsLiquidationOrderEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// DepthEvent							*futures.WsDepthEvent
func(mw *MarketWatchFutures) WsDepth (symbol string) {
	wsDepthEvent := func (event *futures.WsDepthEvent)  {
		mw.DepthEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsDiffDepthServe(symbol,wsDepthEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// BLVTInfoEvent						*futures.WsBLVTInfoEvent
func(mw *MarketWatchFutures) WsBLVTInfo (symbol string) {
	wsBLVTInfoEvent := func (event *futures.WsBLVTInfoEvent)  {
		mw.BLVTInfoEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsBLVTInfoServe(symbol,wsBLVTInfoEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// BLVTKlineEvent						*futures.WsBLVTKlineEvent
func(mw *MarketWatchFutures) WsBLVTKline (symbol string,timefreame ENUM_TIMEFRAME) {
	wsBLVTKlineEvent := func (event *futures.WsBLVTKlineEvent)  {
		mw.BLVTKlineEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsBLVTKlineServe(symbol,string(timefreame),wsBLVTKlineEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// CompositeIndexEvent					*futures.WsCompositeIndexEvent
func(mw *MarketWatchFutures) WsCompositeIndex (symbol string) {
	wsCompositeIndexEvent := func (event *futures.WsCompositeIndexEvent)  {
		mw.CompositeIndexEvent[symbol] = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsCompositiveIndexServe(symbol,wsCompositeIndexEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

// UserDataEvent						*futures.WsUserDataEvent
func(mw *MarketWatchFutures) WsUserData (listenKey string) {
	wsUserDataEvent := func (event *futures.WsUserDataEvent)  {
		mw.UserDataEvent = event 
	}
	errHandler := func(err error) {
		fmt.Println("error AllMiniMarketTicker")
		fmt.Println(err)
	}
	doneC, stopC, err := futures.WsUserDataServe(listenKey,wsUserDataEvent, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		for {
			if mw.StopServices {
				stopC <- struct{}{}
			}
			time.Sleep(time.Second * 5)
		}
	}()
	<-doneC
}

//+----------------------------------------------------------------------------+
//|
//+----------------------------------------------------------------------------+