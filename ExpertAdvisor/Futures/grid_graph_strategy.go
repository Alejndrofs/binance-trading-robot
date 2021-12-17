package futures

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type AssetList struct {
	List 			[]string 	`json:"assets"`
}

var Assets AssetList

type GridGraphStrategy struct {
	Expert    		*ExpertAdvisorFutures
	Strategis 		[]*GridStrategy
	AssetsToReset 	[]string
}

func NewGridGraph(ea *ExpertAdvisorFutures) *GridGraphStrategy {
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
				var symbolNotFound bool = false
				symbol, found := ea.Expert.Exchange.FindSymbol(Assets.List[k], Assets.List[j])
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
		if len(ea.Strategis) > 0 {
			break
		} else {
			fmt.Println("No symbols initialized! Press enter to exit")
			var null string
			fmt.Scanln(&null)
			return
		}
	}
}
