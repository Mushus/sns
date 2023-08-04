package array

func Map[I any, O any](s []I, f func(I) O) []O {
	r := make([]O, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

func MapErr[I any, O any](s []I, f func(I) (O, error)) ([]O, error) {
	r := make([]O, len(s))
	for i, v := range s {
		v, err := f(v)
		if err != nil {
			return nil, err
		}
		r[i] = v
	}
	return r, nil
}
