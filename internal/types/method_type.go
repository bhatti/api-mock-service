package types

import (
	"fmt"
	"strings"
)

// MethodType for HTTP
type MethodType string

const (
	// Get HTTP request
	Get MethodType = "GET"
	// Post HTTP request
	Post MethodType = "POST"
	// Put HTTP request
	Put MethodType = "PUT"
	// Delete HTTP request
	Delete MethodType = "DELETE"
	// Option HTTP request
	Option MethodType = "OPTION"
	// Head HTTP request
	Head MethodType = "HEAD"
	// Patch HTTP request
	Patch MethodType = "PATCH"
	// Connect HTTP request
	Connect MethodType = "CONNECT"
	// Options HTTP request
	Options MethodType = "OPTIONS"
	// Trace HTTP request
	Trace MethodType = "TRACE"
)

// ToMethod converts string to method
func ToMethod(val string) (MethodType, error) {
	val = strings.ToUpper(val)
	switch {
	case val == string(Get):
		return Get, nil
	case val == string(Post):
		return Post, nil
	case val == string(Put):
		return Put, nil
	case val == string(Delete):
		return Delete, nil
	case val == string(Option):
		return Option, nil
	case val == string(Head):
		return Head, nil
	case val == string(Patch):
		return Patch, nil
	case val == string(Connect):
		return Connect, nil
	case val == string(Options):
		return Options, nil
	case val == string(Trace):
		return Trace, nil
	default:
		return MethodType(val), fmt.Errorf("invalid method '%s'", val)
	}
}
