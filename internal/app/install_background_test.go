package app

import "testing"

func TestNextInstallProgressPercent_keepsMovingPast90(t *testing.T) {
	p := 0.9
	moved := 0
	for i := 0; i < 50 && p < 0.97; i++ {
		next := nextInstallProgressPercent(p)
		if next > p {
			moved++
		}
		p = next
	}
	if moved < 3 {
		t.Fatalf("expected several steps past 90%%, only moved %d times, final=%.2f", moved, p)
	}
	if p < 0.95 {
		t.Fatalf("expected to creep past 95%%, got %.2f", p)
	}
}

func TestNextInstallProgressPercent_capsBeforeComplete(t *testing.T) {
	p := 0.97
	if next := nextInstallProgressPercent(p); next != p {
		t.Fatalf("expected cap at 97%%, got %.2f -> %.2f", p, next)
	}
}
