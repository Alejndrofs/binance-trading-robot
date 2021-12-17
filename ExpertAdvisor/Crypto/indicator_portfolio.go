package expert

/*
import (
	"errors"
	"fmt"
)

type PortfolioSerie struct {
	Expert            	*ExpertAdvisorCrypto
	Initialized			bool
	Symbols        		[]*SymbolInfo
	FirstPrice        	[]float64
	VirtualVolume		[]float64
	Buy					[]bool
	Serie				[]*Renko
	Grid				float64
	Range				float64
	CountLong			int64
	CountShort			int64
}

func NewPortfolioSerie(ea *ExpertAdvisorCrypto,symbol []*SymbolInfo,symbolBuy []bool,grid float64,_range float64) *PortfolioSerie {
	portfolio := &PortfolioSerie{
		Expert: ea,
		Symbols: make([]*SymbolInfo, 0),
		FirstPrice: make([]float64, 0),
		VirtualVolume: make([]float64, 0),
		Buy: symbolBuy,
		Serie: make([]*Renko, 0),
		Grid: grid,
		Range: _range,
		CountLong: 0,
		CountShort: 0,
	}
	return portfolio
}

func (r *PortfolioSerie) Initialize() bool {
	r.Initialized = false
	r.FirstPrice = make([]float64, 0)
	r.VirtualVolume = make([]float64, 0)
	for i := 0; i < len(r.Symbols); i++ {
		// Get first price
		PriceFound := false
		price, err := r.Expert.Exchange.ConvertFactor(r.Symbols[i].BaseAsset.AssetName, r.Symbols[i].ProfAsset.AssetName)
		if err == nil {
			PriceFound = true
		}
		// Get lot from USDT 1000 to current symbol base asset
		VolumeFound := false
		volume, err := r.Symbols[i].Expert.Exchange.ConvertFactor(r.Symbols[i].Expert.MainAsset, r.Symbols[i].BaseAsset.AssetName)
		if err == nil {
			VolumeFound = true
		}
		if !PriceFound || !VolumeFound {
			r.Expert.UpdateErrorsLog("error initialaizing symbol " + r.Symbols[i].SymbolName + " first price not found")
			return false
		} else {
			r.FirstPrice = append(r.FirstPrice, price)
			r.VirtualVolume = append(r.VirtualVolume, 1000*volume)
		}
	}
	r.Initialized = true
	return true
}

func (r *PortfolioSerie) UpdatePortfolio() ENUM_RENKO_NEW_BAR {
	var profit float64
	var performance float64 = 0
	var err error
	for i := 0; i < len(r.Symbols); i++ {
		profit, err = r.GetVirtualProfit(uint(i))
		if err != nil {
			return RENKO_NO_NEW_BAR
		} else {
			performance += profit
		}
	}

	UP := r.Grid * (r.Range + float64(r.CountLong-r.CountShort))
	DOWN := -1 * r.Grid * (r.Range + float64(r.CountShort-r.CountLong))

	if r.Serie[len(r.Serie)-1].High < performance {
		r.Serie[len(r.Serie)-1].High = performance
	}

	if r.Serie[len(r.Serie)-1].Low > performance {
		r.Serie[len(r.Serie)-1].Low = performance
	}

	r.Serie[len(r.Serie)-1].Close = performance

	if performance > UP {
		r.CountLong++
		r.AddRenko(UP)
		return RENKO_NEW_BAR_UP
	}
	if performance < DOWN {
		r.CountShort++
		r.AddRenko(DOWN)
		return RENKO_NEW_BAR_DOWN
	}

	return RENKO_NO_NEW_BAR
}

func (r *PortfolioSerie) GetVirtualProfit(index uint) (float64, error) {
	price, err := r.Expert.Exchange.ConvertFactor(r.Symbols[index].BaseAsset.AssetName, r.Symbols[index].ProfAsset.AssetName)
	if err != nil {
		r.Expert.UpdateErrorsLog("update renko, get convert factor error: "+err.Error())
	}
	if price== 0 {
		return 0, errors.New("portfolio serie: get virtual profit error: no price")
	}
	var profit float64
	var points float64 = price - r.FirstPrice[index]
	conversion, err := r.Expert.Exchange.ConvertFactor(r.Symbols[index].ProfAsset.AssetName, r.Expert.MainAsset)
	if err == nil {
		profit = r.VirtualVolume[index] * points * conversion
		if !r.Buy[index] {
			profit = -1*profit
		}
	} else {
		r.Expert.UpdateErrorsLog("get virtual profit err: convert not found")
		return 0, errors.New("virtual profit not found")
	}
	return profit, nil
}

func (r *PortfolioSerie) AddRenko(open float64) {
	renko := NewRenko(open)
	r.Serie = append(r.Serie, renko)
}

func (r *PortfolioSerie) At(index uint) float64 {
	return r.Serie[index].Open
}

func (r *PortfolioSerie) LastOpen() float64 {
	return r.Serie[len(r.Serie)-1].Open
}

func (r *PortfolioSerie) LastHigh() float64 {
	return r.Serie[len(r.Serie)-1].High
}

func (r *PortfolioSerie) LastLow() float64 {
	return r.Serie[len(r.Serie)-1].Low
}

func (r *PortfolioSerie) LastClose() float64 {
	return r.Serie[len(r.Serie)-1].Close
}

func (r *PortfolioSerie) BarsTotal() uint64 {
	return uint64(len(r.Serie))
}

func NewTriangle(symbols [3]string) *Hedge {
	var hedge Hedge
	// add symbols
	for _, s := range symbols {
		var symbol *Symbol = new(Symbol)
		symbol.Name = s
		symbol.BaseAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_BASE_ASSET)
		symbol.ProfAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_PROF_ASSET)
		symbol.MaxLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MAX_LOT)
		symbol.MinLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MIN_LOT)
		symbol.StepLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_STEP_LOT)
		hedge.Symbols = append(hedge.Symbols,*symbol)
		hedge.MagicNumber = MG.NextPrime()
	}
	// buy symbol[0]
	hedge.Symbols[0].BUY = true

	// sell symbol[0] throught symbol[1] & symbol[2]
	if hedge.Symbols[0].BaseAsset == hedge.Symbols[1].BaseAsset {
		hedge.Symbols[1].BUY = false
	} else if hedge.Symbols[0].BaseAsset == hedge.Symbols[1].ProfAsset {
		hedge.Symbols[1].BUY = true
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[1].BaseAsset {
		hedge.Symbols[1].BUY = true
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[1].ProfAsset {
		hedge.Symbols[1].BUY = false
	}

	if hedge.Symbols[0].BaseAsset == hedge.Symbols[2].BaseAsset {
		hedge.Symbols[2].BUY = false
	} else if hedge.Symbols[0].BaseAsset == hedge.Symbols[2].ProfAsset {
		hedge.Symbols[2].BUY = true
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[2].BaseAsset {
		hedge.Symbols[2].BUY = true
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[2].ProfAsset {
		hedge.Symbols[2].BUY = false
	}

	// get formula
	hedge.Formula += "+"+hedge.Symbols[0].Name

	if hedge.Symbols[0].BaseAsset == hedge.Symbols[1].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[1].Name
	} else if hedge.Symbols[0].BaseAsset == hedge.Symbols[1].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[1].Name
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[1].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[1].Name
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[1].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[1].Name
	}

	if hedge.Symbols[0].BaseAsset == hedge.Symbols[2].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[2].Name
	} else if hedge.Symbols[0].BaseAsset == hedge.Symbols[2].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[2].Name
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[2].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[2].Name
	} else if hedge.Symbols[0].ProfAsset == hedge.Symbols[2].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[2].Name
	}

	fmt.Println(">>> new hedge:",hedge.Formula)

	return &hedge
}

func NewFour(name string, A [2]string, B [2]string) *Hedge {
	var hedge Hedge
	// set name
	hedge.SymbolBase.Name = name
	hedge.SymbolBase.BaseAsset,_ = EA.Exchange.SymbolInfoString(name,expert.SYMBOL_INFO_BASE_ASSET)
	hedge.SymbolBase.ProfAsset,_ = EA.Exchange.SymbolInfoString(name,expert.SYMBOL_INFO_PROF_ASSET)
	hedge.MagicNumber = MG.NextPrime()

	// add symbols
	for _, s := range A {
		var symbol *Symbol = new(Symbol)
		symbol.Name = s
		symbol.BaseAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_BASE_ASSET)
		symbol.ProfAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_PROF_ASSET)
		symbol.MaxLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MAX_LOT)
		symbol.MinLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MIN_LOT)
		symbol.StepLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_STEP_LOT)
		hedge.Symbols = append(hedge.Symbols,*symbol)
	}
	// add symbols
	for _, s := range B {
		var symbol *Symbol = new(Symbol)
		symbol.Name = s
		symbol.BaseAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_BASE_ASSET)
		symbol.ProfAsset,_ = EA.Exchange.SymbolInfoString(s,expert.SYMBOL_INFO_PROF_ASSET)
		symbol.MaxLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MAX_LOT)
		symbol.MinLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_MIN_LOT)
		symbol.StepLot,_ = EA.Exchange.SymbolInfoDouble(s,expert.SYMBOL_INFO_STEP_LOT)
		hedge.Symbols = append(hedge.Symbols,*symbol)
	}

	//--- Buy symbol throught symbolsA
	if hedge.SymbolBase.BaseAsset == hedge.Symbols[0].BaseAsset {
		hedge.Symbols[0].BUY = true
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[0].ProfAsset {
		hedge.Symbols[0].BUY = false
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[0].BaseAsset {
		hedge.Symbols[0].BUY = false
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[0].ProfAsset {
		hedge.Symbols[0].BUY = true
	}

	if hedge.SymbolBase.BaseAsset == hedge.Symbols[1].BaseAsset {
		hedge.Symbols[1].BUY = true
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[1].ProfAsset {
		hedge.Symbols[1].BUY = false
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[1].BaseAsset {
		hedge.Symbols[1].BUY = false
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[1].ProfAsset {
		hedge.Symbols[1].BUY = true
	}

	//--- Sell symbol throught symbolsB
	if hedge.SymbolBase.BaseAsset == hedge.Symbols[2].BaseAsset {
		hedge.Symbols[2].BUY = false
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[2].ProfAsset {
		hedge.Symbols[2].BUY = true
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[2].BaseAsset {
		hedge.Symbols[2].BUY = true
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[2].ProfAsset {
		hedge.Symbols[2].BUY = false
	}

	if hedge.SymbolBase.BaseAsset == hedge.Symbols[3].BaseAsset {
		hedge.Symbols[3].BUY = false
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[3].ProfAsset {
		hedge.Symbols[3].BUY = true
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[3].BaseAsset {
		hedge.Symbols[3].BUY = true
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[3].ProfAsset {
		hedge.Symbols[3].BUY = false
	}

	// get formula
	//--- Buy symbol throught symbolsA
	if hedge.SymbolBase.BaseAsset == hedge.Symbols[0].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[0].Name
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[0].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[0].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[0].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[0].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[0].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[0].Name
	}

	if hedge.SymbolBase.BaseAsset == hedge.Symbols[1].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[1].Name
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[1].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[1].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[1].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[1].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[1].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[1].Name
	}

	//--- Sell symbol throught symbolsB
	if hedge.SymbolBase.BaseAsset == hedge.Symbols[2].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[2].Name
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[2].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[2].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[2].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[2].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[2].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[2].Name
	}

	if hedge.SymbolBase.BaseAsset == hedge.Symbols[3].BaseAsset {
		hedge.Formula += "-"+hedge.Symbols[3].Name
	} else if hedge.SymbolBase.BaseAsset == hedge.Symbols[3].ProfAsset {
		hedge.Formula += "+"+hedge.Symbols[3].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[3].BaseAsset {
		hedge.Formula += "+"+hedge.Symbols[3].Name
	} else if hedge.SymbolBase.ProfAsset == hedge.Symbols[3].ProfAsset {
		hedge.Formula += "-"+hedge.Symbols[3].Name
	}

	fmt.Println(">>> new hedge:",hedge.Formula)


	return &hedge
}

type Portfolio struct{
	Assets			[]Hedge
}

func GetThrees(symbols []string) *Portfolio {

	fmt.Println(">>> building portfolio!")

	var portfolio Portfolio
	total := len(symbols);

	//+------------------------------------------------------------------+
	for i:=0 ; i<total-2 ;i++ {
		//---
		sm1  := symbols[i];
		sm1base,_ := EA.Exchange.SymbolInfoString(sm1,expert.SYMBOL_INFO_BASE_ASSET);
		sm1profit,_ := EA.Exchange.SymbolInfoString(sm1,expert.SYMBOL_INFO_PROF_ASSET);

		//+------------------------------------------------------------------+
		for j:=i+1 ; j<total-1 ;j++ {
			//---
			sm2 := symbols[j];
			sm2base,_ := EA.Exchange.SymbolInfoString(sm2,expert.SYMBOL_INFO_BASE_ASSET);
			sm2profit,_ := EA.Exchange.SymbolInfoString(sm2,expert.SYMBOL_INFO_PROF_ASSET);

			//---
			if(sm1base==sm2base || sm1base==sm2profit || sm1profit==sm2base || sm1profit==sm2profit){

				//+------------------------------------------------------------------+
				for k:=j+1; k<total ; k++ {
					//---
					sm3 := symbols[k];
					sm3base,_ := EA.Exchange.SymbolInfoString(sm3,expert.SYMBOL_INFO_BASE_ASSET);
					sm3profit,_ := EA.Exchange.SymbolInfoString(sm3,expert.SYMBOL_INFO_PROF_ASSET);

					//---
					if sm3base==sm1base || sm3base==sm1profit || sm3base==sm2base || sm3base==sm2profit {
						if sm3profit==sm1base || sm3profit==sm1profit || sm3profit==sm2base || sm3profit==sm2profit {
							if sm1 != sm2 && sm1 != sm3 && sm2 != sm3 {
								if !(sm1base == sm2profit && sm1profit == sm2base){
									if !(sm1base == sm3profit && sm1profit == sm3base){
										if !(sm2base == sm3profit && sm2profit == sm3base){
											symbols := [3]string{sm1,sm2,sm3}
											three := NewTriangle(symbols)
											portfolio.Assets = append(portfolio.Assets,*three)
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return &portfolio
}

func GetFours(symbols []string) *Portfolio {
	threes := GetThrees(symbols)
	fours := new(Portfolio)
	total := len(threes.Assets);
   	for i:=0 ; i<total-1 ; i++ {
		ONE := threes.Assets[i]
		for j:=i+1 ; j<total ; j++ {
			TWO := threes.Assets[j]

			if ONE.Symbols[0].Name == TWO.Symbols[0].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[1].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[1].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[0].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[0].Name == TWO.Symbols[1].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[1].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[0].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[0].Name == TWO.Symbols[2].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[1].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[1].Name
				FourPointer := NewFour(ONE.Symbols[0].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[1].Name == TWO.Symbols[0].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[1].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[1].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[1].Name == TWO.Symbols[1].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[1].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[1].Name == TWO.Symbols[2].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[2].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[1].Name
				FourPointer := NewFour(ONE.Symbols[1].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[2].Name == TWO.Symbols[0].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[1].Name
				B[0] = TWO.Symbols[1].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[2].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[2].Name == TWO.Symbols[1].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[1].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[2].Name
				FourPointer := NewFour(ONE.Symbols[2].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}
			if ONE.Symbols[2].Name == TWO.Symbols[2].Name {
				var A, B [2]string
				A[0] = ONE.Symbols[0].Name
				A[1] = ONE.Symbols[1].Name
				B[0] = TWO.Symbols[0].Name
				B[1] = TWO.Symbols[1].Name
				FourPointer := NewFour(ONE.Symbols[2].Name,A,B);
				fours.Assets = append(fours.Assets, *FourPointer)
				continue;
			}

		}
	}
	return fours
}
*/