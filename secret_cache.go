package secretly

// SecretCache describes a secret cache,
// which are used to limit calls
// to the upstream secret manager service.
type SecretCache interface {
	// Add adds a secret with its version to the cache.
	Add(name, version string, content []byte)

	// Get gets the secret version from the cache.
	// A bool is returned to indicate a cache hit or miss.
	Get(name, version string) ([]byte, bool)
}

type noOpSecretCache struct{}

// NewNoOpSecretCache constructs a no-op secret cache,
// meant to be used for disabling secret caching.
func NewNoOpSecretCache() noOpSecretCache { return noOpSecretCache{} }

func (noOpSecretCache) Add(name, version string, content []byte) {}
func (noOpSecretCache) Get(name, version string) ([]byte, bool)  { return nil, false }

// secretCacheEntry is a map of versions to the secret content.
type secretCacheEntry map[string][]byte

// secretCache contains the cache, mapping secrets to a [secretCacheEntry].
type secretCache struct {
	cache map[string]secretCacheEntry
}

// NewSecretCache constructs a secretCache.
func NewSecretCache() secretCache {
	return secretCache{cache: make(map[string]secretCacheEntry)}
}

func (sc secretCache) Add(name, version string, content []byte) {
	if sc.cache[name] == nil {
		sc.cache[name] = make(secretCacheEntry)
	}
	sc.cache[name][version] = content
}

func (sc secretCache) Get(name, version string) ([]byte, bool) {
	if _, ok := sc.cache[name]; !ok {
		return nil, false
	}
	if b, ok := sc.cache[name][version]; ok {
		return b, true
	}
	return nil, false
}
