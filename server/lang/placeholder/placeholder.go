package placeholder

import (
	"strings"
	"sync"
)

type PlaceholderContext struct {
	mu   sync.RWMutex
	data map[string]string
	with []*PlaceholderContext
}

func New(data map[string]string, with ...*PlaceholderContext) *PlaceholderContext {
	return &PlaceholderContext{data: data, with: with}
}

func (ctx *PlaceholderContext) Parse(str string) string {
	ctx.mu.RLock()
	data := ctx.data
	ctx.mu.RUnlock()

	for _, c := range ctx.with {
		if c == nil {
			continue
		}
		c.mu.RLock()
		for k, v := range c.data {
			data[k] = v
		}
		c.mu.RUnlock()
	}

	for k, v := range data {
		str = strings.ReplaceAll(str, "%"+k+"%", v)
	}
	return str
}

func (ctx *PlaceholderContext) Set(key, value string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.data[key] = value
}

func (ctx *PlaceholderContext) Get(key string) string {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.data[key]
}
