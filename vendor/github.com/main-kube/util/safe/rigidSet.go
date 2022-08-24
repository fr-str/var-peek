package safe

import (
	"sync"

	"github.com/main-kube/util/set"

	"golang.org/x/exp/constraints"
)

type RigidSet[T comparable, S constraints.Unsigned] struct {
	set.Rigid[T, S]
	lock sync.RWMutex
}

func NewRigidSet[T comparable, S constraints.Unsigned](size S) *RigidSet[T, S] {
	return &RigidSet[T, S]{
		*set.NewRigid[T](size),
		sync.RWMutex{},
	}
}

func (r *RigidSet[T, S]) Add(x ...T) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Rigid.Add(x...)
}

func (r *RigidSet[T, S]) GetAll() []T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.GetAll()
}

func (r *RigidSet[T, S]) Get(idx S) *T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.Get(idx)
}

func (r *RigidSet[T, S]) GetLast(amount S) []T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.Rigid.GetLast(amount)
}
