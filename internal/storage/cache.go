package storage

import (
	"order-service/internal/models"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]models.Order),
	}
}

func (c *Cache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[order.OrderUID] = order
}

func (c *Cache) Get(uid string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.data[uid]
	return order, ok
}
