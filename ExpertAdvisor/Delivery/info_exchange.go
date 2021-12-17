package expert

import (
	"Include/go-binance/delivery"
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
	SYMBOL_INFO_NAME            ENUM_SYMBOL_INFO_STRING = "Name"
	SYMBOL_INFO_BASE_ASSET      ENUM_SYMBOL_INFO_STRING = "BaseAsset"
	SYMBOL_INFO_PROF_ASSET      ENUM_SYMBOL_INFO_STRING = "ProfitAsset"
	SYMBOL_INFO_STATUS          ENUM_SYMBOL_INFO_STRING = "Status"
	SYMBOL_INFO_BASE_PRECISION  ENUM_SYMBOL_INFO_INT    = "BasePrecision"
	SYMBOL_INFO_PROF_PRECISION  ENUM_SYMBOL_INFO_INT    = "ProfitPrecision"
	SYMBOL_INFO_ICEBERG_ALLOWED ENUM_SYMBOL_INFO_BOOL   = "icebergAllowed"
	SYMBOL_INFO_OCO_ALLOWED     ENUM_SYMBOL_INFO_BOOL   = "ocoAllowed"
	SYMBOL_INFO_SPOT_ALLOWED    ENUM_SYMBOL_INFO_BOOL   = "isSpotTradingAllowed"
	SYMBOL_INFO_MARGIN_ALLOWED  ENUM_SYMBOL_INFO_BOOL   = "isMarginTradingAllowed"

	SYMBOL_INFO_MIN_LOT  ENUM_SYMBOL_INFO_DOUBLE = "minQty"
	SYMBOL_INFO_MAX_LOT  ENUM_SYMBOL_INFO_DOUBLE = "maxQty"
	SYMBOL_INFO_STEP_LOT ENUM_SYMBOL_INFO_DOUBLE = "stepSize"

	SYMBOL_INFO_MAX_PRICE ENUM_SYMBOL_INFO_DOUBLE = "maxPrice"
	SYMBOL_INFO_MIN_PRICE ENUM_SYMBOL_INFO_DOUBLE = "minPrice"
	SYMBOL_INFO_TICK_SIZE ENUM_SYMBOL_INFO_DOUBLE = "tickSize"

	SYMBOL_INFO_AVG_PRICE_MINS  ENUM_SYMBOL_INFO_INT    = "avgPriceMins"
	SYMBOL_INFO_MULTIPLIER_UP   ENUM_SYMBOL_INFO_DOUBLE = "multiplierUp"
	SYMBOL_INFO_MULTIPLIER_DOWN ENUM_SYMBOL_INFO_DOUBLE = "multiplierDown"

	SYMBOL_INFO_MIN_NOTIONAL           ENUM_SYMBOL_INFO_DOUBLE = "minNotional"
	SYMBOL_INFO_MINNOTIONAL_PRICE_MINS ENUM_SYMBOL_INFO_INT    = "avgPriceMins"
	SYMBOL_INFO_NOTIONAL_APPLY         ENUM_SYMBOL_INFO_BOOL   = "applyToMarket"

	SYMBOL_INFO_ICEB_PARTS_LIMIT ENUM_SYMBOL_INFO_INT = "limit"

	SYMBOL_INFO_MARKET_MIN_LOT  ENUM_SYMBOL_INFO_DOUBLE = "minQty"
	SYMBOL_INFO_MARKET_MAX_LOT  ENUM_SYMBOL_INFO_DOUBLE = "maxQty"
	SYMBOL_INFO_MARKET_STEP_LOT ENUM_SYMBOL_INFO_DOUBLE = "stepSize"

	SYMBOL_INFO_MAX_ALGO_ORDERS ENUM_SYMBOL_INFO_INT = "maxNumAlgoOrders"
)

type ConvertReturn struct {
	FinalAsset string
	Price      float64
}

type ExchangeInfo struct {
	Expert          *ExpertAdvisorDelivery
	Info            *delivery.ExchangeInfo
	ExchangeInfoSet bool
}

func NewExchangeInfo(expert *ExpertAdvisorDelivery) *ExchangeInfo {
	Ex := &ExchangeInfo{
		Expert:          expert,
		ExchangeInfoSet: false,
	}
	return Ex
}

func (ea *ExchangeInfo) FindSymbolByAsset(base Asset, prof Asset) (*SymbolInfo, bool) {
	for i := range ea.Expert.Account.SymbolsList {
		if (ea.Expert.Account.SymbolsList[i].BaseAsset.AssetName == base.AssetName && ea.Expert.Account.SymbolsList[i].ProfAsset.AssetName == prof.AssetName) || (ea.Expert.Account.SymbolsList[i].BaseAsset.AssetName == prof.AssetName && ea.Expert.Account.SymbolsList[i].ProfAsset.AssetName == base.AssetName) {
			return ea.Expert.Account.SymbolsList[i], true
		}
	}
	return &SymbolInfo{}, false
}

func (ea *ExchangeInfo) FindSymbolByName(base string, prof string) (*SymbolInfo, bool) {
	for i := range ea.Expert.Account.SymbolsList {
		if (ea.Expert.Account.SymbolsList[i].BaseAsset.AssetName == base && ea.Expert.Account.SymbolsList[i].ProfAsset.AssetName == prof) || (ea.Expert.Account.SymbolsList[i].BaseAsset.AssetName == prof && ea.Expert.Account.SymbolsList[i].ProfAsset.AssetName == base) {
			return ea.Expert.Account.SymbolsList[i], true
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
					price, err := ea.Expert.MWCrypto.GetLastPrice(s)
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
					price, err := ea.Expert.MWCrypto.GetLastPrice(s)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
				price1, err1 := ea.Expert.MWCrypto.GetLastPrice(v1)
				price2, err2 := ea.Expert.MWCrypto.GetLastPrice(v2)
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
			price, err := ea.Expert.MWCrypto.GetLastPrice(symbolsFrom[i])
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
			price, err := ea.Expert.MWCrypto.GetLastPrice(symbolsFrom[i])
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
func (ea *ExchangeInfo) ExchangeInfoString(x ENUM_EXCHANGE_INFO_STRING) (ret string, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	switch x {
	case "TimeZone":
		return ea.Info.Timezone, nil
	}
	return "", nil
}

func (ea *ExchangeInfo) ExchangeInfoInt(x ENUM_EXCHANGE_INFO_INT) (ret int64, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	switch x {
	case "ServerTime":
		return ea.Info.ServerTime, nil
	}
	return 0, nil
}

func (ea *ExchangeInfo) ExchangeSymbols() (ret []string, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
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
func (ea *ExchangeInfo) SymbolInfoString(symbol string, x ENUM_SYMBOL_INFO_STRING) (ret string, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}
	for _, s := range ea.Info.Symbols {
		if s.Symbol == symbol {
			switch x {
			case "BaseAsset":
				return s.BaseAsset, nil
			case "ProfitAsset":
				return s.QuoteAsset, nil
			}
		}
	}
	return "", errors.New("symbol info string error: symbol not found")
}

func (ea *ExchangeInfo) SymbolInfoInt(symbol string, x ENUM_SYMBOL_INFO_INT) (ret int, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}

	if x == SYMBOL_INFO_ICEB_PARTS_LIMIT {
		return ea.IcebergPartsFilter(symbol, SYMBOL_INFO_ICEB_PARTS_LIMIT)
	}
	if x == SYMBOL_INFO_MAX_ALGO_ORDERS {
		return ea.MaxNumAlgoOrdersFilter(symbol, SYMBOL_INFO_MAX_ALGO_ORDERS)
	}
	if x == SYMBOL_INFO_AVG_PRICE_MINS {
		return ea.PercentPriceFilterInt(symbol, SYMBOL_INFO_AVG_PRICE_MINS)
	}
	if x == SYMBOL_INFO_MINNOTIONAL_PRICE_MINS {
		return ea.MinNotionalFilterInt(symbol, SYMBOL_INFO_MINNOTIONAL_PRICE_MINS)
	}

	for _, s := range ea.Info.Symbols {
		if s.Symbol == symbol {
			switch x {
			case "BasePrecision":
				return int(s.BaseAssetPrecision), nil
			case "ProfitPrecision":
				return int(s.QuotePrecision), nil
			}
		}
	}
	return 0, errors.New("symbol info int error: symbol not found")
}

func (ea *ExchangeInfo) SymbolInfoDouble(symbol string, x ENUM_SYMBOL_INFO_DOUBLE) (ret float64, err error) {
	if !ea.ExchangeInfoSet {
		exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			ea.Info = exchange
			ea.ExchangeInfoSet = true
		}
	}

	if x == SYMBOL_INFO_MAX_LOT {
		return ea.LotSizeFilter(symbol, SYMBOL_INFO_MAX_LOT)
	}
	if x == SYMBOL_INFO_MIN_LOT {
		return ea.LotSizeFilter(symbol, SYMBOL_INFO_MIN_LOT)
	}
	if x == SYMBOL_INFO_STEP_LOT {
		return ea.LotSizeFilter(symbol, SYMBOL_INFO_STEP_LOT)
	}
	if x == SYMBOL_INFO_MAX_PRICE {
		return ea.PriceFilter(symbol, SYMBOL_INFO_MAX_PRICE)
	}
	if x == SYMBOL_INFO_MIN_PRICE {
		return ea.PriceFilter(symbol, SYMBOL_INFO_MIN_PRICE)
	}
	if x == SYMBOL_INFO_TICK_SIZE {
		return ea.PriceFilter(symbol, SYMBOL_INFO_TICK_SIZE)
	}
	if x == SYMBOL_INFO_MULTIPLIER_UP {
		return ea.PercentPriceFilterFloat(symbol, SYMBOL_INFO_MULTIPLIER_UP)
	}
	if x == SYMBOL_INFO_MULTIPLIER_DOWN {
		return ea.PercentPriceFilterFloat(symbol, SYMBOL_INFO_MULTIPLIER_DOWN)
	}
	if x == SYMBOL_INFO_MIN_NOTIONAL {
		return ea.MinNotionalFilterFloat(symbol, SYMBOL_INFO_MIN_NOTIONAL)
	}
	if x == SYMBOL_INFO_MARKET_MIN_LOT {
		return ea.MarketLotSizeFilter(symbol, SYMBOL_INFO_MARKET_MIN_LOT)
	}
	if x == SYMBOL_INFO_MARKET_MAX_LOT {
		return ea.MarketLotSizeFilter(symbol, SYMBOL_INFO_MARKET_MAX_LOT)
	}
	if x == SYMBOL_INFO_MARKET_STEP_LOT {
		return ea.MarketLotSizeFilter(symbol, SYMBOL_INFO_MARKET_STEP_LOT)
	}

	return 0, errors.New("symbol info double error: symbol not found")
}

// LotSizeFilter return lot size filter of symbol
func (ea *ExchangeInfo) LotSizeFilter(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeLotSize) {
					if f == "maxQty" {
						if i, ok := filter["maxQty"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "minQty" {
						if i, ok := filter["minQty"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "stepSize" {
						if i, ok := filter["stepSize"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("lot size filter: filter not found")
}

// PriceFilter return price filter of symbol
func (ea *ExchangeInfo) PriceFilter(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypePriceFilter) {
					if f == "maxPrice" {
						if i, ok := filter["maxPrice"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "minPrice" {
						if i, ok := filter["minPrice"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "tickSize" {
						if i, ok := filter["tickSize"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("price filter: filter not found")
}

// PercentPriceFilter return percent price filter of symbol
func (ea *ExchangeInfo) PercentPriceFilterInt(symbol string, f ENUM_SYMBOL_INFO_INT) (int, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypePercentPrice) {
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
func (ea *ExchangeInfo) PercentPriceFilterFloat(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypePercentPrice) {
					if f == SYMBOL_INFO_MULTIPLIER_UP {
						if i, ok := filter["multiplierUp"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == SYMBOL_INFO_MULTIPLIER_DOWN {
						if i, ok := filter["multiplierDown"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("percent price filter float: filter not found")
}

// MinNotionalFilter return min notional filter of symbol
func (ea *ExchangeInfo) MinNotionalFilterFloat(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeMinNotional) {
					if f == "minNotional" {
						if i, ok := filter["minNotional"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("min Nototional filter: filter not found")
}

// MinNotionalFilter return min notional filter of symbol
func (ea *ExchangeInfo) MinNotionalFilterInt(symbol string, f ENUM_SYMBOL_INFO_INT) (int, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeMinNotional) {
					if f == "avgPriceMins" {
						if i, ok := filter["avgPriceMins"]; ok {
							return int(i.(float64)), nil
						}
					}
				}
			}
		}
	}
	return 0, errors.New("min Nototional filter: filter not found")
}

// MinNotionalFilter return min notional filter of symbol
func (ea *ExchangeInfo) MinNotionalFilterBool(symbol string, f ENUM_SYMBOL_INFO_BOOL) (bool, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeMinNotional) {
					if f == "applyToMarket" {
						if i, ok := filter["applyToMarket"]; ok {
							return i.(bool), nil
						}
					}
				}
			}
		}
	}
	return false, errors.New("min Nototional filter: filter not found")
}

// IcebergPartsFilter return iceberg part filter of symbol
func (ea *ExchangeInfo) IcebergPartsFilter(symbol string, f ENUM_SYMBOL_INFO_INT) (int, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeIcebergParts) {
					if f == "limit" {
						if i, ok := filter["limit"]; ok {
							return int(i.(float64)), nil
						}
					}
				}
			}
		}
	}
	return 0, errors.New("iceberg parts filter: filter not found")
}

// MarketLotSizeFilter return market lot size filter of symbol
func (ea *ExchangeInfo) MarketLotSizeFilter(symbol string, f ENUM_SYMBOL_INFO_DOUBLE) (float64, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeMarketLotSize) {
					if f == "maxQty" {
						if i, ok := filter["maxQty"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "minQty" {
						if i, ok := filter["minQty"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
					if f == "stepSize" {
						if i, ok := filter["stepSize"]; ok {
							return strconv.ParseFloat(i.(string), 64)
						}
					}
				}
			}
		}
	}
	return 0, errors.New("market lotSize filter: filter not found")
}

// MaxNumAlgoOrdersFilter return max num algo orders filter of symbol
func (ea *ExchangeInfo) MaxNumAlgoOrdersFilter(symbol string, f ENUM_SYMBOL_INFO_INT) (int, error) {
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
				if filter["filterType"].(string) == string(delivery.SymbolFilterTypeMaxNumAlgoOrders) {
					if f == "maxNumAlgoOrders" {
						if i, ok := filter["maxNumAlgoOrders"]; ok {
							return int(i.(float64)), nil
						}
					}
				}
			}
		}
	}
	return 0, errors.New("max num algo orders filter: filter not found")
}
