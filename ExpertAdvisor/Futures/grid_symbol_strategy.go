package futures

import (
	"errors"
	"fmt"
	"time"
)

type GridStrategy struct {
	Expert            	*ExpertAdvisorFutures
	SymbolInfo        	*SymbolInfo
	TimeSerie			*RenkoSerie
	LastPrice         	float64
	Performace        	float64
	VirtualVolume     	float64
	ResetAssets			bool 
	Trend				bool
}

func NewGridStrategy(ea *ExpertAdvisorFutures, symbol *SymbolInfo,grid float64,_range float64,trend bool) (*GridStrategy, error) {
	strategy := &GridStrategy{
		Expert:            ea,
		SymbolInfo:        symbol,
		Trend: trend,
	}
	strategy.TimeSerie =  NewRenkoSerie(strategy.SymbolInfo,grid,_range)
	return strategy, nil
}

func NewGridSyntheticStrategy(ea *ExpertAdvisorFutures,symbolName string,base string,profit string,grid float64,_range float64,trend bool) (*GridStrategy, error) {
	strategy := &GridStrategy{
		Expert:            ea,
		Trend: trend,
	}
	Base, _ := ea.Account.GetAsset(base)
	Prof, _ := ea.Account.GetAsset(profit)
	symbol, err := NewSyntheticSymbol(ea, symbolName, Base, Prof)
	if err != nil {
		strategy.Expert.UpdateErrorsLog("new strategy error: new symbol error")
		return &GridStrategy{}, errors.New("new strategy error: new symbol error")
	} else {
		strategy.SymbolInfo = symbol
	}
	strategy.TimeSerie =  NewRenkoSerie(strategy.SymbolInfo,grid,_range)
	return strategy, nil
}

func (strategy *GridStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	if strategy.TimeSerie.Initialized {
		price, err := strategy.Expert.Exchange.ConvertFactor(strategy.SymbolInfo.BaseAsset.AssetName, strategy.SymbolInfo.ProfAsset.AssetName)
		if err == nil {
			strategy.LastPrice = price
		}
		if strategy.LastPrice == 0 {
			return
		}
		performace, err := strategy.TimeSerie.GetVirtualProfit(strategy.LastPrice)
		if err != nil {
			return
		}
		newBar := strategy.TimeSerie.UpdateRenko(performace)
		open := strategy.TimeSerie.LastValue()
		if err == nil {
			if newBar == RENKO_NEW_BAR_UP {
				if strategy.Trend {
					succes, err := strategy.SymbolInfo.Buy(price)
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open buy error: "+err.Error())
					}
				} else if !strategy.Trend {
					succes, err := strategy.SymbolInfo.Sell(price)
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open sell error: "+err.Error())
					}
				}
			} else if newBar == RENKO_NEW_BAR_DOWN {
				if strategy.Trend {
					succes, err := strategy.SymbolInfo.Sell(price)
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open sell error: "+err.Error())
					}
				} else if !strategy.Trend {
					succes, err := strategy.SymbolInfo.Buy(price)
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open buy error: "+err.Error())
					}
				} 
			}
		}		
		comment += strategy.SymbolInfo.SymbolName + " Performace: "+ fmt.Sprint(performace)+" " + strategy.Expert.MainAsset + "\n"
		comment += "Bars Count: "+ fmt.Sprint(strategy.TimeSerie.BarsTotal())+"\n"
		comment += "Last Close: "+fmt.Sprint(open)+"\n"
		comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
		comment += Space + "\n"
		fmt.Printf("%s", comment)
	} else {
		strategy.TimeSerie.Initialize()
	}
}
