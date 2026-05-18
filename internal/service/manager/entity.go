package manager

import "reflect"

// isNilEntity reports whether v is an untyped nil or a nil pointer-like value.
func isNilEntity[T any](v T) bool {
	if any(v) == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}
