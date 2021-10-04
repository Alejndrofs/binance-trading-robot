package expert

import (
	"Include/go-binance"
	"errors"
	"fmt"
	"math"
)

type ConvertVolumeReturn struct {
	FinalAsset string
	Volume     float64
}

type Conversion struct {
	Symbol *SymbolInfo
	Buy    bool
	Volume float64
}

func NewConversion(symbol *SymbolInfo, buy bool, volume float64) *Conversion {
	conversion := &Conversion{
		Symbol: symbol,
		Buy:    buy,
		Volume: volume,
	}
	return conversion
}

func FloatToString(num float64) string {
	return fmt.Sprintf("%f", num)
}

type Trade struct {
	TradeID uint        "json:'tradeID'"
	Symbol  *SymbolInfo "json:'symbol'"
	Buy     bool        "json:'buy'"
	Price   float64     "json:'price'"
}

var lastID uint = 0

type ConversionManager struct {
	Expert        *ExpertAdvisorCrypto
	Queue         []Conversion
	HistoryTrades []*Trade
}

func (cq *ConversionManager) newTrade(symbol *SymbolInfo, buy bool, price float64) *Trade {
	trade := &Trade{
		TradeID: lastID,
		Buy:     buy,
		Price:   price,
	}
	lastID++
	cq.HistoryTrades = append(cq.HistoryTrades, trade)
	return trade
}

func (cq *ConversionManager) isPosition(symbol *SymbolInfo, buy bool, price float64) *Trade {
	trade := &Trade{
		TradeID: lastID,
		Buy:     buy,
		Price:   price,
	}
	lastID++
	cq.HistoryTrades = append(cq.HistoryTrades, trade)
	return trade
}

func NewConversionManager(expert *ExpertAdvisorCrypto) *ConversionManager {
	Queue := &ConversionManager{
		Expert: expert,
		Queue:  make([]Conversion, 0),
	}
	return Queue
}

func (cq *ConversionManager) AddTransaction(c *Conversion) {
	cq.Queue = append(cq.Queue, *c)
}

func (cq *ConversionManager) ExecuteQueue() (ConvertVolumeReturn, error) {
	for i := len(cq.Queue) - 1; i >= 0; i-- {
		if i == len(cq.Queue)-1 {
			cq.Queue[i].Symbol.NormalizeLot(cq.Queue[i].Volume)
		} else {
			var enough bool = false
			var increment float64 = 0
			pond, err := cq.Expert.Exchange.ConvertFactor(cq.Queue[i+1].Symbol.BaseAsset.AssetName, cq.Queue[i].Symbol.BaseAsset.AssetName)
			if err != nil {
				cq.Expert.UpdateErrorsLog("execute queue error: get conver factor error: " + err.Error())
				return ConvertVolumeReturn{}, err
			}
			for !enough {
				cq.Queue[i].Volume, err = cq.Queue[i].Symbol.NormalizeLot(cq.Queue[i].Volume + increment)
				if err != nil {
					cq.Expert.UpdateErrorsLog("execute queue error: normalize lot error: " + err.Error())
					continue
				}
				cq.Queue[i+1].Volume, err = cq.Queue[i+1].Symbol.NormalizeLot(cq.Queue[i+1].Volume)
				if err != nil {
					cq.Expert.UpdateErrorsLog("execute queue error: normalize lot error: " + err.Error())
					continue
				}
				if cq.Queue[i+1].Volume*pond < cq.Queue[i].Volume {
					enough = true
				}
				increment += cq.Queue[i].Symbol.LotStep
			}
		}
	}
	for i := 0; i < len(cq.Queue); i++ {
		conv, err := cq.Expert.Exchange.ConvertFactor(cq.Queue[i].Symbol.BaseAsset.AssetName, cq.Queue[i].Symbol.ProfAsset.AssetName)
		if err == nil {
			if cq.Queue[i].Buy && cq.Queue[i].Symbol.ProfAsset.IsEnoughMoney(cq.Queue[i].Volume*conv) {
				var orderID int64 = -1
				var try uint = 2
				for orderID == -1 && try > 0 {
					side := binance.SIDE_TYPE_BUY
					orderID = cq.Expert.OpenPosition(cq.Queue[i].Symbol.SymbolName, side, cq.Queue[i].Volume)
					if orderID == -1 {
						cq.Expert.UpdateErrorsLog("execute queue error: place market order error")
					} else {
						cq.Queue[i].Symbol.BaseAsset.Balance += cq.Queue[i].Volume
						cq.Queue[i].Symbol.ProfAsset.Balance -= cq.Queue[i].Volume * conv
					}
					try--
				}
				if orderID == -1 {
					return ConvertVolumeReturn{}, err
				}
			} else if !cq.Queue[i].Buy && cq.Queue[i].Symbol.BaseAsset.IsEnoughMoney(cq.Queue[i].Volume) {
				var orderID int64 = -1
				var try uint = 2
				for orderID == -1 && try > 0 {
					side := binance.SIDE_TYPE_SELL
					orderID = cq.Expert.OpenPosition(cq.Queue[i].Symbol.SymbolName, side, cq.Queue[i].Volume)
					if orderID == -1 {
						cq.Expert.UpdateErrorsLog("execute queue error: place market order error")
					} else {
						cq.Queue[i].Symbol.ProfAsset.Balance += cq.Queue[i].Volume * conv
						cq.Queue[i].Symbol.BaseAsset.Balance -= cq.Queue[i].Volume
					}
					try--
				}
				if orderID == -1 {
					return ConvertVolumeReturn{}, err
				}
			} else {
				return ConvertVolumeReturn{}, errors.New("not enough money")
			}
		}
	}
	var lastAsset string
	var vol float64
	if cq.Queue[len(cq.Queue)-1].Buy {
		lastAsset = cq.Queue[len(cq.Queue)-1].Symbol.BaseAsset.AssetName
		vol = cq.Queue[len(cq.Queue)-1].Volume
	} else {
		price, err := cq.Expert.Exchange.ConvertFactor(cq.Queue[0].Symbol.BaseAsset.AssetName, cq.Queue[0].Symbol.ProfAsset.AssetName)
		if err != nil {
			cq.Expert.UpdateErrorsLog("execute queue error: get conver factor error: " + err.Error())
			return ConvertVolumeReturn{}, err
		}
		lastAsset = cq.Queue[len(cq.Queue)-1].Symbol.ProfAsset.AssetName
		vol = cq.Queue[len(cq.Queue)-1].Volume * price
	}

	result := ConvertVolumeReturn{
		FinalAsset: lastAsset,
		Volume:     vol,
	}

	return result, nil
}

func (ea *ConversionManager) ConvertFactorVolume(volume float64, from, to string) (ConvertVolumeReturn, error) {
	if from == to {
		ret := ConvertVolumeReturn{
			FinalAsset: to,
			Volume:     volume,
		}
		return ret, nil
	}
	conversionTree, err := ea.ConvertTree(volume, from, to)
	if err != nil {
		ea.Expert.UpdateErrorsLog("convert factor voume error: " + err.Error())
		return ConvertVolumeReturn{}, err
	}
	var minimumTreeIndex int
	var minimumTreeLen int
	for i := range conversionTree.Conversions {
		if i == 0 || len(conversionTree.Conversions[i].Queue) < minimumTreeLen {
			minimumTreeIndex = i
			minimumTreeLen = len(conversionTree.Conversions[i].Queue)
		}
	}

	AssetFrom, err1 := ea.Expert.Account.GetAsset(from)
	if err1 != nil {
		return ConvertVolumeReturn{}, err
	}
	AssetTo, err2 := ea.Expert.Account.GetAsset(to)
	if err2 != nil {
		return ConvertVolumeReturn{}, err
	}

	Symbol, err3 := ea.Expert.Account.GetSymbol(AssetFrom, AssetTo)
	if err3 != nil {
		return ConvertVolumeReturn{}, err
	}

	Price, err4 := ea.Expert.MWCrypto.GetLastPrice(Symbol)
	if err4 != nil {
		return ConvertVolumeReturn{}, err
	}

	var Buy bool = true
	if from == Symbol.BaseAsset.AssetName {
		Buy = false
	}

	ret, err := conversionTree.Conversions[minimumTreeIndex].ExecuteQueue()
	if err == nil {
		ea.newTrade(Symbol, Buy, Price)
		return ret, nil
	}
	return ConvertVolumeReturn{}, err
}

type ConversionTree struct {
	Conversions []*ConversionManager
}

func NewConversionTree() *ConversionTree {
	conversionTree := ConversionTree{
		Conversions: make([]*ConversionManager, 0),
	}
	return &conversionTree
}

func (ct *ConversionTree) AddConversion(conversion *ConversionManager) {
	ct.Conversions = append(ct.Conversions, conversion)
}

func (ea *ConversionManager) ConvertTree(volume float64, from, to string) (*ConversionTree, error) {
	conversionTree := NewConversionTree()
	if from == to {
		return conversionTree, nil
	}
	for i := range ea.Expert.Account.AssetsList {
		if ea.Expert.Account.AssetsList[i].AssetName == from || ea.Expert.Account.AssetsList[i].AssetName == to {
			for x, s := range ea.Expert.Account.AssetsList[i].Symbols {
				if from == s.BaseAsset.AssetName && to == s.ProfAsset.AssetName {
					conversionManager := NewConversionManager(ea.Expert)
					lot, err := s.NormalizeLot(volume)
					if err != nil {
						ea.Expert.UpdateErrorsLog("convert tree error: " + err.Error())
						return conversionTree, err
					}
					conversion := NewConversion(ea.Expert.Account.AssetsList[i].Symbols[x], false, lot)
					conversionManager.AddTransaction(conversion)
					conversionTree.AddConversion(conversionManager)
					return conversionTree, nil
				}
				if to == s.BaseAsset.AssetName && from == s.ProfAsset.AssetName {
					pond, err := ea.Expert.Exchange.ConvertFactor(from, s.BaseAsset.AssetName)
					if err == nil {
						conversionManager := NewConversionManager(ea.Expert)
						lot, err := s.NormalizeLot(pond * volume)
						if err != nil {
							ea.Expert.UpdateErrorsLog("convert tree error: " + err.Error())
							return conversionTree, err
						}
						conversion := NewConversion(ea.Expert.Account.AssetsList[i].Symbols[x], true, lot)
						conversionManager.AddTransaction(conversion)
						conversionTree.AddConversion(conversionManager)
						return conversionTree, nil
					} else {
						ea.Expert.UpdateErrorsLog("convert tree error: convert factor not found: " + err.Error())
						return conversionTree, err
					}
				}
			}
		}
	}
	var symbolsFrom []*SymbolInfo
	var symbolsTo []*SymbolInfo
	var asset1 bool
	var asset2 bool
	for i := range ea.Expert.Account.AssetsList {
		if ea.Expert.Account.AssetsList[i].AssetName == from {
			symbolsFrom = ea.Expert.Account.AssetsList[i].Symbols
			asset1 = true
		}
		if ea.Expert.Account.AssetsList[i].AssetName == to {
			symbolsTo = ea.Expert.Account.AssetsList[i].Symbols
			asset2 = true
		}
		if asset1 && asset2 {
			break
		}
	}
	for i := range symbolsFrom {
		for j := range symbolsTo {
			v1 := symbolsFrom[i]
			v2 := symbolsTo[j]
			if v1.SymbolName == v2.SymbolName {
				continue
			}
			base1 := v1.BaseAsset.AssetName
			prof1 := v1.ProfAsset.AssetName
			base2 := v2.BaseAsset.AssetName
			prof2 := v2.ProfAsset.AssetName
			if base1 == base2 && from == prof1 && to == prof2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v1, true, pond1*mainVolume)
				conversion2 := NewConversion(v2, false, pond2*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if base1 == base2 && to == prof1 && from == prof2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v2, true, pond2*mainVolume)
				conversion2 := NewConversion(v1, false, pond1*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if base1 == prof2 && from == prof1 && to == base2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v1, true, pond1*mainVolume)
				conversion2 := NewConversion(v2, true, pond2*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if base1 == prof2 && to == prof1 && from == base2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v2, false, pond2*mainVolume)
				conversion2 := NewConversion(v1, false, pond1*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if base2 == prof1 && from == prof2 && to == base1 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v1, true, pond1*mainVolume)
				conversion2 := NewConversion(v2, true, pond2*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if base2 == prof1 && to == prof2 && from == base1 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v1, false, pond1*mainVolume)
				conversion2 := NewConversion(v2, false, pond2*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if prof1 == prof2 && from == base1 && to == base2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v1, false, pond1*mainVolume)
				conversion2 := NewConversion(v2, true, pond2*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
			if prof1 == prof2 && to == base1 && from == base2 {
				conversionManager := NewConversionManager(ea.Expert)
				pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, base1)
				pond2, err2 := ea.Expert.Exchange.ConvertFactor(from, base2)
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: convert not found: " + err2.Error())
					continue
				}
				minLot1, err1 := v1.GetMinimumLotSize()
				minLot2, err2 := v2.GetMinimumLotSize()
				if err1 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err1.Error())
					continue
				} else if err2 != nil {
					ea.Expert.UpdateErrorsLog("convert tree error: minimum lot size not found: " + err2.Error())
					continue
				}
				mainVolume := math.Max(minLot1*(1/pond1), minLot2*(1/pond2))
				mainVolume = math.Max(mainVolume, volume)
				conversion1 := NewConversion(v2, false, pond2*mainVolume)
				conversion2 := NewConversion(v1, true, pond1*mainVolume)
				conversionManager.AddTransaction(conversion1)
				conversionManager.AddTransaction(conversion2)
				conversionTree.AddConversion(conversionManager)
				return conversionTree, nil
			}
		}
	}
	for i := range symbolsFrom {
		if from == symbolsFrom[i].BaseAsset.AssetName {
			conversionManager := NewConversionManager(ea.Expert)
			lot, err := symbolsFrom[i].NormalizeLot(volume)
			if err != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: normalize lot size error: " + err.Error())
				continue
			}
			conversion := NewConversion(symbolsFrom[i], false, lot)
			conversionManager.AddTransaction(conversion)
			pond, err := ea.Expert.Exchange.ConvertFactor(from, symbolsFrom[i].BaseAsset.AssetName)
			if err != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: convert factor error: " + err.Error())
				continue
			}
			convTree, err := ea.ConvertTree(pond*(lot-symbolsFrom[i].LotStep), symbolsFrom[i].ProfAsset.AssetName, to)
			if err != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: " + err.Error())
				continue
			}
			if len(convTree.Conversions) != 0 {
				var minTreeIndex int
				var minTreeLength int
				for j := range convTree.Conversions {
					if j == 0 || len(convTree.Conversions[j].Queue) < minTreeLength {
						minTreeIndex = j
						minTreeLength = len(convTree.Conversions[j].Queue)
					}
				}
				for _, conv := range convTree.Conversions[minTreeIndex].Queue {
					conversionManager.AddTransaction(&conv)
				}
				conversionTree.AddConversion(conversionManager)
			} else {
				continue
			}
		} else if from == symbolsFrom[i].ProfAsset.AssetName {
			conversionManager := NewConversionManager(ea.Expert)
			pond1, err1 := ea.Expert.Exchange.ConvertFactor(from, symbolsFrom[i].BaseAsset.AssetName)
			if err1 != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: convert factor error: " + err1.Error())
				continue
			}
			lot, err := symbolsFrom[i].NormalizeLot(pond1 * volume)
			if err != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: normalize lot size error: " + err.Error())
				break
			}
			conversion := NewConversion(symbolsFrom[i], true, lot)
			conversionManager.AddTransaction(conversion)
			convTree, err := ea.ConvertTree((lot - symbolsFrom[i].LotStep), symbolsFrom[i].BaseAsset.AssetName, to)
			if err != nil {
				ea.Expert.UpdateErrorsLog("convert tree error error: " + err.Error())
				continue
			}
			if len(convTree.Conversions) != 0 {
				var minTreeIndex int
				var minTreeLength int
				for j := range convTree.Conversions {
					if j == 0 || len(convTree.Conversions[j].Queue) < minTreeLength {
						minTreeIndex = j
						minTreeLength = len(convTree.Conversions[j].Queue)
					}
				}
				for _, conv := range convTree.Conversions[minTreeIndex].Queue {
					conversionManager.AddTransaction(&conv)
				}
				conversionTree.AddConversion(conversionManager)
			} else {
				continue
			}
		}
	}
	return conversionTree, nil
}
