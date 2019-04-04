package v3

import "context"

// c is a alias for a testing context
func c() context.Context { return context.Background() }

type cMap map[interface{}]interface{}

// ctxFromMap creates a new context with given keys and values
func cFromMap(m cMap) context.Context {
	c := c()
	for k, v := range m {
		c = context.WithValue(c, k, v)
	}
	return c
}
