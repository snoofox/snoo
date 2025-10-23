package feed

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Provider)
	mu       sync.RWMutex
)

func Register(provider Provider) {
	mu.Lock()
	defer mu.Unlock()
	registry[provider.Type()] = provider
}

func Get(providerType string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()

	provider, ok := registry[providerType]
	if !ok {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
	return provider, nil
}

func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	return types
}
