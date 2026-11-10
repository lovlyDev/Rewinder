//go:build windows && integration

package hotkey

import (
	"errors"
	"testing"
)

func TestRebindConflictRollsBackOnPumpThread(t *testing.T) {
	original := Combination{Mods: 7, Key: 0x86} // Ctrl+Alt+Shift+F23
	occupied := Combination{Mods: 7, Key: 0x87}
	available := Combination{Mods: 7, Key: 0x85}
	first := &Listener{Initial: original}
	first.Start()
	defer first.Stop()
	<-first.ready
	second := &Listener{Initial: occupied}
	second.Start()
	defer second.Stop()
	<-second.ready
	if err := first.Rebind(available); err != nil {
		t.Fatal(err)
	}
	if err := first.Rebind(original); err != nil {
		t.Fatal(err)
	}
	if err := first.Rebind(occupied); !errors.Is(err, ErrTaken) {
		t.Fatalf("conflict = %v", err)
	}
	if err := second.Rebind(original); !errors.Is(err, ErrTaken) {
		t.Fatalf("old shortcut was not restored: %v", err)
	}
	if err := first.Rebind(available); err != nil {
		t.Fatal(err)
	}
	if err := second.Rebind(original); err != nil {
		t.Fatalf("old shortcut was not released: %v", err)
	}
	first.Stop()
	if err := first.Rebind(original); err == nil {
		t.Fatal("rebind after stop succeeded")
	}
}
