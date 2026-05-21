package render

import (
	"strings"
	"sync"
	"testing"
)

func TestMarkdown_concurrentRendersDoNotPanic(t *testing.T) {
	const tableMD = `
| A | B |
|---|---|
| 1 | 2 |
`
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			_, err := Markdown(tableMD, 40+w)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
	}
}

func TestMarkdown_recoversFromBadTableWidth(t *testing.T) {
	md := strings.Repeat("| col |\n|-----|\n| val |\n", 3)
	_, err := Markdown(md, 1)
	if err == nil {
		// narrow width may still succeed; test only ensures no panic
		return
	}
	if !strings.Contains(err.Error(), "panic") && !strings.Contains(err.Error(), "render") {
		t.Fatalf("unexpected error: %v", err)
	}
}
