package expert

type LinearRegression struct {
	// F = A*X+B
	A				float64
	B				float64
}

func (l *LinearRegression) Regression(x []float64,y []float64) bool {
	if len(x) != len(y) {
		return false
	}
	
	// calculate the mean of the arrays
	x_mean := l.Mean(x)
	y_mean := l.Mean(y)

	// calculate SS_xx
	SS_xy := 0.0
	for i := 0; i < len(x); i++ {
		SS_xy += (x[i]-x_mean)*(y[i]-y_mean)
	}

	// calculate SS_xy
	SS_xx := 0.0
	for i := 0; i < len(x); i++ {
		SS_xx += (x[i]-x_mean)*(x[i]-x_mean)
	}

	// A = SS_xy / SS_xx
	l.A = SS_xy/SS_xx

	// B = y_mean - A*x_mean
	l.B = y_mean - l.A*x_mean

	return true
}

func (l *LinearRegression) Mean(x []float64) float64 {
	mean := 0.0
	for i := 0; i < len(x); i++ {
		mean += x[i]
	}
	if len(x) != 0 {
		return mean/float64(len(x))
	}
	return 0
}