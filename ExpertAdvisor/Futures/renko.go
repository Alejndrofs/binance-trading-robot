package futures

import (
	"errors"
	"time"
)

type ENUM_RENKO_NEW_BAR string

const (
	RENKO_NEW_BAR_UP 	ENUM_RENKO_NEW_BAR = "UP"
	RENKO_NEW_BAR_DOWN 	ENUM_RENKO_NEW_BAR = "DOWN"
	RENKO_NO_NEW_BAR	ENUM_RENKO_NEW_BAR = "NO"
)

type Renko struct {
	Time 				time.Time
	Open 				float64
	High 				float64
	Low 				float64
	Close 				float64
	Tick_volume 		uint64
	Volume 				uint64
	Spread 				int64
}

func NewRenko(open float64) *Renko {
	renko := &Renko{
		Time: time.Now(),
		Open: open,
	}
	return renko
}

type RenkoSerie struct {
	Initialized			bool
	VirtualVolume		float64
	SymbolInfo        	*SymbolInfo
	Serie				[]*Renko
	Grid				float64
	Range				float64
	CountLong			int64
	CountShort			int64
	FirstPrice        	float64
}

func NewRenkoSerie(symbol *SymbolInfo,grid float64,_range float64) *RenkoSerie {
	renkoSerie := &RenkoSerie{
		SymbolInfo: symbol,
		Serie: make([]*Renko, 0),
		Grid: grid,
		Range: _range,
		CountLong: 0,
		CountShort: 0,
	}
	renkoSerie.AddRenko(0)
	renkoSerie.Initialize()
	return renkoSerie
}

func (r *RenkoSerie) Initialize() bool {
	r.Initialized = false
	// Get first price
	PriceFound := false
	price, err := r.SymbolInfo.Expert.Exchange.ConvertFactor(r.SymbolInfo.BaseAsset.AssetName, r.SymbolInfo.ProfAsset.AssetName)
	if err == nil {
		PriceFound = true
	}
	// Get lot from USDT 1000 to current symbol base asset
	VolumeFound := false
	volume, err := r.SymbolInfo.Expert.Exchange.ConvertFactor(r.SymbolInfo.Expert.MainAsset, r.SymbolInfo.BaseAsset.AssetName)
	if err == nil {
		VolumeFound = true
	}
	if PriceFound && VolumeFound {
		r.FirstPrice = price
		r.VirtualVolume = 1000 * volume
		r.Initialized = true
		return true
	} else {
		r.SymbolInfo.Expert.UpdateErrorsLog("error initialaizing symbol " + r.SymbolInfo.SymbolName + " first price not found")
	}
	return false
}

func (r *RenkoSerie) UpdateRenko(lastPrice float64) ENUM_RENKO_NEW_BAR {
	UP := r.Grid * (r.Range + float64(r.CountLong-r.CountShort))
	DOWN := -1 * r.Grid * (r.Range + float64(r.CountShort-r.CountLong))
	
	if r.Serie[len(r.Serie)-1].High < lastPrice {
		r.Serie[len(r.Serie)-1].High = lastPrice
	}

	if r.Serie[len(r.Serie)-1].Low > lastPrice {
		r.Serie[len(r.Serie)-1].Low = lastPrice
	}

	r.Serie[len(r.Serie)-1].Close = lastPrice 
	
	if lastPrice > UP {
		r.CountLong++
		r.AddRenko(UP)
		return RENKO_NEW_BAR_UP
	} 
	if lastPrice < DOWN {
		r.CountShort++
		r.AddRenko(DOWN)
		return RENKO_NEW_BAR_DOWN
	}
	
	return RENKO_NO_NEW_BAR
}

func (s *RenkoSerie) GetVirtualProfit(currentPrice float64) (float64, error) {
	var profit float64
	// Get Points
	var points float64 = currentPrice - s.FirstPrice
	// Get Conversion Factor from points of current symbol to USDT
	conversion, err := s.SymbolInfo.Expert.Exchange.ConvertFactor(s.SymbolInfo.ProfAsset.AssetName, s.SymbolInfo.Expert.MainAsset)
	if err == nil {
		profit = s.VirtualVolume * points * conversion
	} else {
		s.SymbolInfo.Expert.UpdateErrorsLog("get virtual profit err: convert not found")
		return 0, errors.New("virtual profit not found")
	}
	return profit, nil
}

func (r *RenkoSerie) AddRenko(open float64) {
	renko := NewRenko(open)
	r.Serie = append(r.Serie, renko)
}

func (r *RenkoSerie) At(index uint) float64 {
	return r.Serie[index].Open
}

func (r *RenkoSerie) LastValue() float64 {
	return r.Serie[len(r.Serie)-1].Open
}

func (r *RenkoSerie) BarsTotal() uint64 {
	return uint64(len(r.Serie))
}