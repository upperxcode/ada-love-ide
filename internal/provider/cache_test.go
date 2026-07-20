package provider

import (
	"testing"
	"time"
)

func TestFetchCache_SetGet(t *testing.T) {
	c := NewFetchCache(time.Minute)
	want := []DiscoveredModel{{ID: "gpt-4o"}}
	c.Set("openai", want)

	got, ok := c.Get("openai")
	if !ok {
		t.Fatal("miss after Set")
	}
	if len(got) != 1 || got[0].ID != "gpt-4o" {
		t.Errorf("got = %+v", got)
	}
}

func TestFetchCache_TTLExpiry(t *testing.T) {
	c := NewFetchCache(10 * time.Millisecond)
	c.Set("p", []DiscoveredModel{{ID: "x"}})
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.Get("p"); ok {
		t.Error("expected miss after TTL expiry")
	}
}

func TestFetchCache_Invalidate(t *testing.T) {
	c := NewFetchCache(time.Minute)
	c.Set("p", []DiscoveredModel{{ID: "x"}})
	c.Invalidate("p")
	if _, ok := c.Get("p"); ok {
		t.Error("expected miss after Invalidate")
	}
}
