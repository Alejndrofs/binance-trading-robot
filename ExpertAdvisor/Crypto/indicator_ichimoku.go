package expert

import "errors"

type Ichimoku struct {
	Serie    []float64
	Period   uint
}

func NewIchimoku(period uint) *Ichimoku {
	ma := &Ichimoku{
		Serie:   make([]float64, 0),
		Period:  period,
	}
	return ma
}

func (ma *Ichimoku) Update(renkoSerie *RenkoSerie) {
	var serie []float64
	for i := range renkoSerie.Serie {
		if i >= int(ma.Period) {
         var min, max float64
			for j := i - int(ma.Period); j < i; j++ {
				if renkoSerie.Serie[j].High > max {
               max = renkoSerie.Serie[j].High
            }
            if renkoSerie.Serie[j].Low < min {
               min = renkoSerie.Serie[j].Low
            }
         }
			serie = append(ma.Serie, (max+min)/2)
		} else if i != 0 {
			serie = append(ma.Serie, renkoSerie.Serie[i].Open)
		}
	}
	ma.Serie = serie
}

func (ma *Ichimoku) LastValue() (float64, error) {
	if len(ma.Serie) > 0 {
		return ma.Serie[len(ma.Serie)-1], nil
	} else {
		return 0, errors.New("not enough data")
	}
}

func (ma *Ichimoku) At(index uint) (float64, error) {
	if len(ma.Serie) > 0 && int(index) < len(ma.Serie) {
		return ma.Serie[index], nil
	} else {
		return 0, errors.New("not enough data")
	}
}

func (ma *Ichimoku) BarsTotal() int {
	return len(ma.Serie)
}