package main

type tuple3[T1 any, T2 any, T3 any] struct {
	e1 T1
	e2 T2
	e3 T3
}

func tuple3_new[T1 any, T2 any, T3 any](e1 T1, e2 T2, e3 T3) *tuple3[T1, T2, T3] {
	return &tuple3[T1, T2, T3]{e1: e1, e2: e2, e3: e3}
}

func (t *tuple3[T1, T2, T3]) Elements() (T1, T2, T3) {
	return t.e1, t.e2, t.e3
}
