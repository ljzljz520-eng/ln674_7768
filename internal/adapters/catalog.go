package adapters

import (
	"context"
	"fmt"
	"time"

	"campus-device-hub/internal/domain"
)

type Catalog struct {
	items map[string]VendorAdapter
}

func NewCatalog(items []VendorAdapter) *Catalog {
	catalog := &Catalog{items: make(map[string]VendorAdapter, len(items))}
	for _, item := range items {
		if item != nil {
			catalog.items[item.Name()] = item
		}
	}
	return catalog
}

func (c *Catalog) Names() []string {
	names := make([]string, 0, len(c.items))
	for name := range c.items {
		names = append(names, name)
	}
	return names
}

func (c *Catalog) Fetch(ctx context.Context, name string, now time.Time) ([]domain.Device, error) {
	adapter, ok := c.items[name]
	if !ok {
		return nil, fmt.Errorf("vendor %q is not registered", name)
	}
	return adapter.Fetch(ctx, now)
}

func (c *Catalog) All(ctx context.Context, now time.Time) (map[string][]domain.Device, map[string]error) {
	devices := make(map[string][]domain.Device, len(c.items))
	errors := make(map[string]error)
	for name := range c.items {
		items, err := c.Fetch(ctx, name, now)
		if err != nil {
			errors[name] = err
			continue
		}
		devices[name] = items
	}
	return devices, errors
}
