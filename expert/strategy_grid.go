package expert

import (
	"errors"
	"fmt"
	"time"
)

type GridStrategy struct {
	Expert             *ExpertAdvisorCrypto
	SymbolInfo         *SymbolInfo
	TimeSerie          *RenkoSerie
	LastPrice          float64
	Performace         float64
	VirtualVolume      float64
	ResetAssets        bool
	Trend              bool
	longOrdersOnQueue  uint
	shortOrdersOnQueue uint
	lastTradeDown      bool
	totalLong          int
	totalShort         int
}

func NewGridStrategy(ea *ExpertAdvisorCrypto, symbol *SymbolInfo, grid float64, _range float64, trend bool) (*GridStrategy, error) {
	strategy := &GridStrategy{
		Expert:     ea,
		SymbolInfo: symbol,
		Trend:      trend,
	}
	strategy.TimeSerie = NewRenkoSerie(ea, strategy.SymbolInfo, grid, _range)
	return strategy, nil
}

func NewGridSyntheticStrategy(ea *ExpertAdvisorCrypto, symbolName string, base string, profit string, grid float64, _range float64, trend bool) (*GridStrategy, error) {
	strategy := &GridStrategy{
		Expert: ea,
		Trend:  trend,
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
	strategy.TimeSerie = NewRenkoSerie(ea, strategy.SymbolInfo, grid, _range)
	return strategy, nil
}

func (strategy *GridStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	var close float64
	
	for {
		if strategy.TimeSerie.Initialized {
			newBar := strategy.TimeSerie.UpdateRenko()
			close = strategy.TimeSerie.LastClose()

			if strategy.Trend {
				if newBar == RENKO_NEW_BAR_UP {
					if strategy.longOrdersOnQueue > 0 {
						succes, err := strategy.SymbolInfo.Buy()
						if succes {
							strategy.ResetAssets = true
							strategy.totalLong++
						}
						if err != nil {
							strategy.Expert.UpdateErrorsLog("open buy error: " + err.Error())
							break
						}
						strategy.longOrdersOnQueue = 0
						strategy.lastTradeDown = false
						break
					}
					if strategy.lastTradeDown && strategy.totalLong-strategy.totalShort >= 0 {
						strategy.shortOrdersOnQueue++
						break
					}
					succes, err := strategy.SymbolInfo.Sell()
					if succes {
						strategy.ResetAssets = true
						strategy.lastTradeDown = false
						strategy.totalShort++
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open sell error: " + err.Error())
					}
				}
				if newBar == RENKO_NEW_BAR_DOWN {
					if strategy.shortOrdersOnQueue > 0 {
						succes, err := strategy.SymbolInfo.Sell()
						if succes {
							strategy.ResetAssets = true
							strategy.totalShort++
						}
						if err != nil {
							strategy.Expert.UpdateErrorsLog("open sell error: " + err.Error())
						}
						strategy.shortOrdersOnQueue = 0
						strategy.lastTradeDown = true
						break
					}
					if !strategy.lastTradeDown && strategy.totalShort-strategy.totalLong >= 0 {
						strategy.longOrdersOnQueue++
						break
					}
					succes, err := strategy.SymbolInfo.Buy()
					if succes {
						strategy.ResetAssets = true
						strategy.lastTradeDown = true
						strategy.totalLong++
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open buy error: " + err.Error())
					}
				}
			} else {
				if newBar == RENKO_NEW_BAR_UP {
					succes, err := strategy.SymbolInfo.Sell()
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open sell error: " + err.Error())
					}
				} else if newBar == RENKO_NEW_BAR_DOWN {
					succes, err := strategy.SymbolInfo.Buy()
					if succes {
						strategy.ResetAssets = true
					}
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open buy error: " + err.Error())
					}
				}
			}
		} else {
			strategy.TimeSerie.Initialize()
		}
		if true {
			break
		}		
	}
	
	comment += strategy.SymbolInfo.SymbolName + " Performace: " + fmt.Sprint(close) + " " + strategy.Expert.MainAsset + "\n"
	comment += "Bars Count: " + fmt.Sprint(strategy.TimeSerie.BarsTotal()) + "\n"
	comment += "Last Close: " + fmt.Sprint(strategy.TimeSerie.LastOpen()) + "\n"
	comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
	comment += Space + "\n"
	fmt.Printf("%s", comment)
}
