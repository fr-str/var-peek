package safe

import (
	"sync"

	"github.com/main-kube/util/slice"

	"golang.org/x/exp/constraints"
)

type RigidSlice[T any, S constraints.Unsigned] struct {
	slice.Rigid[T, S]
	lock sync.RWMutex
}

func NewRigidSlice[T any, S constraints.Unsigned](size S) *RigidSlice[T, S] {
	return &RigidSlice[T, S]{
		*slice.NewRigid[T](size),
		sync.RWMutex{},
	}
}

func (r *RigidSlice[T, S]) Add(x ...T) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Rigid.Add(x...)
}

func (r *RigidSlice[T, S]) GetAll() []T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.GetAll()
}

func (r *RigidSlice[T, S]) Get(idx S) *T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.Get(idx)
}

func (r *RigidSlice[T, S]) GetLast(amount S) []T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.GetLast(amount)
}
