//go:build integration

package redis

import (
	"context"
	"os"
	"testing"
	"time"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var testAddr string

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		panic("start redis container: " + err.Error())
	}
	defer func() { _ = container.Terminate(ctx) }()

	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		panic("get endpoint: " + err.Error())
	}

	testAddr = addr
	os.Exit(m.Run())
}

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	c, err := New(context.Background(), testAddr, 0)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestGetSet_Roundtrip(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	if v, err := c.Get(ctx, "missing-key"); err != nil || v != "" {
		t.Errorf("Get(missing) = (%q, %v), want (\"\", nil)", v, err)
	}

	if err := c.Set(ctx, "greeting", "hello", 60); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	v, err := c.Get(ctx, "greeting")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if v != "hello" {
		t.Errorf("Get() = %q, want \"hello\"", v)
	}
}

func TestIncrement_CountsAndSetsTTLOnlyOnFirstCall(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	for i, want := int64(1), int64(1); i <= 3; i, want = i+1, want+1 {
		got, err := c.Increment(ctx, "counter", 60)
		if err != nil {
			t.Fatalf("Increment() #%d error = %v", i, err)
		}
		if got != want {
			t.Errorf("Increment() #%d = %d, want %d", i, got, want)
		}
	}

	ttl, err := c.client.TTL(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl <= 0 || ttl > 60*time.Second {
		t.Errorf("TTL(counter) = %v, want between 0s and 60s", ttl)
	}
}

func TestIncrement_SeparateKeysAreIndependent(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	a, err := c.Increment(ctx, "key-a", 60)
	if err != nil {
		t.Fatalf("Increment(key-a) error = %v", err)
	}
	b, err := c.Increment(ctx, "key-b", 60)
	if err != nil {
		t.Fatalf("Increment(key-b) error = %v", err)
	}
	if a != 1 || b != 1 {
		t.Errorf("a=%d b=%d, want both 1 (independent counters)", a, b)
	}
}
