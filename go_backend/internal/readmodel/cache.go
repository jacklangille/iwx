package readmodel

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"iwx/go_backend/internal/domain"
)

type marketCache struct {
	ttl        time.Duration
	mu         sync.RWMutex
	states     map[int64]cachedMarketState
	executions map[string]cachedExecutions
}

type cachedMarketState struct {
	value     domain.MarketState
	expiresAt time.Time
}

type cachedExecutions struct {
	value     []domain.Execution
	expiresAt time.Time
}

func newMarketCache(ttl time.Duration) *marketCache {
	return &marketCache{
		ttl:        ttl,
		states:     map[int64]cachedMarketState{},
		executions: map[string]cachedExecutions{},
	}
}

func (c *marketCache) getState(contractID int64) (domain.MarketState, bool) {
	if c == nil {
		return domain.MarketState{}, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.states[contractID]
	if !ok || time.Now().UTC().After(entry.expiresAt) {
		return domain.MarketState{}, false
	}

	return entry.value, true
}

func (c *marketCache) setState(contractID int64, state domain.MarketState) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.states[contractID] = cachedMarketState{
		value:     state,
		expiresAt: time.Now().UTC().Add(c.ttl),
	}
}

func (c *marketCache) getExecutions(contractID int64, limit int) ([]domain.Execution, bool) {
	if c == nil {
		return nil, false
	}

	key := executionsCacheKey(contractID, limit)

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.executions[key]
	if !ok || time.Now().UTC().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

func (c *marketCache) setExecutions(contractID int64, limit int, executions []domain.Execution) {
	if c == nil {
		return
	}

	key := executionsCacheKey(contractID, limit)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.executions[key] = cachedExecutions{
		value:     executions,
		expiresAt: time.Now().UTC().Add(c.ttl),
	}
}

func (c *marketCache) invalidateContract(contractID int64) {
	if c == nil {
		return
	}

	prefix := strconv.FormatInt(contractID, 10) + ":"

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.states, contractID)
	for key := range c.executions {
		if strings.HasPrefix(key, prefix) {
			delete(c.executions, key)
		}
	}
}

func executionsCacheKey(contractID int64, limit int) string {
	return strconv.FormatInt(contractID, 10) + ":" + strconv.Itoa(limit)
}
