package nullable

import (
	"encoding/json"
	"time"
	"user-manager/util"

	"github.com/volatiletech/null/v8"
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

func FromNullString(n null.String) Nullable[string] {
	if n.Valid {
		return Of(n.String)
	}
	return Empty[string]()
}

func FromNullBool(n null.Bool) Nullable[bool] {
	if n.Valid {
		return Of(n.Bool)
	}
	return Empty[bool]()
}

func FromNullTime(n null.Time) Nullable[time.Time] {
	if n.Valid {
		return Of(n.Time)
	}
	return Empty[time.Time]()
}
