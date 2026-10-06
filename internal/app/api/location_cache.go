package api

import (
	"net/http"
	"sync"
	"time"
)

const (
	locationCacheTTL     = 10 * time.Minute
	locationCacheMax     = 500
	locationMinInterval  = time.Second
	locationClientTimout = 8 * time.Second
)

type locationCacheEntry struct {
	body    []byte
	expires time.Time
}

var (
	locationClient = &http.Client{Timeout: locationClientTimout}

	locationCacheMu sync.Mutex
	locationCache   = make(map[string]locationCacheEntry)

	locationThrottleMu sync.Mutex
	locationLastCall   time.Time
)

func locationCacheGet(key string) ([]byte, bool) {
	locationCacheMu.Lock()
	defer locationCacheMu.Unlock()

	entry, ok := locationCache[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expires) {
		delete(locationCache, key)
		return nil, false
	}

	return entry.body, true
}

func locationCacheSet(key string, body []byte) {
	locationCacheMu.Lock()
	defer locationCacheMu.Unlock()

	if len(locationCache) >= locationCacheMax {
		now := time.Now()

		for k, entry := range locationCache {
			if now.After(entry.expires) {
				delete(locationCache, k)
			}
		}

		if len(locationCache) >= locationCacheMax {
			locationCache = make(map[string]locationCacheEntry)
		}
	}

	locationCache[key] = locationCacheEntry{
		body:    body,
		expires: time.Now().Add(locationCacheTTL),
	}
}

func locationThrottle() {
	locationThrottleMu.Lock()
	defer locationThrottleMu.Unlock()

	if wait := locationMinInterval - time.Since(locationLastCall); wait > 0 {
		time.Sleep(wait)
	}

	locationLastCall = time.Now()
}