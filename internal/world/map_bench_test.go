package world

import (
	"strings"
	"testing"
)

func BenchmarkGetLine_Diagonal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getLine(0, 0, 20, 20, func(x, y int) bool {
			return true
		})
	}
}

func BenchmarkComputeFOV_SmallRadius(b *testing.B) {
	m := newTestMap(strings.Repeat(".", 80*40))
	m.Width = 80
	m.Height = 40

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ComputeFOV(40, 20, 5, func(x, y int) bool {
			return !m.IsWalkable(x, y)
		})
	}
}

func BenchmarkComputeFOV_LargeRadius(b *testing.B) {
	m := newTestMap(strings.Repeat(".", 100*100))
	m.Width = 100
	m.Height = 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ComputeFOV(50, 50, 20, func(x, y int) bool {
			return !m.IsWalkable(x, y)
		})
	}
}

func BenchmarkMapReset_Large(b *testing.B) {
	m := newTestMap(strings.Repeat(".", 100*100))
	m.Width = 100
	m.Height = 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range m.Tiles {
			m.Tiles[j].Visible = false
		}
	}
}
