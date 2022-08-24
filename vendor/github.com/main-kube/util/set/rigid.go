package set

import (
	"github.com/main-kube/util/slice"

	"golang.org/x/exp/constraints"
)

type Rigid[T comparable, S constraints.Unsigned] struct {
	slice.Rigid[T, S]
	set Set[T]
}

func NewRigid[T comparable, S constraints.Unsigned](size S) *Rigid[T, S] {
	return &Rigid[T, S]{
		*slice.NewRigid[T](size),
		*New[T](),
	}
}

func (r *Rigid[T, S]) Add(x ...T) {
	toAdd := make([]T, 0, len(x))
	for _, v := range x {
		if !r.set.Contains(v) {
			toAdd = append(toAdd, v)
		}
	}

	r.set.Remove(r.Rigid.Add(toAdd...)...)
}
