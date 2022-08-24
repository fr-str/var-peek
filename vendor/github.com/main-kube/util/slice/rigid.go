package slice

import "golang.org/x/exp/constraints"

type Rigid[T any, S constraints.Unsigned] struct {
	data []T
	size S
}

func NewRigid[T any, S constraints.Unsigned](size S) *Rigid[T, S] {
	return &Rigid[T, S]{
		data: make([]T, 0, size),
		size: size,
	}
}

func (r *Rigid[T, S]) Add(x ...T) (removed []T) {
	r.data = append(r.data, x...)
	if S(len(r.data)) > r.size {
		removed = r.data[:S(len(r.data))-r.size]
		r.data = r.data[S(len(r.data))-r.size:]
	}

	return
}

func (r *Rigid[T, S]) GetAll() []T {
	return r.data
}

func (r *Rigid[T, S]) Get(idx S) *T {
	if len(r.data) == 0 {
		return nil
	}
	if idx > S(len(r.data)) {
		idx = S(len(r.data) - 1)
	}
	return &r.data[idx]
}

func (r *Rigid[T, S]) GetLast(amount S) []T {
	if amount > r.size {
		amount = r.size
	}
	return r.data[S(len(r.data))-amount:]
}
