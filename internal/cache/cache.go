package cache

import (
	"orderHub_L0/internal/model"
	"sync"
)

type Cache struct {
	orders map[string]model.Order
	mu     sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]model.Order),
	}
}

func (c *Cache) Set(order model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(orderUID string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.Unlock()
	order, exists := c.orders[orderUID]
	return order, exists
}
