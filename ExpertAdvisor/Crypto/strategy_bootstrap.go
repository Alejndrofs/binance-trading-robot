package expert

import (
	"fmt"
	"time"
)

type Performance struct {
	Serie       *RenkoSerie
	isBaseAsset bool
}

type BootsTrap struct {
	Asset        string
	Agreegation  float64
	Performances []*Performance
}

func (b *BootsTrap) Aggregate() {
	b.Agreegation = 0
	for j := range b.Performances {
		var init float64 = 0
		var total uint64 = b.Performances[j].Serie.BarsTotal()
		if total > 30 {
			init = b.Performances[j].Serie.At((uint(total) - 1) - 30)
		}
		if b.Performances[j].isBaseAsset {
			b.Agreegation += (b.Performances[j].Serie.LastClose() - init)
		} else {
			b.Agreegation -= (b.Performances[j].Serie.LastClose() - init)
		}
	}
}

type BootstrapStrategy struct {
	Expert *ExpertAdvisorCrypto
	Series []*RenkoSerie
	Data   []BootsTrap
}

func NewBootstrapStrategy(ea *ExpertAdvisorCrypto) *BootstrapStrategy {
	gridGraph := &BootstrapStrategy{
		Expert: ea,
		Series: make([]*RenkoSerie, 0),
	}
	gridGraph.InitGridGraph()
	return gridGraph
}

func (ea *BootstrapStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	// update prices
	var newBar bool
	for i := range ea.Series {
		bar, err := ea.Series[i].UpdateRenko()
		if bar != RENKO_NO_NEW_BAR && err == nil {
			newBar = true
		}
	}
	// Get Aggregations
	for i := range ea.Data {
		ea.Data[i].Aggregate()
	}
	var max_init bool = false
	var max float64
	var max_index uint = 0
	var min_init bool = false
	var min float64
	var min_index uint = 0
	var close float64
	for i := range ea.Data {
		close = ea.Data[i].Agreegation
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
		comment += ea.Data[i].Asset + ": " + fmt.Sprint(close) + " " + ea.Expert.MainAsset + "\n"
	}
	if newBar {
		var skip_max []uint = make([]uint, 0)
		skip_max = append(skip_max, min_index)
		var succes bool = false
		for !succes && len(skip_max) < len(ea.Data) {
			var skip_min []uint = make([]uint, 0)
			skip_min = append(skip_min, skip_max...)
			for !succes && len(skip_min) < len(ea.Data) {
				var buy bool
				var index uint = 0
				for i := range ea.Data[max_index].Performances {
					var omit bool = false
					for j := range skip_min {
						if i == int(skip_min[j]) {
							omit = true
						}
					}
					if omit {
						continue
					}
					if ea.Data[max_index].Performances[i].Serie.Symbol.BaseAsset.AssetName == ea.Data[max_index].Asset && ea.Data[max_index].Performances[i].Serie.Symbol.ProfAsset.AssetName == ea.Data[min_index].Asset {
						buy = true // trend = true
						index = uint(i)
						break
					}
					if ea.Data[max_index].Performances[i].Serie.Symbol.ProfAsset.AssetName == ea.Data[max_index].Asset && ea.Data[max_index].Performances[i].Serie.Symbol.BaseAsset.AssetName == ea.Data[min_index].Asset {
						buy = false // trend = false
						index = uint(i)
						break
					}
				}
				if buy { // MaxMinLotSizeMainAsset
					_, err := ea.Data[max_index].Performances[index].Serie.Symbol.Buy()
					if err != nil {
						skip_min = append(skip_min, min_index)
						_, min_index = ea.FindOtherMinMax(skip_min)
					} else {
						succes = true
					}
				} else {
					_, err := ea.Data[max_index].Performances[index].Serie.Symbol.Sell()
					if err != nil {
						skip_min = append(skip_min, min_index)
						_, min_index = ea.FindOtherMinMax(skip_min)
					} else {
						succes = true
					}
				}
			}
			if succes {
				break
			} else {
				skip_max = append(skip_max, max_index)
				max_index, _ = ea.FindOtherMinMax(skip_max)
			}
		}
	}
	comment += Space + "\n"
	comment += "Max: " + fmt.Sprint(max) + ", Min: " + fmt.Sprint(min) + "\n"
	comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
	bal, _ := ea.Expert.Account.GetAccountTotalBalanceAs(ea.Expert.MainAsset)
	comment += "Account value: " + fmt.Sprint(bal) + " " + ea.Expert.MainAsset + "\n"
	comment += Space + "\n"
	fmt.Printf("%s", comment)
}

func (ea *BootstrapStrategy) AddToTheList(asset string) {
	for i := range ea.Data {
		if ea.Data[i].Asset == asset {
			return
		}
	}
	bootsTrap := BootsTrap{
		Asset:       asset,
		Agreegation: 0,
	}
	ea.Data = append(ea.Data, bootsTrap)
}

func (mn *BootstrapStrategy) FindOtherMinMax(skip []uint) (max_index uint, min_index uint) {
	var max_init bool = false
	var max float64
	var min_init bool = false
	var min float64
	for i := range mn.Data {
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
		Agg := mn.Data[i].Agreegation

		if max < Agg || !max_init {
			max = Agg
			max_index = uint(i)
			max_init = true
		}
		if min > Agg || !min_init {
			min = Agg
			min_index = uint(i)
			min_init = true
		}
	}
	if !max_init || !min_init {
		return 0, 0
	}
	return max_index, min_index
}

func (ea *BootstrapStrategy) InitGridGraph() {
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
			var found bool
			symbol, err := ea.Expert.Exchange.FindSymbolByName(Assets.List[k], Assets.List[j])
			if err == nil {
				found = true
			}
			if found {
				fmt.Println("Initializing symbol:", symbol.SymbolName)
				serie := NewRenkoSerie(ea.Expert, symbol, AveragePerformace, Range)
				if err == nil {
					for {
						var succes bool = false
						var try uint = 10000
						for try > 0 {
							succes = serie.Initialize()
							if succes {
								ea.Series = append(ea.Series, serie)
								break
							} else {
								try--
								if try == 0 {
									break
								}
							}
						}
						if !succes {
							fmt.Println("Error initialaizing symbol:", serie.Symbol.SymbolName)
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
				Base, _ := ea.Expert.Account.GetAsset(Assets.List[k])
				Prof, _ := ea.Expert.Account.GetAsset(Assets.List[j])
				symbol, err := NewSyntheticSymbol(ea.Expert, Assets.List[k]+Assets.List[j], Base, Prof)
				if err != nil {
					ea.Expert.UpdateErrorsLog("new synth symbol error: new symbol error")
				}
				serie := NewRenkoSerie(ea.Expert, symbol, AveragePerformace, Range)
				if err == nil {
					for {
						var succes bool = false
						var try uint = 10000
						for try > 0 {
							succes = serie.Initialize()
							if succes {
								ea.Series = append(ea.Series, serie)
								break
							} else {
								try--
							}
						}
						if !succes {
							fmt.Println("Error initialaizing synthetic symbol:", serie.Symbol.SymbolName)
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
	for i := range ea.Series {
		ea.AddToTheList(ea.Series[i].Symbol.BaseAsset.AssetName)
		ea.AddToTheList(ea.Series[i].Symbol.ProfAsset.AssetName)
	}
	for i := range ea.Data {
		for j := range ea.Series {
			if ea.Series[j].Symbol.BaseAsset.AssetName == ea.Data[i].Asset {
				performance := &Performance{
					Serie:       ea.Series[j],
					isBaseAsset: true,
				}
				ea.Data[i].Performances = append(ea.Data[i].Performances, performance)
			}
			if ea.Series[j].Symbol.ProfAsset.AssetName == ea.Data[i].Asset {
				performance := &Performance{
					Serie:       ea.Series[j],
					isBaseAsset: false,
				}
				ea.Data[i].Performances = append(ea.Data[i].Performances, performance)
			}
		}
	}
}
