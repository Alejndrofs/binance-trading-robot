package expert

import (
	"fmt"
	"math"
	"time"
)

type FractalPatern struct {
	PaternID    uint
	Occurrences []uint
}

func NewFractalPatern(id uint) *FractalPatern {
	newPatern := FractalPatern{
		PaternID:    uint(id),
		Occurrences: make([]uint, 0),
	}
	newPatern.Occurrences = append(newPatern.Occurrences, id)
	return &newPatern
}

type FractalStrategy struct {
	Expert        *ExpertAdvisorCrypto
	MatrizFractal [][]int
	NCols         uint
	Paterns       []*FractalPatern
	SymbolInfo    *SymbolInfo
	TimeSerie     *RenkoSerie
	LastPrice     float64
	Performace    float64
	VirtualVolume float64
}

func NewFractalStrategy(ea *ExpertAdvisorCrypto, columns uint) *FractalStrategy {
	fractal := &FractalStrategy{
		Expert:        ea,
		MatrizFractal: make([][]int, 0),
	}
	return fractal
}

func (f *FractalStrategy) Prediction() {

	// get last col
	col := f.MatrizFractal[len(f.MatrizFractal)-1]

	// get last N elements
	N := int(math.Round(float64(f.NCols / 2)))
	// get first omited elements
	Omited := int(f.NCols) - N

	for i := 0; i < len(f.MatrizFractal); i++ {
		found := true
		for j := Omited - 1; j < Omited+N; j++ {
			if f.MatrizFractal[i][j] != col[j] {
				found = false
				break
			}
		}
		if found {

		}
	}
}

func (f *FractalStrategy) Update() {
	var serie []int
	for i := len(f.TimeSerie.Serie) - 1; i >= 0; i-- {
		j := i - 1
		count := 0
		for j >= 0 {
			if f.TimeSerie.Serie[j].Open < f.TimeSerie.Serie[j+1].Open {
				if count != 0 && count < 0 {
					serie = append(serie, count)
					count = 0
				}
				count++
			}
			if f.TimeSerie.Serie[j].Open > f.TimeSerie.Serie[j+1].Open {
				if count != 0 && count > 0 {
					serie = append(serie, count)
					count = 0
				}
				count--
			}
			if len(serie) >= int(f.NCols) {
				var sorted_serie []int
				for k := len(serie) - 1; k >= 0; k-- {
					sorted_serie = append(sorted_serie, serie[k])
				}
				if sorted_serie[len(sorted_serie)-1] > 0 && sorted_serie[len(sorted_serie)-1] > f.MatrizFractal[len(f.MatrizFractal)-1][len(f.MatrizFractal[len(f.MatrizFractal)-1])-1] {
					f.MatrizFractal[len(f.MatrizFractal)-1] = sorted_serie
					f.SearchLastPatern()
				}
				if sorted_serie[len(sorted_serie)-1] < 0 && sorted_serie[len(sorted_serie)-1] < f.MatrizFractal[len(f.MatrizFractal)-1][len(f.MatrizFractal[len(f.MatrizFractal)-1])-1] {
					f.MatrizFractal[len(f.MatrizFractal)-1] = sorted_serie
					f.SearchLastPatern()
				}
				if sorted_serie[len(sorted_serie)-1] > 0 && sorted_serie[len(sorted_serie)-1] < f.MatrizFractal[len(f.MatrizFractal)-1][len(f.MatrizFractal[len(f.MatrizFractal)-1])-1] {
					f.MatrizFractal = append(f.MatrizFractal, sorted_serie)
					f.SearchAllPaterns()
				}
				if sorted_serie[len(sorted_serie)-1] < 0 && sorted_serie[len(sorted_serie)-1] > f.MatrizFractal[len(f.MatrizFractal)-1][len(f.MatrizFractal[len(f.MatrizFractal)-1])-1] {
					f.MatrizFractal = append(f.MatrizFractal, sorted_serie)
					f.SearchAllPaterns()
				}
				break
			}
			j--
		}
	}

}

func (f *FractalStrategy) SearchAllPaterns() {
	f.Paterns = make([]*FractalPatern, 0)
	for i := 0; i < len(f.MatrizFractal); i++ {
		exist := false
		for p := range f.Paterns {
			if f.Paterns[p].PaternID == uint(i) {
				exist = true
				break
			}
		}
		if !exist {
			patern := NewFractalPatern(uint(i))
			f.Paterns = append(f.Paterns, patern)
		}
		for j := i + 1; j < len(f.MatrizFractal); j++ {
			if len(f.MatrizFractal[i]) == len(f.MatrizFractal[j]) {
				found := true
				for k := 0; k < len(f.MatrizFractal[i]); k++ {
					if f.MatrizFractal[i][k] != f.MatrizFractal[j][k] {
						found = false
						break
					}
				}
				if found {
					for p := range f.Paterns {
						if f.Paterns[p].PaternID == uint(i) {
							f.Paterns[p].Occurrences = append(f.Paterns[p].Occurrences, uint(j))
						}
					}
				}
			}
		}
	}
}

func (f *FractalStrategy) SearchLastPatern() {
	for i := 0; i < len(f.MatrizFractal); i++ {
		exist := false
		for p := range f.Paterns {
			if f.Paterns[p].PaternID == uint(i) {
				exist = true
				break
			}
		}
		if !exist {
			patern := NewFractalPatern(uint(i))
			f.Paterns = append(f.Paterns, patern)
		}
		if len(f.MatrizFractal[i]) == len(f.MatrizFractal[len(f.MatrizFractal)-1]) {
			found := true
			for k := 0; k < len(f.MatrizFractal[i]); k++ {
				if f.MatrizFractal[i][k] != f.MatrizFractal[len(f.MatrizFractal)-1][k] {
					found = false
					break
				}
			}
			if found {
				for p := range f.Paterns {
					if f.Paterns[p].PaternID == uint(i) {
						f.Paterns[p].Occurrences = append(f.Paterns[p].Occurrences, uint(len(f.MatrizFractal)-1))
					}
				}
			}
		}
	}
}

func (f *FractalStrategy) Ontimer() {
	currentTime := time.Now().String()
	currentTimeSince := time.Now()
	comment := currentTime + "\n"
	comment += Space + "\n"
	if f.TimeSerie.Initialized {
		newBar, _ := f.TimeSerie.UpdateRenko()
		if newBar != RENKO_NO_NEW_BAR {
			f.Update()
		}
		close := f.TimeSerie.LastClose()
		comment += f.SymbolInfo.SymbolName + " Performace: " + fmt.Sprint(close) + " " + f.Expert.MainAsset + "\n"
		comment += "Total Paterns: " + fmt.Sprint(len(f.MatrizFractal)) + "\n"
		comment += "Bars Count: " + fmt.Sprint(f.TimeSerie.BarsTotal()) + "\n"
		comment += "Last Close: " + fmt.Sprint(f.TimeSerie.LastOpen()) + "\n"
		comment += "Execution time: " + time.Since(currentTimeSince).String() + "\n"
		comment += Space + "\n"
		fmt.Printf("%s", comment)
	} else {
		f.TimeSerie.Initialize()
	}
}
