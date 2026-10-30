package root

import "testing"

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil, nil)
	if repo == nil {
		t.Fatal("expected NewRepository to return a non-nil repository")
	}
}
