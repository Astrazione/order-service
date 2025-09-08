package storage

import (
	"container/list"
	"order-service/internal/models"
	"sync"
)

const maxSize = 1000

type Cache struct {
	mu    sync.RWMutex
	data  map[string]models.Order
	queue *list.List // Для отслеживания порядка добавления
}

// orderEntry используется для хранения UID в очереди
type orderEntry struct {
	uid string
}

func NewCache() *Cache {
	return &Cache{
		data:  make(map[string]models.Order),
		queue: list.New(),
	}
}

func (c *Cache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[order.OrderUID]; exists {
		// удаление старой записи
		for e := c.queue.Front(); e != nil; e = e.Next() {
			entry := e.Value.(orderEntry)
			if entry.uid == order.OrderUID {
				c.queue.Remove(e)
				break
			}
		}
	} else if len(c.data) >= maxSize {
		// переполнение кэша
		front := c.queue.Front()
		if front != nil {
			oldest := c.queue.Remove(front).(orderEntry)
			delete(c.data, oldest.uid)
		}
	}

	// добавление элемента в кэш
	c.data[order.OrderUID] = order
	c.queue.PushBack(orderEntry{uid: order.OrderUID})
}

func (c *Cache) Get(uid string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.data[uid]
	return order, ok
}
