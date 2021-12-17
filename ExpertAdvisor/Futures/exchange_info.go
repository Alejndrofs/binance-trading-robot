package futures

import (
	"Include/go-binance/futures"
	"context"
	"errors"
	"strconv"
)

type ENUM_EXCHANGE_INFO_STRING string
type ENUM_EXCHANGE_INFO_INT string

const (
	EXCHANGE_INFO_TIME_ZONE  ENUM_EXCHANGE_INFO_STRING = "TimeZone"
	EXCHANGE_INFO_SERVERTIME ENUM_EXCHANGE_INFO_INT    = "ServerTime"
)

type ENUM_SYMBOL_INFO_STRING string
type ENUM_SYMBOL_INFO_INT string
type ENUM_SYMBOL_INFO_DOUBLE string
type ENUM_SYMBOL_INFO_BOOL string

const (
	SYMBOL_INFO_NAME 	   ENUM_SYMBOL_INFO_STRING = "Name"
	SYMBOL_INFO_BASE_ASSET ENUM_SYMBOL_INFO_STRING = "BaseAsset"
	SYMBOL_INFO_PROF_ASSET ENUM_SYMBOL_INFO_STRING = "ProfitAsset"
	SYMBOL_INFO_STATUS     ENUM_SYMBOL_INFO_STRING = "Status"

	SYMBOL_INFO_BASE_PRECISION   ENUM_SYMBOL_INFO_INT = "BasePrecision"
	SYMBOL_INFO_PROF_PRECISION   ENUM_SYMBOL_INFO_INT = "ProfitPrecision"
	SYMBOL_INFO_ICEB_PARTS_LIMIT ENUM_SYMBOL_INFO_INT = "IcebergPartsLimit"
	SYMBOL_INFO_MAX_ALGO_ORDERS  ENUM_SYMBOL_INFO_INT = "maxNumAlgoOrders"

	SYMBOL_INFO_MIN_LOT         ENUM_SYMBOL_INFO_DOUBLE = "minQty"
	SYMBOL_INFO_MAX_LOT         ENUM_SYMBOL_INFO_DOUBLE = "maxQty"
	SYMBOL_INFO_STEP_LOT        ENUM_SYMBOL_INFO_DOUBLE = "stepSize"
	SYMBOL_INFO_MARKET_MIN_LOT  ENUM_SYMBOL_INFO_DOUBLE = "minQty"
	SYMBOL_INFO_MARKET_MAX_LOT  ENUM_SYMBOL_INFO_DOUBLE = "maxQty"
	SYMBOL_INFO_MARKET_STEP_LOT ENUM_SYMBOL_INFO_DOUBLE = "stepSize"	
	SYMBOL_INFO_MAX_PRICE       ENUM_SYMBOL_INFO_DOUBLE = "maxPrice"
	SYMBOL_INFO_MIN_PRICE       ENUM_SYMBOL_INFO_DOUBLE = "minPrice"
	SYMBOL_INFO_TICK_SIZE       ENUM_SYMBOL_INFO_DOUBLE = "tickSize"
	SYMBOL_INFO_MULTIPLIER_DECIMAL ENUM_SYMBOL_INFO_DOUBLE = "multiplierDecimal"
	SYMBOL_INFO_MULTIPLIER_UP   ENUM_SYMBOL_INFO_DOUBLE = "multiplierUp"
	SYMBOL_INFO_MULTIPLIER_DOWN ENUM_SYMBOL_INFO_DOUBLE = "multiplierDown"

	SYMBOL_INFO_ICEBERG_ALLOWED ENUM_SYMBOL_INFO_BOOL = "icebergAllowed"
	SYMBOL_INFO_OCO_ALLOWED     ENUM_SYMBOL_INFO_BOOL = "ocoAllowed"
	SYMBOL_INFO_SPOT_ALLOWED    ENUM_SYMBOL_INFO_BOOL = "isSpotTradingAllowed"
	SYMBOL_INFO_MARGIN_ALLOWED  ENUM_SYMBOL_INFO_BOOL = "isMarginTradingAllowed"
)

type ConvertReturn struct {
	FinalAsset string
	Price      float64
}

type ExchangeInfo struct {
	Expert 				*ExpertAdvisorFutures
	Info				*futures.ExchangeInfo
	ExchangeInfoSet		bool
}

func NewExchangeInfo(expert *ExpertAdvisorFutures) *ExchangeInfo {
	Ex := &ExchangeInfo{
		Expert: expert,
		ExchangeInfoSet: false,
	}
	return Ex
}

func (ea *ExchangeInfo) FindSymbol(base, prof string) (*SymbolInfo, bool) {
	for _, asset := range ea.Expert.Account.AssetsList {
		if asset.AssetName == base || asset.AssetName == prof {
			for i := range asset.Symbols {
				if asset.Symbols[i].BaseAsset.AssetName == base && asset.Symbols[i].ProfAsset.AssetName == prof {
					return asset.Symbols[i], true
				}
				if asset.Symbols[i].BaseAsset.AssetName == prof && asset.Symbols[i].ProfAsset.AssetName == base {
					return asset.Symbols[i], true
				}
			}
		}
	}
	return &SymbolInfo{}, false
}

func (ea *ExchangeInfo) ConvertFactor(from, to string) (float64, error) {
	if from == to {
		return 1, nil
	}
	for i := range ea.Expert.Account.AssetsList {
		if ea.Expert.Account.AssetsList[i].AssetName == from || ea.Expert.Account.AssetsList[i].AssetName == to {
			var notFound bool
			for j, s := range ea.Expert.Account.AssetsList[i].Symbols {
				base := ea.Expert.Account.AssetsList[i].Symbols[j].BaseAsset
				prof := ea.Expert.Account.AssetsList[i].Symbols[j].ProfAsset
				if from == base.AssetName && to == prof.AssetName {
					price, err := ea.Expert.MWFutures.GetLastPrice(s)
					if err != nil {
						if ea.Expert.Debug {
							ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err.Error())
						}
						notFound = true
						break
					}
					return price, nil
				}
				if to == base.AssetName && from == prof.AssetName {
					price, err := ea.Expert.MWFutures.GetLastPrice(s)
					if err != nil {
						if ea.Expert.Debug {
							ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err.Error())
						}
						notFound = true
						break
					}
					return (1 / price), nil
				}
			}
			if notFound {
				break
			}
		}
	}
	var symbolsFrom []*SymbolInfo
	var symbolsTo []*SymbolInfo
	for i := range ea.Expert.Account.AssetsList {
		if ea.Expert.Account.AssetsList[i].AssetName == from {
			symbolsFrom = ea.Expert.Account.AssetsList[i].Symbols
		}
		if ea.Expert.Account.AssetsList[i].AssetName == to {
			symbolsTo = ea.Expert.Account.AssetsList[i].Symbols
		}
	}
	for i, v1 := range symbolsFrom {
		for j, v2 := range symbolsTo {
			if v1 == v2 {
				continue
			}
			base1 := symbolsFrom[i].BaseAsset.AssetName
			prof1 := symbolsFrom[i].ProfAsset.AssetName
			base2 := symbolsTo[j].BaseAsset.AssetName
			prof2 := symbolsTo[j].ProfAsset.AssetName
			if base1 == base2 && from == prof1 && to == prof2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err2
				}
				return (1 / price1) * price2, nil
			}
			if base1 == base2 && to == prof1 && from == prof2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return (1 / price2) * price1, nil
			}
			if base1 == prof2 && from == prof1 && to == base2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return (1 / price1) * (1 / price2), nil
			}
			if base1 == prof2 && to == prof1 && from == base2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return price2 * price1, nil
			}
			if base2 == prof1 && from == prof2 && to == base1 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return (1 / price1) * (1 / price2), nil
			}
			if base2 == prof1 && to == prof2 && from == base1 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return price2 * price1, nil
			}
			if prof1 == prof2 && from == base1 && to == base2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return price1 * (1 / price2), nil
			}
			if prof1 == prof2 && to == base1 && from == base2 {
				price1, err1 := ea.Expert.MWFutures.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWFutures.GetLastPrice(v2)
				if err1 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err1.Error())
					}
					return 1, err1
				} else if err2 != nil {
					if ea.Expert.Debug {
						ea.Expert.UpdateErrorsLog("convert factor: get price error: " + err2.Error())
					}
					return 1, err2
				}
				return price2 * (1 / price1), nil
			}
		}
	}
	for i := range symbolsFrom {
		if from == symbolsFrom[i].BaseAsset.AssetName {
			pond, err := ea.ConvertFactor(symbolsFrom[i].ProfAsset.AssetName, to)
			if err != nil {
				if ea.Expert.Debug {
					ea.Expert.UpdateErrorsLog("convert factor: " + err.Error())
				}
				continue
			}
			price, err := ea.Expert.MWFutures.GetLastPrice(symbolsFrom[i])
			if err != nil {
				if ea.Expert.Debug {
					ea.Expert.UpdateErrorsLog("convert factor: " + err.Error())
				}
				continue
			}
			return price * pond, nil
		} else if from == symbolsFrom[i].ProfAsset.AssetName {
			pond, err := ea.ConvertFactor(symbolsFrom[i].BaseAsset.AssetName, to)
			if err != nil {
				if ea.Expert.Debug {
					ea.Expert.UpdateErrorsLog("convert factor: " + err.Error())
				}
				continue
			}
			price, err := ea.Expert.MWFutures.GetLastPrice(symbolsFrom[i])
			if err != nil {
				if ea.Expert.Debug {
					ea.Expert.UpdateErrorsLog("convert factor: " + err.Error())
				}
				continue
			}
			return (1 / price) * pond, nil
		}
	}
	return 1, errors.New("conversion factor: conversion not found")
}

//+------------------------------------------------------+
// Exchange info
//+------------------------------------------------------+
func (ea *ExchangeInfo) ExchangeInfoString(x ENUM_EXCHANGE_INFO_STRING) (ret string,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  "", errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	switch x {
	case "TimeZone":
		return ea.Info.Timezone, nil
	}
	return "", nil
}

func (ea *ExchangeInfo) ExchangeInfoInt(x ENUM_EXCHANGE_INFO_INT) (ret int64,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  0, errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	switch x {
	case "ServerTime":
		return ea.Info.ServerTime, nil
	}
	return 0, nil
}

func (ea *ExchangeInfo) ExchangeSymbols() (ret []string,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			var symbols []string
			return  symbols, errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	var symbols []string
	for _, s := range ea.Info.Symbols {
		symbols = append(symbols, s.Symbol)
	}
	return symbols, nil
}

//+------------------------------------------------------+
// Symbol info
//+------------------------------------------------------+

//
func (ea *ExchangeInfo) SymbolInfoString(symbol string, x ENUM_SYMBOL_INFO_STRING) (ret string,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  "", errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	for _, s := range ea.Info.Symbols {
		if s.Symbol == symbol {
			switch x {
			case "BaseAsset":
				return s.BaseAsset, nil
			case "ProfitAsset":
				return s.QuoteAsset, nil
			case "Status":
				return s.Status, nil
			}
		}
	}
	return "", errors.New("symbol info string error: symbol not found")
}

func (ea *ExchangeInfo) SymbolInfoInt(symbol string, x ENUM_SYMBOL_INFO_INT) (ret int64,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  0, errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	for _, s := range ea.Info.Symbols {
		if s.Symbol == symbol {
			switch x {
			case "BasePrecision":
				return int64(s.BaseAssetPrecision), nil
			case "ProfitPrecision":
				return int64(s.QuotePrecision), nil
			}
		}
	}
	return 0, errors.New("symbol info int error: symbol not found")
}

func (ea *ExchangeInfo) SymbolInfoDouble(symbol string, x ENUM_SYMBOL_INFO_DOUBLE) (ret float64,err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  0, errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	for _, s := range ea.Info.Symbols {
		if s.Symbol == symbol {
			for f := range s.Filters {
				if s.Filters[f].Type == "LOT_SIZE" {
					switch x {
					case "MinLot":
						return s.Filters[f].MinQty, nil
					case "MaxLot":
						return s.Filters[f].MaxQty, nil
					case "StepSize":
						return s.Filters[f].StepSize, nil
					}

				} else if s.Filters[f].Type == "MIN_NOTIONAL" {
					switch x {
					case "MinNotional":
						return s.Filters[f].MinNotional, nil
					}
				}
			}
		}
	}
	return 0, errors.New("symbol info double error: symbol not found")
}

func (ea *ExchangeInfo) SymbolInfoBool(name string,x ENUM_ACCOUNT_INFO_BOOL) (bool,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		} else {
			return  "", errors.New("exchange symbols error: NewExchangeInfoService error")
		}
	}
	for _, s := range ea.Info.Symbols {
		if s.Symbol == name{
			switch x {
			case "icebergAllowed": return s.IcebergAllowed, nil
			case "ocoAllowed": return s.OcoAllowed, nil
			case "isSpotTradingAllowed": return s.IsSpotTradingAllowed, nil
			case "isMarginTradingAllowed": return s.IsMarginTradingAllowed, nil
			}
		}
	}
	return false, errors.New("symbol info bool error: info not found!")
}

//+------------------------------------------------------+
// Symbols Filter
//+------------------------------------------------------+

// LotSizeFilter return lot size filter of symbol
func (ea *ExchangeInfo) LotSizeFilter(symbol string,f ENUM_SYMBOL_INFO_DOUBLE) (float64,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypeLotSize) {
					if f == "maxQty" {
						if i, ok := filter["maxQty"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "minQty" {
						if i, ok := filter["minQty"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "stepSize"{
						if i, ok := filter["stepSize"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
				}
			}
		}	
	}
	return 0, errors.New("lot size filter: filter not found")
}

// PriceFilter return price filter of symbol
func (ea *ExchangeInfo) PriceFilter(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypePrice) {
					if f == "maxPrice"{
						if i, ok := filter["maxPrice"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "minPrice"{
						if i, ok := filter["minPrice"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "tickSize"{
						if i, ok := filter["tickSize"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
				}
			}		
		}
	}
	return 0, errors.New("price filter: filter not found")
}

// PercentPriceFilter return percent price filter of symbol
func (ea *ExchangeInfo) PercentPriceFilterInt(symbol string,f ENUM_SYMBOL_INFO_INT) (int,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypePercentPrice) {
					if f == "avgPriceMins" {
						if i, ok := filter["avgPriceMins"]; ok {
							return int(i.(float64)), nil
						}
					}
				}
			}
		}
	}
	return 0, errors.New("percent price filter int: filter not found")
}

// PercentPriceFilter return percent price filter of symbol
func (ea *ExchangeInfo) PercentPriceFilterFloat(symbol string,f ENUM_SYMBOL_INFO_DOUBLE) (float64,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypePercentPrice) {
					if f == SYMBOL_INFO_MULTIPLIER_UP {
						if i, ok := filter["multiplierUp"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == SYMBOL_INFO_MULTIPLIER_DOWN {
						if i, ok := filter["multiplierDown"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("percent price filter float: filter not found")
}

// MarketLotSizeFilter return market lot size filter of symbol
func (ea *ExchangeInfo) MarketLotSizeFilter(symbol string,f ENUM_SYMBOL_INFO_DOUBLE) (float64,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypeMarketLotSize) {
					if f == "maxQty" {
						if i, ok := filter["maxQty"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "minQty" {
						if i, ok := filter["minQty"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}
					if f == "stepSize" {
						if i, ok := filter["stepSize"]; ok {
							return strconv.ParseFloat(i.(string),64)
						}
					}	
				}
			}
		}
	}
	return 0, errors.New("market lotSize filter: filter not found")
}

// MaxNumAlgoOrdersFilter return max num algo orders filter of symbol
func (ea *ExchangeInfo) MaxNumOrdersFilter(symbol string,f ENUM_SYMBOL_INFO_INT) (int,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypeMaxNumOrders) {
					if f == "maxNumOrders" {
						if i, ok := filter["maxNumOrders"]; ok {
							return int(i.(float64)), nil 
						}
					}
				}
			}
		}
	}
	return 0, errors.New("Mmx num algo orders filter: filter not found")
}

// MaxNumAlgoOrdersFilter return max num algo orders filter of symbol
func (ea *ExchangeInfo) MaxNumAlgoOrdersFilter(symbol string,f ENUM_SYMBOL_INFO_INT) (int,error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for i := range ea.Info.Symbols {
		if ea.Info.Symbols[i].Symbol == symbol {
			for _, filter := range ea.Info.Symbols[i].Filters {
				if filter["filterType"].(string) == string(futures.SymbolFilterTypeMaxNumAlgoOrders) {
					if f == "maxNumAlgoOrders" {
						if i, ok := filter["maxNumAlgoOrders"]; ok {
							return int(i.(float64)), nil 
						}
					}
				}
			}
		}
	}
	return 0, errors.New("Mmx num algo orders filter: filter not found")
}