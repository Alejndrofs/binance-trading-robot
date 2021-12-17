package expert

import (
	"fmt"
	"time"
)

type GridGraphStrategy struct {
	Expert    		*ExpertAdvisorDelivery
	Strategis 		[]*GridStrategy
	AssetsToReset 	[]string
}

func NewGridGraph(ea *ExpertAdvisorDelivery) *GridGraphStrategy {
	gridGraph := &GridGraphStrategy{
		Expert: ea,
		Strategis: make([]*GridStrategy, 0),
	}
	gridGraph.InitGridGraph()
	return gridGraph
}

func (ea *GridGraphStrategy) Ontimer() {
	for i := range ea.Strategis {
		go ea.Strategis[i].Ontimer()
		time.Sleep(time.Millisecond * 10)
	}
	/*
	for i := range ea.Strategis {
		if ea.Strategis[i].ResetAssets {
			ea.AddToTheList(ea.Strategis[i].SymbolInfo.BaseAsset)
			ea.AddToTheList(ea.Strategis[i].SymbolInfo.ProfAsset)
		}
		ea.Strategis[i].ResetAssets = false
	}
	for i := range ea.AssetsToReset {
		for j := range ea.Strategis {
			if ea.Strategis[j].SymbolInfo.BaseAsset == ea.AssetsToReset[i] || ea.Strategis[j].SymbolInfo.ProfAsset == ea.AssetsToReset[i] {
				ea.Strategis[j].InitStrategy()
			}
		}
	}
	*/
	ea.AssetsToReset = make([]string, 0)
}

func (ea *GridGraphStrategy) AddToTheList(asset string) {
	for i := range ea.AssetsToReset {
		if ea.AssetsToReset[i] == asset {
			return 
		}
	}
	ea.AssetsToReset = append(ea.AssetsToReset, asset)
}

func (ea *GridGraphStrategy) InitGridGraph() {
	var Trend bool
	var null string
	for {
		fmt.Print(">>> select strategy mode: Trend (t) or Counter (c): ")
		fmt.Scanln(&null)
		if null == "t" {
			Trend = true
			break
		} else if null == "c" { 
			Trend = false
			break
		}else {
			fmt.Println(">>> wrong value!")
		}
	}
	var AveragePerformace float64
	for {
		fmt.Print(">>> select grid size", "(", ea.Expert.MainAsset, "): ")
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
	for k := 0; k < len(Assets.List)-1; k++ {
		for j := k + 1; j < len(Assets.List); j++ {
			var symbolNotFound bool = false
			symbol, found := ea.Expert.Exchange.FindSymbolByName(Assets.List[k], Assets.List[j])
			if found {
				fmt.Println("Initializing symbol:", symbol.SymbolName)
				strategy, err := NewGridStrategy(ea.Expert, symbol, AveragePerformace,Range,Trend)
				if err == nil {
					for {
						var succes bool = false
						var try uint = 10
						for try > 0 {
							succes = strategy.TimeSerie.Initialize()
							if succes {
								ea.Strategis = append(ea.Strategis, strategy)
								break
							} else {
								try--
								if try == 0 {
									break
								}
							}
						}
						if !succes {
							fmt.Println("Error initialaizing symbol:", strategy.SymbolInfo.SymbolName)
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
									symbolNotFound = true
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
				} else {
					fmt.Println("Error creting symbol", symbol)
				}

			}
			if symbolNotFound || !found {
				fmt.Println("Initializing synthetic symbol:", Assets.List[k]+Assets.List[j])
				strategy, err := NewGridSyntheticStrategy(ea.Expert, Assets.List[k]+Assets.List[j], Assets.List[k], Assets.List[j], AveragePerformace,Range,Trend)
				if err == nil {
					for {
						var succes bool = false
						var try uint = 10
						for try > 0 {
							succes = strategy.TimeSerie.Initialize()
							if succes {
								ea.Strategis = append(ea.Strategis, strategy)
								break
							} else {
								try--
							}
						}
						if !succes {
							fmt.Println("Error initialaizing synthetic symbol:", strategy.SymbolInfo.SymbolName)
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
				} else {
					fmt.Println("Error creting symbol", symbol.SymbolName)
				}
			}
		}
	}
}
