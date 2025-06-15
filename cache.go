package main

import (
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
)

type Cache[T any] struct {
	sf   singleflight.Group
	data *expirable.LRU[string, T]
}

func NewCache[T any](size int, ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		data: expirable.NewLRU[string, T](size, nil, ttl),
	}
}

func (c *Cache[T]) GetOrEmplace(key string, fn func(string) T) T {
	if value, found := c.data.Get(key); found {
		return value
	}

	result, err, _ := c.sf.Do(key, func() (any, error) {
		if value, found := c.data.Get(key); found {
			return value, nil
		}
		value := fn(key)
		c.data.Add(key, value)
		return value, nil
	})

	if err != nil {
		panic(err)
	}

	return result.(T)
}
