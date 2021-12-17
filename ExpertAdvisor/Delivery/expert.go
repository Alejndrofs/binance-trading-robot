package expert

import (
	FileManager "Include/File"
	"Include/go-binance/delivery"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

// https://github.com/adshao/go-binance

var BUY string = "██████╗ ██╗   ██╗██╗   ██╗\n██╔══██╗██║   ██║╚██╗ ██╔╝\n██████╔╝██║   ██║ ╚████╔╝ \n██╔══██╗██║   ██║  ╚██╔╝  \n██████╔╝╚██████╔╝   ██║   \n╚═════╝  ╚═════╝    ╚═╝   \n"
var SELL string = "███████╗███████╗██╗     ██╗     \n██╔════╝██╔════╝██║     ██║     \n███████╗█████╗  ██║     ██║     \n╚════██║██╔══╝  ██║     ██║     \n███████║███████╗███████╗███████╗\n╚══════╝╚══════╝╚══════╝╚══════╝\n"
var Space string = "//+--------------------------------------------------------+"
var Banner string = " ██████╗ ██████╗ ██╗██████╗ \n██╔════╝ ██╔══██╗██║██╔══██╗\n██║  ███╗██████╔╝██║██║  ██║\n██║   ██║██╔══██╗██║██║  ██║\n╚██████╔╝██║  ██║██║██████╔╝\n ╚═════╝ ╚═╝  ╚═╝╚═╝╚═════╝ \n"

type AssetList struct {
	List []string `json:"assets"`
}

var Assets AssetList

type ENUM_STRATEGY string

const (
	STRATEGY_GRID ENUM_STRATEGY = "grid"
	STRATEGY_2MA  ENUM_STRATEGY = "2MA"
	STRATEGY_ICHI ENUM_STRATEGY = "ICHI"
	STRATEGY_MN	  ENUM_STRATEGY = "MN"
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

type ExpertAdvisorDelivery struct {
	Debug        bool
	Client       *delivery.Client
	Exchange     *ExchangeInfo
	Account      *AccountInfo
	MWCrypto     *MarketWatchCrypto
	TradeManager *ConversionManager
	File         *FileManager.FileManager
	OrdersLog    string
	ErrorsLog    string
	PositionsID  []int64
	OrdersID     []int64

	MainAsset              string
	MoneyManagemenType     ENUM_MONEY_MANAGEMENT_TYPE
	FixTradeVolume         float64
	DynamicTradePercentage float64

	Strategy ENUM_STRATEGY
	Grid     *GridGraphStrategy
	MA       *MAStrategy
	ICHI     *IchiStrategy
	MN		 *MarketNewtralStrategy
}

func NewExpertAdvisor(key, secret string) *ExpertAdvisorDelivery {
	EA := &ExpertAdvisorDelivery{
		Client:   delivery.NewClient(key, secret),
		File:     new(FileManager.FileManager),
		Debug:    false,
	}
	fmt.Println("Loading exchange info")
	EA.Exchange = NewExchangeInfo(EA)
	fmt.Println("Loading account info")
	EA.Account = NewAccountInfo(EA)
	fmt.Println("Loading market watch")
	EA.MWCrypto = NewMarketWatchCrypto(EA)
	fmt.Println("Loading conversion manager")
	EA.TradeManager = NewConversionManager(EA)
	return EA
}

func (ea *ExpertAdvisorDelivery) Init() {
	fmt.Println("Select Assets:")
	ea.SelectAssets()
	fmt.Println("Money Management:")
	ea.GetMoneyManagement()
	fmt.Println("Select Strategy:")
	ea.SelectStrategy()
}

func (ea *ExpertAdvisorDelivery) Run() {
	fmt.Println(Banner)
	updateTime := time.Now()
	for {
		if time.Since(updateTime) > time.Minute*5 {
			ea.Account.UpdateSymbolsInfo()
			ea.Account.UpdateBalances()
			updateTime = time.Now()
		}
		if ea.Strategy == STRATEGY_GRID {
			ea.Grid.Ontimer()
		} else if ea.Strategy == STRATEGY_2MA {
			ea.MA.Ontimer()
		} else if ea.Strategy == STRATEGY_ICHI {
			ea.ICHI.Ontimer()
		} else if ea.Strategy == STRATEGY_MN {
			ea.MN.Ontimer()
		}
		time.Sleep(time.Millisecond*500)
	}
}

func (ea *ExpertAdvisorDelivery) GetMoneyManagement() {
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

func (ea *ExpertAdvisorDelivery) SelectStrategy() {
	var selection uint
	for {
		fmt.Print(">>> select strategy ( 0 == grid , 1 == moving average, 2 == Ichimoku, 3 = market neutral ): ")
		fmt.Scanln(&selection)
		if selection == 0 {
			ea.Strategy = STRATEGY_GRID
			ea.Grid = NewGridGraph(ea)
			break
		} else if selection == 1 {
			ea.Strategy = STRATEGY_2MA
			ea.MA = NewMAStrategy(ea)
			break
		} else if selection == 2 {
			ea.Strategy = STRATEGY_ICHI
			ea.ICHI = NewIchiStrategy(ea)
			break
		}else if selection == 3 {
			ea.Strategy = STRATEGY_MN
			ea.MN, _ = NewMarketNewtralStrategy(ea,Assets.List)
			break
		} else {
			fmt.Println(">>> wrong value!")
		}		
	}
}

func(ea *ExpertAdvisorDelivery) SelectAssets() {
	for {
		fmt.Print(">>> select main asset: ")
		fmt.Scanln(&ea.MainAsset)
		if ea.Account.AssetExists(ea.MainAsset) {
			break
		} else {
			fmt.Println(">>> invalid asset!")
		}
	}
	for {
		var gotFile bool
		file, err := ioutil.ReadFile("assets.json")
		if err == nil {
			err = json.Unmarshal([]byte(file), &Assets)
			if err == nil {
				gotFile = true
			}
		}
		var UseStoredAssets bool
		if gotFile {
			var UseAssets string
			for {
				fmt.Print(">>> use assets file (y) or (n): ")
				fmt.Scanln(&UseAssets)
				if UseAssets == "y" {
					UseStoredAssets = true
					break
				} else if UseAssets == "n" {
					UseStoredAssets = false
					break
				}else {
					fmt.Println(">>> wrong value!")
				}
			}
		}
		if !UseStoredAssets {
			Assets.List = make([]string, 0)
			for {
				var asset string
				fmt.Println("Type 'exit' to continue")
				for asset != "exit" {
					fmt.Print(">>> select asset: ")
					fmt.Scanln(&asset)
					exist := ea.Account.AssetExists(asset)
					if exist {
						Assets.List = append(Assets.List, asset)
					} else {
						if asset != "exit" {
							fmt.Println("Asset not found!")
						}
					}
				}
				if len(Assets.List) > 0 {
					file, _ := json.MarshalIndent(Assets, "", " ")
					_ = ioutil.WriteFile("assets.json", file, 0644)
					break
				} else {
					fmt.Println("Not Enough Assets!")
				}
			}
			if len(Assets.List) > 0 {
				break
			}
		} else {
			break
		}
	}
	fmt.Println(len(Assets.List),"assets selected.")
}

func (ea *ExpertAdvisorDelivery) UpdateOrdersLog(log string) {
	banner := "//+--------------------------------------------------------+\n"
	log = log + "\n" + banner
	ea.OrdersLog = log + ea.OrdersLog
	ea.File.WriteText("orders.md", ea.OrdersLog)
}

func (ea *ExpertAdvisorDelivery) UpdateSocketStatus(log string) {
	ea.File.WriteText("socket.md", log)
}

func (ea *ExpertAdvisorDelivery) UpdateErrorsLog(log string) {
	banner := "//+--------------------------------------------------------+\n"
	ea.ErrorsLog = time.Now().String() + "\n" + log + "\n" + banner + ea.ErrorsLog
	ea.File.WriteText("errors.md", ea.ErrorsLog)
}

func (ea *ExpertAdvisorDelivery) OpenPosition(symbol string, side delivery.SideType, quantity float64) int64 {
	order, err := ea.Client.NewCreateOrderService().Symbol(symbol).Side(side).Type(delivery.ORDER_TYPE_MARKET).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	if err != nil {
		ea.UpdateErrorsLog(err.Error())
		return -1
	} else {
		ea.PositionsID = append(ea.PositionsID, order.OrderID)
		return order.OrderID
	}
}

func (ea *ExpertAdvisorDelivery) GetPosition(symbol string, ID int64) (*delivery.Order, error) {
	order, err := ea.Client.NewGetOrderService().Symbol(symbol).OrderID(ID).Do(context.Background())
	if err != nil {
		ea.UpdateErrorsLog(err.Error())
		return &delivery.Order{}, errors.New("position not found")
	}
	return order, nil
}

func (ea *ExpertAdvisorDelivery) OpenOrder(symbol string, side delivery.SideType, quantity float64) int64 {
	order, err := ea.Client.NewCreateOrderService().Symbol(symbol).Side(side).Type(delivery.ORDER_TYPE_MARKET).Quantity(fmt.Sprintf("%f", quantity)).Do(context.Background())
	if err != nil {
		ea.UpdateErrorsLog(err.Error())
		return -1
	} else {
		ea.OrdersID = append(ea.OrdersID, order.OrderID)
		return order.OrderID
	}
}

func (ea *ExpertAdvisorDelivery) CancelOrder(symbol string, ID int64) (*delivery.Order, error) {
	order, err := ea.Client.NewGetOrderService().Symbol(symbol).OrderID(ID).Do(context.Background())
	if err != nil {
		ea.UpdateErrorsLog(err.Error())
		return &delivery.Order{}, errors.New("position not found")
	}
	return order, nil
}
