package tracing

import (
	"context"
)

type cloneContext struct {
	context.Context // embedded context
	parent          context.Context
}

func (c *cloneContext) Value(key interface{}) interface{} {
	// look for value in new context
	val := c.Context.Value(key)
	if val != nil {
		return val
	}
	// if not found return from old context
	return c.parent.Value(key)
}

// CloneContextValues clones a given context values and returns a new context obj which is not affected by Cancel, Deadline etc
// Deprecated: The function name is a bit confusing, use CloneContextValues instead
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

// MergeParentContext merged the given main context with a parent context, Cancel/Deadline etc are used from the main context and values are looked in both the contexts
// Deprecated: The function name is a bit confusing, use MergeContextValues instead
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
