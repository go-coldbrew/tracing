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
func CloneContextValues(parent context.Context) context.Context {
	return &cloneContext{
		parent:  parent,
		Context: context.Background(),
	}
}

// MergeContextValues merged the given main context with a parent conetext, Cancel/Deadline etc are used from the main context and values are look in both the context
func MergeContextValues(parent context.Context, main context.Context) context.Context {
	return &cloneContext{
		parent:  parent,
		Context: main,
	}
}
