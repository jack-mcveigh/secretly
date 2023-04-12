package internal

// secretCacheEntry is a map of versions to the secret content
type secretCacheEntry map[string][]byte

// SecretCache contains the cache, mapping secrets to a [secretCacheEntry]
type SecretCache struct {
	cache map[string]secretCacheEntry
}

// NewSecretCache constructs a SecretCache.
func NewSecretCache() SecretCache {
	return SecretCache{cache: make(map[string]secretCacheEntry)}
}

// Add adds a secret with its version to the cache.
func (sc SecretCache) Add(name, version string, content []byte) {
	if sc.cache[name] == nil {
		sc.cache[name] = make(secretCacheEntry)
	}
	sc.cache[name][version] = content
}

// Get gets the secret version from the cache.
// A bool is returned to indicate a cache hit or miss.
func (sc SecretCache) Get(name, version string) ([]byte, bool) {
	if _, ok := sc.cache[name]; !ok {
		return nil, false
	}
	if b, ok := sc.cache[name][version]; ok {
		return b, true
	}
	return nil, false
}
