package internal

type (
	secretCacheEntry map[string][]byte

	SecretCache struct {
		cache map[string]secretCacheEntry
	}
)

func NewSecretCache() SecretCache {
	return SecretCache{cache: make(map[string]secretCacheEntry)}
}

func (sc SecretCache) Add(name, version string, content []byte) {
	if sc.cache[name] == nil {
		sc.cache[name] = make(secretCacheEntry)
	}
	sc.cache[name][version] = content
}

func (sc SecretCache) Get(name, version string) ([]byte, bool) {
	if _, ok := sc.cache[name]; !ok {
		return nil, false
	}
	if b, ok := sc.cache[name][version]; ok {
		return b, true
	}
	return nil, false
}
