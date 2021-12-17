package expert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type IchiStrategy struct {
	Expert     *ExpertAdvisorCrypto
	SymbolInfo *SymbolInfo
	TimeSerie  *RenkoSerie
	LastPrice  float64
	Performace float64
	Ichimoku   *Ichimoku
	Period     uint
}

func NewIchiStrategy(ea *ExpertAdvisorCrypto) *IchiStrategy {
	strategy := &IchiStrategy{
		Expert: ea,
	}
	strategy.InitIchimoku()
	return strategy
}

func NewSyntheticIchiStrategy(ea *ExpertAdvisorCrypto) *IchiStrategy {
	strategy := &IchiStrategy{
		Expert: ea,
	}
	strategy.InitIchimoku()
	return strategy
}

func (ea *IchiStrategy) InitIchimoku() {
	var gridSize float64
	for {
		fmt.Print(">>> select grid size", "(", ea.Expert.MainAsset, "): ")
		fmt.Scanln(&gridSize)
		if gridSize > 0 {
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
	var Period uint
	for {
		fmt.Print(">>> select ichimoku period: ")
		fmt.Scanln(&Period)
		if Period > 0 {
			ea.Period = Period
			break
		} else {
			fmt.Println(">>> wrong value!")
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
				} else {
					fmt.Println(">>> wrong value!")
				}
			}
		}
		if !UseStoredAssets {
			Assets.List = make([]string, 0)
			for {
				var asset string
				fmt.Println("Type 'exit' to continue")
				for asset != "exit" && len(Assets.List) < 2 {
					fmt.Print(">>> select asset: ")
					fmt.Scanln(&asset)
					exist := ea.Expert.Account.AssetExists(asset)
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
		}
		for k := 0; k < len(Assets.List)-1; k++ {
			for j := k + 1; j < len(Assets.List); j++ {
				symbol, err := ea.Expert.Exchange.FindSymbolByName(Assets.List[k], Assets.List[j])
				var found bool
				if err == nil {
					found = true
				}
				if found {
					ea.SymbolInfo = symbol
				} else {
					fmt.Println("Using synthetic symbol:", Assets.List[k]+Assets.List[j])
					for {
						var succes bool = false
						var try uint = 10
						for try > 0 {
							Base, err1 := ea.Expert.Account.GetAsset(Assets.List[k])
							Prof, err2 := ea.Expert.Account.GetAsset(Assets.List[j])
							symbol, err3 := NewSyntheticSymbol(ea.Expert, Assets.List[k]+Assets.List[j], Base, Prof)
							if err1 != nil || err2 != nil || err3 != nil {
								ea.Expert.UpdateErrorsLog("new strategy error: new symbol error")
							} else {
								ea.SymbolInfo = symbol
							}
						}
						if !succes {
							fmt.Println("Error initialaizing synthetic symbol:", ea.SymbolInfo.SymbolName)
							var retry bool = false
							for {
								fmt.Print(">>> Do you want to try again?(y/n):")
								var res string
								fmt.Scanln(&res)
								if res == "y" {
									retry = true
									break
								} else if res == "n" {
									retry = false
									break
								}
							}
							if !retry {
								break
							}
						} else {
							break
						}
					}
				}
			}
		}
		var exit bool = ea.SymbolInfo != &SymbolInfo{}
		if exit {
			break
		} else {
			fmt.Println("No symbols initialized! Press enter to exit")
			var null string
			fmt.Scanln(&null)
			return
		}
	}
	ea.TimeSerie = NewRenkoSerie(ea.Expert, ea.SymbolInfo, gridSize, Range)
	ea.Ichimoku = NewIchimoku(Period)
}

func (strategy *IchiStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	if strategy.TimeSerie.Initialized {
		newBar, _ := strategy.TimeSerie.UpdateRenko()
		if newBar != RENKO_NO_NEW_BAR {
			strategy.Ichimoku.Update(strategy.TimeSerie)
			if strategy.Ichimoku.BarsTotal() >= int(strategy.Period) && strategy.Ichimoku.BarsTotal() >= 2 {
				prev_price := strategy.TimeSerie.At(uint(strategy.TimeSerie.BarsTotal() - 2))
				prev_ichi, _ := strategy.Ichimoku.At(uint(strategy.Ichimoku.BarsTotal() - 2))

				curr_price := strategy.TimeSerie.At(uint(strategy.TimeSerie.BarsTotal() - 1))
				curr_ichi, _ := strategy.Ichimoku.At(uint(strategy.Ichimoku.BarsTotal() - 1))

				if prev_price < prev_ichi && curr_price > curr_ichi {
					_, err := strategy.SymbolInfo.Buy()
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open buy error: " + err.Error())
					}
				} else if prev_price > prev_ichi && curr_price < curr_ichi {
					_, err := strategy.SymbolInfo.Sell()
					if err != nil {
						strategy.Expert.UpdateErrorsLog("open sell error: " + err.Error())
					}
				}
			}
		}
		close := strategy.TimeSerie.LastClose()
		comment += strategy.SymbolInfo.SymbolName + " Performace: " + fmt.Sprint(close) + " " + strategy.Expert.MainAsset + "\n"
		comment += "Bars Count: " + fmt.Sprint(strategy.TimeSerie.BarsTotal()) + "\n"
		comment += "Last Close: " + fmt.Sprint(strategy.TimeSerie.LastOpen()) + "\n"
		comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
		comment += Space + "\n"
		fmt.Printf("%s", comment)
	} else {
		strategy.TimeSerie.Initialize()
	}
}
