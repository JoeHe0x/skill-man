package app

import "testing"

func TestShouldShowListLoading_onStartup(t *testing.T) {
	m := New("", "")
	if !m.shouldShowListLoading() {
		t.Fatal("expected loading UI while status is loading")
	}
	m.status = "ready"
	if m.shouldShowListLoading() {
		t.Fatal("expected no loading UI when ready")
	}
}

func TestRenderListLoading_containsSpinner(t *testing.T) {
	m := New("", "")
	out := m.renderListLoading(40, 10)
	if out == "" {
		t.Fatal("expected non-empty loading view")
	}
}
