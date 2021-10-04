package expert

/*

import (
	"Include/MagicNumber"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var MG magic.MagicGenerator

type Triangle struct {
	Expert  *ExpertAdvisorCrypto
	Magic   uint64
	Assets  []*Asset
	Symbols []*SymbolInfo
	Formula []bool
}

func NewTriangle(expert *ExpertAdvisorCrypto, assets [3]string) (*Triangle, error) {
	triangle := &Triangle{
		Expert:  expert,
		Magic:   MG.NextPrime(),
		Assets:  make([]*Asset, 0),
		Symbols: make([]*SymbolInfo, 0),
		Formula: make([]bool, 3),
	}
	// add symbols
	for _, s1 := range assets {
		if !triangle.Expert.Account.AssetExists(s1) {
			return &Triangle{}, errors.New("new Triangle: asset not found")
		} else {
			asset, _ := triangle.Expert.Account.GetAsset(s1)
			triangle.Assets = append(triangle.Assets, asset)
		}
	}

	for i := range triangle.Assets {
		for j := i + 1; j < len(triangle.Assets); j++ {
			sym, found := triangle.Expert.Exchange.FindSymbol(triangle.Assets[i].AssetName, triangle.Assets[j].AssetName)
			if found {
				triangle.Symbols = append(triangle.Symbols, sym)
			} else {
				sym, err := NewSyntheticSymbol(expert, triangle.Assets[i].AssetName+triangle.Assets[j].AssetName, triangle.Assets[i], triangle.Assets[j])
				if err != nil {
					return &Triangle{}, err
				} else {
					triangle.Symbols = append(triangle.Symbols, sym)
				}
			}
		}
	}

	// buy symbol[0]
	triangle.Formula[0] = true

	// sell symbol[0] throught symbol[1] & symbol[2]
	if triangle.Symbols[0].BaseAsset == triangle.Symbols[1].BaseAsset {
		triangle.Formula[1] = false
	} else if triangle.Symbols[0].BaseAsset == triangle.Symbols[1].ProfAsset {
		triangle.Formula[1] = true
	} else if triangle.Symbols[0].ProfAsset == triangle.Symbols[1].BaseAsset {
		triangle.Formula[1] = true
	} else if triangle.Symbols[0].ProfAsset == triangle.Symbols[1].ProfAsset {
		triangle.Formula[1] = false
	}

	if triangle.Symbols[0].BaseAsset == triangle.Symbols[2].BaseAsset {
		triangle.Formula[2] = false
	} else if triangle.Symbols[0].BaseAsset == triangle.Symbols[2].ProfAsset {
		triangle.Formula[2] = true
	} else if triangle.Symbols[0].ProfAsset == triangle.Symbols[2].BaseAsset {
		triangle.Formula[2] = true
	} else if triangle.Symbols[0].ProfAsset == triangle.Symbols[2].ProfAsset {
		triangle.Formula[2] = false
	}

	cmmnt := ""
	for i := range triangle.Symbols {
		cmmnt += triangle.Symbols[i].SymbolName + " " + strconv.FormatBool(triangle.Formula[i]) + " "
	}

	fmt.Println(">>> New triangle:", cmmnt)

	return triangle, nil
}

type TringularGridStrategy struct {
	Expert    	*ExpertAdvisorCrypto
	Basket 		*Triangle
	TimeSerie	*RenkoSerie
}

func NewTriangleArbitrageStrategy(expert *ExpertAdvisorCrypto, symbols [3]string) (*TringularGridStrategy, error) {
	TAS := &TringularGridStrategy{
		Expert:    expert,
		Basket: new(Triangle),
	}
	TAS.Basket = NewTriangle()

	return TAS, nil
}

func (strategy *TringularGridStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	if strategy.TimeSerie.Initialized {
		newBar := strategy.TimeSerie.UpdateRenko()
		close := strategy.TimeSerie.LastClose()

		if newBar == RENKO_NEW_BAR_UP {
			if strategy.Trend {
				succes, err := strategy.SymbolInfo.Buy()
				if succes {
					strategy.ResetAssets = true
				}
				if err != nil {
					strategy.Expert.UpdateErrorsLog("open buy error: "+err.Error())
				}
			} else if !strategy.Trend {
				succes, err := strategy.SymbolInfo.Sell()
				if succes {
					strategy.ResetAssets = true
				}
				if err != nil {
					strategy.Expert.UpdateErrorsLog("open sell error: "+err.Error())
				}
			}
		} else if newBar == RENKO_NEW_BAR_DOWN {
			if strategy.Trend {
				succes, err := strategy.SymbolInfo.Sell()
				if succes {
					strategy.ResetAssets = true
				}
				if err != nil {
					strategy.Expert.UpdateErrorsLog("open sell error: "+err.Error())
				}
			} else if !strategy.Trend {
				succes, err := strategy.SymbolInfo.Buy()
				if succes {
					strategy.ResetAssets = true
				}
				if err != nil {
					strategy.Expert.UpdateErrorsLog("open buy error: "+err.Error())
				}
			} 
		}
	
		comment += strategy.SymbolInfo.SymbolName + " Performace: "+ fmt.Sprint(close)+" " + strategy.Expert.MainAsset + "\n"
		comment += "Bars Count: "+ fmt.Sprint(strategy.TimeSerie.BarsTotal())+"\n"
		comment += "Last Close: "+fmt.Sprint(strategy.TimeSerie.LastOpen())+"\n"
		comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
		comment += Space + "\n"
		fmt.Printf("%s", comment)
	} else {
		strategy.TimeSerie.Initialize()
	}
}

*/