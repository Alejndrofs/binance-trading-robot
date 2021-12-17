package expert

import (
	"fmt"
	"time"
)

type MarketNewtralStrategy struct {
	Expert            	*ExpertAdvisorDelivery
	SymbolInfo        	[]*SymbolInfo
	Assets				[]*Asset
	Performance			[]*RenkoSerie
	LastBar				[]ENUM_RENKO_NEW_BAR
	MaxMinLotSizeMainAsset float64
}

func NewMarketNewtralStrategy(ea *ExpertAdvisorDelivery,assets []string) (*MarketNewtralStrategy, error) {
	strategy := &MarketNewtralStrategy{
		Expert: ea,
		SymbolInfo: make([]*SymbolInfo, 0),
		Assets: make([]*Asset, 0),
		Performance: make([]*RenkoSerie, 0),
		LastBar: make([]ENUM_RENKO_NEW_BAR, 0),
	}
	var AveragePerformace float64
	for {
		fmt.Print(">>> select grid size", "(", ea.MainAsset, "): ")
		fmt.Scanln(&AveragePerformace)
		if AveragePerformace > 0 {
			break
		} else {
			fmt.Println(">>> wrong value!")
		}
	}
	var Range float64
	for {
		fmt.Print(">>> select range size: ")
		fmt.Scanln(&Range)
		if Range > 0 {
			break
		} else {
			fmt.Println(">>> wrong value!")
		}
	}
	for i := range assets {	
		if ea.MainAsset != assets[i]{
			activo, err := ea.Account.GetAsset(assets[i])
			if err != nil {
				continue
			} else {
				strategy.Assets = append(strategy.Assets, activo)
				strategy.LastBar = append(strategy.LastBar, RENKO_NO_NEW_BAR)
			}
			main_asset, err := ea.Account.GetAsset(ea.MainAsset)
			if err != nil {
				continue
			}
			sym, ok := ea.Exchange.FindSymbolByAsset(*main_asset,*activo)
			if !ok {
				ea.UpdateErrorsLog("new market neutral strategy: symbol not found")
				continue
			} else {
				strategy.SymbolInfo = append(strategy.SymbolInfo, sym)
				strategy.Performance = append(strategy.Performance, NewRenkoSerie(ea,strategy.SymbolInfo[len(strategy.SymbolInfo)-1],AveragePerformace,Range)) 
			}
		}
	}
	var max_init bool = false
	var max float64
	for i := range strategy.SymbolInfo {

		min_lot, err := strategy.SymbolInfo[i].GetMinimumLotSize()

		if err != nil {
			i--
		}

		min_lot_main_asset, err := ea.Exchange.ConvertFactor(strategy.SymbolInfo[i].BaseAsset.AssetName,ea.MainAsset)

		if err != nil {
			i--
		}

		if max > min_lot_main_asset || !max_init {
			max_init = true
			max = min_lot
		}
	}

	return strategy, nil
}

/*
func NewSyntheticMarketNewtralStrategy(ea *ExpertAdvisorDelivery,assetA string,assetB string,grid float64,_range float64) (*MarketNewtralStrategy, error) {
	if assetA != assetB {
		strategy := &MarketNewtralStrategy{
			Expert: ea,
		}
		main_asset, err := ea.Account.GetAsset(strategy.Expert.MainAsset)
		if err != nil {
			return &MarketNewtralStrategy{}, errors.New("main asset not found")
		}
		strategy.AssetA, err = ea.Account.GetAsset(assetA)
		if err != nil {
			return &MarketNewtralStrategy{}, errors.New("assetA not found")
		}
		strategy.AssetB, err = ea.Account.GetAsset(assetB)
		if err != nil {
			return &MarketNewtralStrategy{}, errors.New("assetB not found")
		}
		var symbol *SymbolInfo
		symbol, err = NewSyntheticSymbol(ea,strategy.Expert.MainAsset+assetA,main_asset,strategy.AssetA)
		if err != nil {
			strategy.Expert.UpdateErrorsLog("new strategy error: new symbol error")
			return &MarketNewtralStrategy{}, errors.New("new strategy error: new symbol error")
		} else {
			strategy.SymbolInfoA = symbol
			strategy.PerformanceA =  NewRenkoSerie(ea,strategy.SymbolInfoA,grid,_range)
		}
		symbol, err = NewSyntheticSymbol(ea,strategy.Expert.MainAsset+assetB,main_asset,strategy.AssetB)
		if err != nil {
			strategy.Expert.UpdateErrorsLog("new strategy error: new symbol error")
			return &MarketNewtralStrategy{}, errors.New("new strategy error: new symbol error")
		} else {
			strategy.SymbolInfoB = symbol
			strategy.PerformanceB =  NewRenkoSerie(ea,strategy.SymbolInfoB,grid,_range)
		}
		return strategy, nil
	}
	return & MarketNewtralStrategy{}, errors.New("assetA == assetB")
}
*/

func (strategy *MarketNewtralStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	for i := range strategy.Performance {
		if !strategy.Performance[i].Initialized{
			strategy.Performance[i].Initialize()
		}
	}
	var movement bool = false
	for i := range strategy.Performance {	
		if strategy.Performance[i].Initialized {
			newBar := strategy.Performance[i].UpdateRenko()
			if newBar != RENKO_NO_NEW_BAR {
				strategy.LastBar[i] = newBar
				movement = true;
			}
		} else {
			strategy.Performance[i].Initialize()
		}
	}
	var max_init bool = false
	var max float64
	var max_index uint = 0
	var min_init bool = false
	var min float64
	var min_index uint = 0
	var close float64
	for i := range strategy.Performance {
		close = strategy.Performance[i].LastClose()
		if strategy.Performance[i].Initialized {
			if strategy.Expert.MainAsset == strategy.SymbolInfo[i].BaseAsset.AssetName {
				close *= 1
			} else if strategy.Expert.MainAsset == strategy.SymbolInfo[i].ProfAsset.AssetName {
				close *= -1
			}
			if max < close || !max_init {
				max = close
				max_index = uint(i)
				max_init = true
			}
			if min > close || !min_init {
				min = close
				min_index = uint(i)
				min_init = true
			}
		}
		comment += strategy.SymbolInfo[i].SymbolName + " Performace: "+ fmt.Sprint(close)+" " + strategy.Expert.MainAsset + "\n"
		comment += strategy.SymbolInfo[i].SymbolName + " Last bar: " + fmt.Sprint(strategy.Performance[i].LastOpen())+"\n"
		comment += strategy.SymbolInfo[i].SymbolName + " Bars count: "+ fmt.Sprint(strategy.Performance[i].BarsTotal())+"\n"
		comment += Space + "\n"
	}
	if movement {
		var skip []uint = make([]uint, 0)
		skip = append(skip, min_index)
		var succes bool = false
		for !succes && len(skip) < len(strategy.Performance)  {
			if strategy.Expert.MainAsset == strategy.SymbolInfo[max_index].BaseAsset.AssetName { // MaxMinLotSizeMainAsset
				_, err := strategy.SymbolInfo[max_index].Buy()
				if err != nil {
					skip = append(skip, max_index)
					max_index, _ = strategy.FindOtherMinMax(skip)
				} else {
					succes = true
				}
			} else if strategy.Expert.MainAsset == strategy.SymbolInfo[max_index].ProfAsset.AssetName {
				_, err := strategy.SymbolInfo[max_index].Sell()
				if err != nil {
					skip = append(skip, max_index)
					max_index, _ = strategy.FindOtherMinMax(skip)
				} else {
					succes = true
				}
			}
		}
		if !succes {
			strategy.Expert.UpdateErrorsLog("open buy error")
		}
		skip = make([]uint, 0)
		skip = append(skip, max_index)
		succes = false
		for !succes && len(skip) < len(strategy.Performance) {
			if strategy.Expert.MainAsset == strategy.SymbolInfo[min_index].BaseAsset.AssetName {
				_, err := strategy.SymbolInfo[min_index].Sell()
				if err != nil {
					skip = append(skip, min_index)
					_, min_index = strategy.FindOtherMinMax(skip)
				} else {
					succes = true
				}
			} else if strategy.Expert.MainAsset == strategy.SymbolInfo[min_index].ProfAsset.AssetName {
				_, err := strategy.SymbolInfo[min_index].Buy()
				if err != nil {
					skip = append(skip, min_index)
					_, min_index = strategy.FindOtherMinMax(skip)
				} else {
					succes = true
				}
			}
		}
		if !succes {
			strategy.Expert.UpdateErrorsLog("open sell error")
		}
	}
	comment += "Max: " + fmt.Sprint(max) + ", Min: " + fmt.Sprint(min) + "\n"
	comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
	comment += Space + "\n"
	fmt.Printf("%s", comment)
}

func(mn *MarketNewtralStrategy) FindOtherMinMax(skip []uint) (max_index uint,min_index uint) {
	var max_init bool = false
	var max float64
	var min_init bool = false
	var min float64
	for i := range mn.Performance {
		var toSkip bool = false
		for j := range skip {
			if i == int(skip[j]) {
				toSkip = true
				break
			}
		}
		if toSkip {
			continue
		}
		close := mn.Performance[i].LastClose()
		if mn.Performance[i].Initialized {	
			if max < close || !max_init {
				max = close
				max_index = uint(i)
				max_init = true
			}
			if min > close || !min_init {
				min = close
				min_index = uint(i)
				min_init = true
			}
		}
	}
	return max_index, min_index
}