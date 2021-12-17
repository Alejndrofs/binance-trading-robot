package expert

import "errors"

type MovingAverage struct {
	MA     []float64
	Period uint
}

func NewMovingAverage(period uint) *MovingAverage {
	ma := &MovingAverage{
		MA:     make([]float64, 0),
		Period: period,
	}
	return ma
}

func (ma *MovingAverage) Update(renkoSerie *RenkoSerie) {
	var MovingAvearege []float64
	sum := 0.0
	for i := range renkoSerie.Serie {
		if i >= int(ma.Period) {
			for j := i - int(ma.Period); j < i; j++ {
				sum += renkoSerie.Serie[j].Open
			}
			MovingAvearege = append(ma.MA, sum/float64(ma.Period))
		} else if i != 0 {
			for j := 0; j < i; j++ {
				sum += renkoSerie.Serie[j].Open
			}
			MovingAvearege = append(ma.MA, sum/float64(i))
		}
	}
	ma.MA = MovingAvearege
}

func (ma *MovingAverage) LastValue() (float64, error) {
	if len(ma.MA) > 0 {
		return ma.MA[len(ma.MA)-1], nil
	} else {
		return 0, errors.New("not enough data")
	}
}

func (ma *MovingAverage) At(index int) (float64, error) {
	if len(ma.MA) > 0 && index >= 0 && index < len(ma.MA) {
		return ma.MA[index], nil
	} else {
		return 0, errors.New("not enough data")
	}
}

func (ma *MovingAverage) BarsTotal() int {
	return len(ma.MA)
}