// Package Transformer transformx a type from itself to another type.
package transformer

import "time"

// String returns a pointer to the string passed to it.
func String(s string) *string {
	return &s
}

// ToString returns the deferenced value of the pointer s passed to it
// If s is nil, an empty string is returned, else the dereferenced value is returned
func ToString(in *string) (out string) {
	if in == nil {
		return out
	}
	return *in
}

// Bool returns a pointer the boolean b
func Bool(b bool) *bool {
	return &b
}

// ToBool take a pointer a boolean b and return the deferenced value
// If b is nil, false is returned, else the dereferenced value is returned
func ToBool(in *bool) (out bool) {
	if in == nil {
		return out
	}
	return *in
}

func Uint(u uint) *uint {
	return &u
}

func ToUint(in *uint) (out uint) {
	if in == nil {
		return out
	}
	return *in
}

func Time(in time.Time) *time.Time {
	return &in
}

func ToTime(in *time.Time) (out time.Time) {
	if in == nil {
		return out
	}
	return *in
}
