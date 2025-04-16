package postgres

import (
	"sync"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
)

type Cache struct {
	mu             sync.RWMutex
	pvzCache       map[uuid.UUID]*models.PVZ
	receptionCache map[uuid.UUID]*models.Reception
	productCache   map[uuid.UUID]*models.Product
}

func NewCache() *Cache {
	return &Cache{
		pvzCache:       make(map[uuid.UUID]*models.PVZ),
		receptionCache: make(map[uuid.UUID]*models.Reception),
		productCache:   make(map[uuid.UUID]*models.Product),
	}
}

func (c *Cache) GetPVZ(id uuid.UUID) (*models.PVZ, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pvz, ok := c.pvzCache[id]
	return pvz, ok
}

func (c *Cache) SetPVZ(pvz *models.PVZ) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pvzCache[pvz.ID] = pvz
}

func (c *Cache) GetReception(id uuid.UUID) (*models.Reception, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	reception, ok := c.receptionCache[id]
	return reception, ok
}

func (c *Cache) SetReception(reception *models.Reception) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.receptionCache[reception.ID] = reception
}

func (c *Cache) GetProduct(id uuid.UUID) (*models.Product, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	product, ok := c.productCache[id]
	return product, ok
}

func (c *Cache) SetProduct(product *models.Product) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.productCache[product.ID] = product
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pvzCache = make(map[uuid.UUID]*models.PVZ)
	c.receptionCache = make(map[uuid.UUID]*models.Reception)
	c.productCache = make(map[uuid.UUID]*models.Product)
}
