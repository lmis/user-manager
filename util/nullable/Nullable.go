package nullable

import (
	"encoding/json"
	"user-manager/util"
)

type Nullable[T interface{}] struct {
	IsPresent bool
	val       T
}

func (n *Nullable[T]) MarshallJSON() ([]byte, error) {
	if !n.IsPresent {
		return json.Marshal(nil)
	}
	return json.Marshal(n.val)
}

func (n Nullable[T]) OrElse(defaultValue T) T {
	if n.IsPresent {
		return n.val
	}
	return defaultValue
}

func (n Nullable[T]) ToPointer() *T {
	if n.IsPresent {
		return &n.val
	}
	return nil
}
func (n Nullable[T]) OrPanic() T {
	if n.IsPresent {
		return n.val
	}
	panic(util.Error("accessing value of empty Nullable"))
}

func (n Nullable[T]) IsEmpty() bool {
	return !n.IsPresent
}

func Of[T interface{}](val T) Nullable[T] {
	return Nullable[T]{IsPresent: true, val: val}
}
func FromPointer[T interface{}](val *T) Nullable[T] {
	if val == nil {
		return Empty[T]()
	}
	return Of(*val)
}

func NeverNil[T interface{}](val *T) Nullable[*T] {
	if val == nil {
		panic(util.Error("nil value in NeverNil"))
	}
	return Of(val)
}

func MaybeNil[T interface{}](val *T) Nullable[*T] {
	if val == nil {
		return Empty[*T]()
	}
	return Of(val)
}

func Empty[T interface{}]() Nullable[T] {
	return Nullable[T]{}
}
