package tracing

import (
	"context"
)

type cloneContext struct {
	context.Context // embedded context
	parent          context.Context
}

func (c *cloneContext) Value(key any) any {
	// look for value in new context
	val := c.Context.Value(key)
	if val != nil {
		return val
	}
	// if not found return from old context
	return c.parent.Value(key)
}

// Deprecated: Use [NewContextWithParentValues] instead.
//
//go:fix inline
func CloneContextValues(parent context.Context) context.Context {
	return NewContextWithParentValues(parent)
}

// NewContextWithParentValues clones a given context values and returns a new context obj which is not affected by Cancel, Deadline etc
// can be used to pass context values to a new context which is not affected by the parent context cancel/deadline etc from parent
func NewContextWithParentValues(parent context.Context) context.Context {
	return &cloneContext{
		parent:  parent,
		Context: context.Background(),
	}
}

// Deprecated: Use [MergeContextValues] instead.
//
//go:fix inline
func MergeParentContext(parent context.Context, main context.Context) context.Context {
	return MergeContextValues(parent, main)
}

// MergeContextValues merged the given main context with a parent context, Cancel/Deadline etc are used from the main context and values are looked in both the contexts
// can be use to merge a parent context with a new context, the new context will have the values from both the contexts
func MergeContextValues(parent context.Context, main context.Context) context.Context {
	return &cloneContext{
		parent:  parent,
		Context: main,
	}
}
