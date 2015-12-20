package math

import (
	"testing"
)

func TestQuadrangle2Cross(t *testing.T) {
	rect := NewQuadrangle2([]Vec2{
		Vec2{0, 0},
		Vec2{0, 10},
		Vec2{10, 10},
		Vec2{10, 0},
	})
	correct := []Vec2{Vec2{0, 5}, Vec2{5, 10}}
	result := rect.Cross(NewLine2KB(1, 5))
	if len(correct) != len(result) {
		t.Error("Quadrangle2 Cross test invalid result len")
	}
	for i := range correct {
		if correct[i] != result[i] {
			t.Error("Quadrangle2 Cross test invalid result")
		}
	}
}

func TestQuadrangle2HasPoint(t *testing.T) {
	rect := NewQuadrangle2([]Vec2{
		Vec2{0, 0},
		Vec2{0, 10},
		Vec2{10, 10},
		Vec2{10, 0},
	})
	if !rect.HasPoint(Vec2{5, 5}) {
		t.Error("rect.HasPoint(Vec2{5,5}) == false")
	}
	if rect.HasPoint(Vec2{0, -5}) {
		t.Error("rect.HasPoint(Vec2{0,-5}) == true")
	}
}
