package secretly

// secretCacheEntry is a map of versions to the secret content.
type secretCacheEntry map[string][]byte

// cache contains the cache, mapping secrets to a [secretCacheEntry].
type cache struct {
	cache map[string]secretCacheEntry
}

// newCache constructs a secretCache.
func newCache() *cache {
	return &cache{cache: make(map[string]secretCacheEntry)}
}

func (sc cache) Add(name, version string, content []byte) {
	if sc.cache[name] == nil {
		sc.cache[name] = make(secretCacheEntry)
	}
	sc.cache[name][version] = content
}

func (sc cache) Get(name, version string) ([]byte, bool) {
	if _, ok := sc.cache[name]; !ok {
		return nil, false
	}
	if b, ok := sc.cache[name][version]; ok {
		return b, true
	}
	return nil, false
}
