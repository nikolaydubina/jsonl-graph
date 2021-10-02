package svgpanzoom

import "gonum.org/v1/gonum/mat"

func identity() *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	})
}

func scale(z float64) *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		z, 0, 0,
		0, z, 0,
		0, 0, 1,
	})
}

func translate(x, y float64) *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		1, 0, x,
		0, 1, y,
		0, 0, 1,
	})
}
