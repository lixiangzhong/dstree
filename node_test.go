package dstree

import "testing"

// TestFind
func TestFind(t *testing.T) {
	tree := NewTree[int]()
	tree.Add("www.abc.com", 1)
	tree.Add("next.api.123.com", 2)
	tree.Add("*.123.com", 3)
	tree.Add("www.123.com", 4)
	tree.Add("*.next.123.com", 5)
	tests := []struct {
		name string
		want int
	}{
		{"www.abc.com", 1},
		{"next.api.123.com", 2},
		{"any.api.123.com", 3},
		{"any.123.com", 3},
		{"www.123.com", 4},
		{"any.any.123.com", 3},
		{"any.next.123.com", 5},
		{"any.any.next.123.com", 5},
		{"any.any.any.next.123.com", 5},
		{"www.any.com", 0},
		{"abc.com", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tree.Find(tt.name).Payload(); got != tt.want {
				t.Errorf("Find(%v) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// benchmark Find
func BenchmarkFind(b *testing.B) {
	tree := NewTree[string]()
	tree.Add("www.abc.com", "wwwabc")
	tree.Add("next.api.123.com", "next/api/123")
	tree.Add("*.123.com", "*123")
	tree.Add("www.123.com", "www123")
	tree.Add("*.api.123.com", "*api123")
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			tree.Find("next.api.123.com")
		}
	})
}

// benchmark FindWithLocker
func BenchmarkFindWithLocker(b *testing.B) {
	tree := NewTree[string](WithLocker[string]())
	tree.Add("www.abc.com", "wwwabc")
	tree.Add("next.api.123.com", "next/api/123")
	tree.Add("*.123.com", "*123")
	tree.Add("www.123.com", "www123")
	tree.Add("*.api.123.com", "*api123")
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			tree.Find("next.api.123.com")
		}
	})
}
