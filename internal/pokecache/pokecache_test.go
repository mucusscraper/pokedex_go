package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second

	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, cse := range cases {
		t.Run(time.Now().String(), func(t *testing.T) {
			cache := NewCache(interval)

			cache.Add(cse.key, cse.val)

			got, ok := cache.Get(cse.key)
			if !ok {
				t.Fatalf("case %d: expected key to exist", i)
			}
			if string(got) != string(cse.val) {
				t.Fatalf("case %d: expected %q, got %q", i, cse.val, got)
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond

	cache := NewCache(baseTime)

	cache.Add("https://example.com", []byte("testdata"))

	if _, ok := cache.Get("https://example.com"); !ok {
		t.Fatalf("expected key to be present before reap")
	}

	time.Sleep(waitTime)

	if _, ok := cache.Get("https://example.com"); ok {
		t.Fatalf("expected key to be gone after reap")
	}
}
