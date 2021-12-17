package futures

import (
	FileManager "Include/File"
	"Include/go-binance/futures"
	"errors"
	"context"
	"fmt"
	"time"
)

// https://github.com/adshao/go-binance

var BUY string = "██████╗ ██╗   ██╗██╗   ██╗\n██╔══██╗██║   ██║╚██╗ ██╔╝\n██████╔╝██║   ██║ ╚████╔╝ \n██╔══██╗██║   ██║  ╚██╔╝  \n██████╔╝╚██████╔╝   ██║   \n╚═════╝  ╚═════╝    ╚═╝   \n"
var SELL string = "███████╗███████╗██╗     ██╗     \n██╔════╝██╔════╝██║     ██║     \n███████╗█████╗  ██║     ██║     \n╚════██║██╔══╝  ██║     ██║     \n███████║███████╗███████╗███████╗\n╚══════╝╚══════╝╚══════╝╚══════╝\n"
var Space string = "//+--------------------------------------------------------+"
var Banner string = " ██████╗ ██████╗ ██╗██████╗ \n██╔════╝ ██╔══██╗██║██╔══██╗\n██║  ███╗██████╔╝██║██║  ██║\n██║   ██║██╔══██╗██║██║  ██║\n╚██████╔╝██║  ██║██║██████╔╝\n ╚═════╝ ╚═╝  ╚═╝╚═╝╚═════╝ \n"

type ENUM_STRATEGY string

const (
	STRATEGY_GRID 	ENUM_STRATEGY = "grid"
	STRATEGY_2MA	ENUM_STRATEGY = "2MA"
	STRATEGY_ARB	ENUM_STRATEGY = "3ARB"
)

type ENUM_TIMEFRAME string

const (
	TIMEFRAME_M1  ENUM_TIMEFRAME = "1m"
	TIMEFRAME_M3  ENUM_TIMEFRAME = "3m"
	TIMEFRAME_M5  ENUM_TIMEFRAME = "5m"
	TIMEFRAME_M15 ENUM_TIMEFRAME = "15m"
	TIMEFRAME_M30 ENUM_TIMEFRAME = "30m"
	TIMEFRAME_H1  ENUM_TIMEFRAME = "1h"
	TIMEFRAME_H2  ENUM_TIMEFRAME = "2h"
	TIMEFRAME_H4  ENUM_TIMEFRAME = "4h"
	TIMEFRAME_H6  ENUM_TIMEFRAME = "6h"
	TIMEFRAME_H8  ENUM_TIMEFRAME = "8h"
	TIMEFRAME_H12 ENUM_TIMEFRAME = "12h"
	TIMEFRAME_D1  ENUM_TIMEFRAME = "1d"
	TIMEFRAME_D3  ENUM_TIMEFRAME = "3d"
	TIMEFRAME_W1  ENUM_TIMEFRAME = "1w"
	TIMEFRAME_1M  ENUM_TIMEFRAME = "1M"
)

type ENUM_MONEY_MANAGEMENT_TYPE string

const (
	MONEY_MANAGEMENT_FIX     ENUM_MONEY_MANAGEMENT_TYPE = "fix"
	MONEY_MANAGEMENT_DYNAMIC ENUM_MONEY_MANAGEMENT_TYPE = "dynamic"
)

type ExpertAdvisorFutures struct {
	Debug        	bool
	Client       	*futures.Client
	Exchange     	*ExchangeInfo
	Account      	*AccountInfo
	MWFutures     	*MarketWatchFutures
	TradeManager 	*ConversionManager
	File         	*FileManager.FileManager
	OrdersLog    	string
	ErrorsLog    	string
	PositionsID		[]int64
	OrdersID		[]int64

	MainAsset              string
	MoneyManagemenType     ENUM_MONEY_MANAGEMENT_TYPE
	FixTradeVolume         float64
	DynamicTradePercentage float64

	Strategy	ENUM_STRATEGY
	Grid 		*GridGraphStrategy
}

func NewExpertAdvisor(key, secret string) *ExpertAdvisorFutures {
	EA := &ExpertAdvisorFutures{
		Client:     futures.NewClient(key, secret),
		File:       new(FileManager.FileManager),
		Debug:      false,
		PositionsID: make([]int64, 0),
		OrdersID: make([]int64, 0),
	}
	fmt.Println("Loading exchange info")
	EA.Exchange = NewExchangeInfo(EA)
	fmt.Println("Loading account info")
	EA.Account = NewAccountInfo(EA)
	EA.Account.UpdateSymbolsInfo()
	fmt.Println("Loading market watch")
	EA.MWFutures = NewMarketWatchCrypto(EA)
	fmt.Println("Loading conversion manager")
	EA.TradeManager = NewConversionManager(EA)
	fmt.Println("Money Management:")
	EA.GetMoneyManagement()
	fmt.Println("Select Strategy:")
	EA.Grid = NewGridGraph(EA)

	EA.Run()
	return EA
}

func (ea *ExpertAdvisorFutures) Run() {
	fmt.Println(Banner)
	updateTime := time.Now()
	for {
		if time.Since(updateTime) > time.Minute*5 {
			ea.Account.UpdateSymbolsInfo()
			ea.Account.UpdateBalances()
			updateTime = time.Now()
		}
		ea.Grid.Ontimer()
	}
}

func (ea *ExpertAdvisorFutures) GetMoneyManagement() {
	for {
		fmt.Print(">>> select main asset: ")
		fmt.Scanln(&ea.MainAsset)
		if ea.Account.AssetExists(ea.MainAsset) {
			break
		} else {
			fmt.Println(">>> invalid asset!")
		}
	}
	var selection string
	for {
		fmt.Println("Select Static volume (s)")
		fmt.Println("Select Dynamic volume (d)")
		fmt.Print(">>> ")
		fmt.Scanln(&selection)
		if selection == "s" {
			ea.MoneyManagemenType = MONEY_MANAGEMENT_FIX
			break
		} else if selection == "d" {
			ea.MoneyManagemenType = MONEY_MANAGEMENT_DYNAMIC
			break
		}
	}
	if ea.MoneyManagemenType == MONEY_MANAGEMENT_FIX {
		for {
			fmt.Print(">>> select trade volume (", ea.MainAsset, "): ")
			fmt.Scanln(&ea.FixTradeVolume)
			if ea.FixTradeVolume > 0 {
				break
			} else {
				fmt.Println("Value is lower than the minimum common lot of all symbols. try again.")
			}
		}
	} else {
		for {
			fmt.Print(">>> select trade volume as a percentage of the account value (from 0.0 to 0.99): ")
			fmt.Scanln(&ea.DynamicTradePercentage)
			if ea.DynamicTradePercentage > 0 && ea.DynamicTradePercentage < 1 {
				break
			} else {
				fmt.Println("Value is lower than zero or higher than one. try again.")
			}
		}
	}
}

func (ea *ExpertAdvisorFutures) UpdateOrdersLog(log string) {
	banner := "//+--------------------------------------------------------+\n"
	log = log + "\n" + banner
	ea.OrdersLog = log + ea.OrdersLog
	ea.File.WriteText("orders.md", ea.OrdersLog)
}

func (ea *ExpertAdvisorFutures) UpdateSocketStatus(log string) {
	ea.File.WriteText("socket.md", log)
}

func (ea *ExpertAdvisorFutures) UpdateErrorsLog(log string) {
	banner := "//+--------------------------------------------------------+\n"
	ea.ErrorsLog = time.Now().String() + "\n" + log + "\n" + banner + ea.ErrorsLog
	ea.File.WriteText("errors.md", ea.ErrorsLog)
}

func (ea *ExpertAdvisorFutures) OpenPosition(symbol string, side futures.ENUM_SIDE_TYPE, quantity float64) int64 {
	order, err := ea.Client.NewCreateOrderService().Symbol(symbol).Side(side).Type(futures.ORDER_TYPE_MARKET).TimeInForce(futures.TimeInForceTypeGTC).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return -1
	} else {
		ea.PositionsID = append(ea.PositionsID,order.OrderID)
		return order.OrderID
	}
}

func (ea *ExpertAdvisorFutures) GetPosition(symbol string, ID int64) (*futures.Order,error) {
	order, err := ea.Client.NewGetOrderService().Symbol(symbol).OrderID(ID).Do(context.Background())
	if err != nil {
    	fmt.Println(err)
   		return &futures.Order{}, errors.New("position not found")
	}
	return order, nil
}

func (ea *ExpertAdvisorFutures) OpenOrder(symbol string, side futures.ENUM_SIDE_TYPE, quantity float64) int64 {
	order, err := ea.Client.NewCreateOrderService().Symbol(symbol).Side(side).Type(futures.ORDER_TYPE_MARKET).TimeInForce(futures.TimeInForceTypeGTC).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return -1
	} else {
		ea.OrdersID = append(ea.OrdersID,order.OrderID)
		return order.OrderID
	}
}

func (ea *ExpertAdvisorFutures) CancelOrder(symbol string, ID int64) (*futures.Order,error) {
	order, err := ea.Client.NewGetOrderService().Symbol(symbol).OrderID(ID).Do(context.Background())
	if err != nil {
    	fmt.Println(err)
   		return &futures.Order{}, errors.New("position not found")
	}
	return order, nil
}

/*

func (ea *ExpertAdvisorFutures) PlaceMarketOrder(symbol string, side ENUM_ORDER_TYPE, quantity float64) (bool, error) {
	marketOrder := binanceTrade.MarketOrder{
		Symbol:     symbol,
		Side:       string(side),
		Type:       "MARKET",
		Quantity:   quantity,
		RecvWindow: 5000,
	}
	_, err := ea.BrokerInfo.PlaceMarketOrder(marketOrder)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	return true, nil
}

func (ea *ExpertAdvisorFutures) PlaceLimitOrder(symbol string, side ENUM_ORDER_TYPE, quantity float64, price float64, recvWindow int64) (bool, error) {
	limitOrder := binanceTrade.LimitOrder{
		Symbol:      symbol,
		Side:        string(side),
		Type:        "LIMIT",
		TimeInForce: "GTC",
		Quantity:    quantity,
		Price:       price,
		RecvWindow:  5000,
	}
	_, err := ea.BrokerInfo.PlaceLimitOrder(limitOrder)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	return true, nil
}

func (ea *ExpertAdvisorFutures) CancelOrder(symbol string, orderId int64) (bool, error) {
	orderQuery := binanceTrade.OrderQuery{
		Symbol:     symbol,
		OrderId:    orderId,
		RecvWindow: 5000,
	}
	_, err := ea.BrokerInfo.CancelOrder(orderQuery)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	return true, nil
}

func (ea *ExpertAdvisorFutures) CheckOrder(symbol string, orderId int64) (status binanceTrade.OrderStatus, err error) {
	orderQuery := binanceTrade.OrderQuery{
		Symbol:     symbol,
		OrderId:    orderId,
		RecvWindow: 5000,
	}
	return ea.BrokerInfo.CheckOrder(orderQuery)
}

func (ea *ExpertAdvisorFutures) GetAllOpenOrders() (orders []binanceTrade.OrderStatus, err error) {
	return ea.BrokerInfo.GetAllOpenOrders()
}

func (ea *ExpertAdvisorFutures) GetOpenOrders(symbol string) (orders []binanceTrade.OrderStatus, err error) {
	openOrdersQuery := binanceTrade.OpenOrdersQuery{
		Symbol:     symbol,
		RecvWindow: 5000,
	}
	return ea.BrokerInfo.GetOpenOrders(openOrdersQuery)
}

func (ea *ExpertAdvisorFutures) GetAllOrders(symbol string, orderId int64) (orders []binanceTrade.OrderStatus, err error) {
	allOrdersQuery := binanceTrade.AllOrdersQuery{
		Symbol:     symbol,
		OrderId:    orderId,
		Limit:      500,
		RecvWindow: 5000,
	}
	return ea.BrokerInfo.GetAllOrders(allOrdersQuery)
}

func (ea *ExpertAdvisorFutures) GetTrades(symbol string) (trades []binanceTrade.Trade, err error) {
	return ea.BrokerInfo.GetTrades(symbol)
}

func (ea *ExpertAdvisorFutures) GetTradesFromOrder(symbol string, id int64) (matchingTrades []binanceTrade.Trade, err error) {
	return ea.BrokerInfo.GetTradesFromOrder(symbol, id)
}

func (ea *ExpertAdvisorFutures) GetWithdrawHistory() (withdraws binanceTrade.WithdrawList, err error) {
	return ea.BrokerInfo.GetWithdrawHistory()
}

func (ea *ExpertAdvisorFutures) GetDepositHistory() (deposits binanceTrade.DepositList, err error) {
	return ea.BrokerInfo.GetDepositHistory()
}

func (ea *ExpertAdvisorFutures) GetOrderBook(symbol string) (book binanceTrade.OrderBook, err error) {
	orderBookQuery := binanceTrade.OrderBookQuery{
		Symbol: symbol,
		Limit:  100,
	}
	return ea.BrokerInfo.GetOrderBook(orderBookQuery)
}

func (ea *ExpertAdvisorFutures) GetAggTrades(symbol string) (trades []binanceTrade.AggTrade, err error) {
	symbolQuery := binanceTrade.SymbolQuery{
		Symbol: symbol,
	}
	return ea.BrokerInfo.GetAggTrades(symbolQuery)
}

func (ea *ExpertAdvisorFutures) GetKlines(symbol string, interval ENUM_TIMEFRAME) (klines []binanceTrade.Kline, err error) {
	klineQuery := binanceTrade.KlineQuery{
		Symbol:   symbol,
		Interval: string(interval),
		Limit:    500,
	}
	return ea.BrokerInfo.GetKlines(klineQuery)
}

func (ea *ExpertAdvisorFutures) Get24Hr(symbol string) (changeStats binanceTrade.ChangeStats, err error) {
	symbolQuery := binanceTrade.SymbolQuery{
		Symbol: symbol,
	}
	return ea.BrokerInfo.Get24Hr(symbolQuery)
}

func (ea *ExpertAdvisorFutures) Ping() (pingResponse binanceTrade.PingResponse, err error) {
	return ea.BrokerInfo.Ping()
}

func (ea *ExpertAdvisorFutures) GetWithdrawalSystemStatus() (withdrawalSystemStatus binanceTrade.WithdrawalSystemStatus, err error) {
	return ea.BrokerInfo.GetWithdrawalSystemStatus()
}
*/