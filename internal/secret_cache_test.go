package internal

import (
	"reflect"
	"testing"
)

func newSecretCacheWithEntries() SecretCache {
	sc := NewSecretCache()

	sc.cache["key1"] = secretCacheEntry{
		"1":      []byte("key1: 1: secret content"),
		"latest": []byte("key1: latest: secret content"),
	}
	return sc
}

type secretInfo struct {
	name    string
	version string
	content []byte
}

func TestSecretCacheAdd(t *testing.T) {
	tests := []struct {
		name       string
		secretInfo secretInfo
		want       []byte
	}{
		{
			name: "Miss Name (Add)",
			secretInfo: secretInfo{
				name:    "key3",
				version: "latest",
				content: []byte("key3: latest: secret content"),
			},
			want: []byte("key3: latest: secret content"),
		},
		{
			name: "Miss Version (Add)",
			secretInfo: secretInfo{
				name:    "key1",
				version: "2",
				content: []byte("key1: 2: secret content"),
			},
			want: []byte("key1: 2: secret content"),
		},
		{
			name: "Hit 1 (Update)",
			secretInfo: secretInfo{
				name:    "key1",
				version: "1",
				content: []byte("key1: 1: new secret content"),
			},
			want: []byte("key1: 1: new secret content"),
		},
		{
			name: "Hit latest (Update)",
			secretInfo: secretInfo{
				name:    "key1",
				version: "latest",
				content: []byte("key1: latest: new secret content"),
			},
			want: []byte("key1: latest: new secret content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := newSecretCacheWithEntries()

			sc.Add(tt.secretInfo.name, tt.secretInfo.version, tt.secretInfo.content)

			entry, ok := sc.cache[tt.secretInfo.name]
			if !ok {
				t.Errorf("Missing secret cache entry. Expected an entry for %v", tt.secretInfo)
			}

			got, ok := entry[tt.secretInfo.version]
			if !ok {
				t.Errorf("Missing secret cache entry version. Expected an entry version for %v", tt.want)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Incorrect secret content. Want %v, got %v", string(tt.want), string(got))
			}
		})
	}
}

func TestSecretCacheGet(t *testing.T) {
	tests := []struct {
		name       string
		secretInfo secretInfo
		want       []byte
		wantOk     bool
	}{
		{
			name: "Miss Name (Don't Get)",
			secretInfo: secretInfo{
				name:    "key3",
				version: "1",
			},
			want: nil,
		},
		{
			name: "Miss Version (Don't Get)",
			secretInfo: secretInfo{
				name:    "key1",
				version: "2",
			},
			want: nil,
		},
		{
			name: "Hit (Get)",
			secretInfo: secretInfo{
				name:    "key1",
				version: "1",
			},
			wantOk: true,
			want:   []byte("key1: 1: secret content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := newSecretCacheWithEntries()

			got, ok := sc.Get(tt.secretInfo.name, tt.secretInfo.version)

			if tt.wantOk != ok {
				t.Errorf("Incorrect ok. Want %v, got %v", tt.wantOk, ok)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Incorrect secret content. Want %v, got %v", string(tt.want), string(got))
			}
		})
	}
}
