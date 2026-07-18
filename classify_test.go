package tui

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// Sample handlers — one canonical implementation per Kind. sampleFor MUST
// return a handler for every Kind in AllKinds(); if a new Kind is appended and
// no sample is added, TestEveryKindHasSample fails. This is the guardrail: the
// contract cannot grow a kind that isn't exercised end-to-end.
// ---------------------------------------------------------------------------

type displaySample struct{}

func (displaySample) Name() string    { return "disp" }
func (displaySample) Content() string { return "hello" }

type editSample struct{}

func (editSample) Name() string    { return "edit" }
func (editSample) Label() string   { return "Edit" }
func (editSample) Value() string   { return "v" }
func (editSample) Change(_ string) {}

type executionSample struct{}

func (executionSample) Name() string  { return "exec" }
func (executionSample) Label() string { return "Run" }
func (executionSample) Execute()      {}

type interactiveSample struct{ editSample }

func (interactiveSample) Name() string         { return "inter" }
func (interactiveSample) WaitingForUser() bool { return false }

type loggableSample struct{}

func (loggableSample) Name() string        { return "log" }
func (loggableSample) SetLog(func(...any)) {}

type selectionSample struct{ editSample }

func (selectionSample) Name() string { return "sel" }
func (selectionSample) Options() []map[string]string {
	return []map[string]string{{"L": "Large"}, {"M": "Medium"}}
}
func (selectionSample) Shortcuts() []map[string]string {
	return []map[string]string{{"l": "Large"}, {"m": "Medium"}}
}

func sampleFor(k Kind) any {
	switch k {
	case KindDisplay:
		return displaySample{}
	case KindEdit:
		return editSample{}
	case KindExecution:
		return executionSample{}
	case KindInteractive:
		return interactiveSample{}
	case KindLoggable:
		return loggableSample{}
	case KindSelection:
		return selectionSample{}
	default:
		return nil
	}
}

// TestKindWireValues freezes the wire protocol: values are contiguous from 0
// and match the documented order. A reorder that silently corrupts the JSON
// protocol fails here.
func TestKindWireValues(t *testing.T) {
	want := []Kind{
		KindDisplay, KindEdit, KindExecution, KindInteractive, KindLoggable, KindSelection,
	}
	got := AllKinds()
	if len(got) != len(want) {
		t.Fatalf("AllKinds length: got %d, want %d", len(got), len(want))
	}
	for i, k := range got {
		if int(k) != i {
			t.Fatalf("kind %s has wire value %d, want %d (values must be contiguous from 0)", k, int(k), i)
		}
		if k != want[i] {
			t.Fatalf("AllKinds order changed at %d: got %s, want %s", i, k, want[i])
		}
	}
}

// TestEveryKindHasSpecAndSample is the completeness guardrail: every declared
// Kind must have exactly one spec and a test sample.
func TestEveryKindHasSpecAndSample(t *testing.T) {
	specByKind := map[Kind]int{}
	for _, s := range specs {
		specByKind[s.kind]++
	}
	for _, k := range AllKinds() {
		if n := specByKind[k]; n != 1 {
			t.Fatalf("kind %s has %d specs, want exactly 1: add/fix it in the specs table", k, n)
		}
		if sampleFor(k) == nil {
			t.Fatalf("kind %s has no sample: add one to sampleFor()", k)
		}
	}
}

// TestClassifyEachKind asserts Classify maps each sample to its Kind and that
// hasField is false only for Loggable.
func TestClassifyEachKind(t *testing.T) {
	for _, k := range AllKinds() {
		gotKind, hasField := Classify(sampleFor(k))
		if gotKind != k {
			t.Errorf("Classify(%s sample) = %s, want %s", k, gotKind, k)
		}
		wantField := k != KindLoggable
		if hasField != wantField {
			t.Errorf("Classify(%s sample) hasField = %v, want %v", k, hasField, wantField)
		}
	}
}

// TestDetectionPrecedence pins the load-bearing ordering: a handler that
// satisfies BOTH Selection and Edit classifies as Selection, and one that
// satisfies BOTH Interactive and Edit classifies as Interactive — never the
// more general Edit.
func TestDetectionPrecedence(t *testing.T) {
	if k, _ := Classify(selectionSample{}); k != KindSelection {
		t.Errorf("Selection+Edit handler classified as %s, want Selection", k)
	}
	if k, _ := Classify(interactiveSample{}); k != KindInteractive {
		t.Errorf("Interactive+Edit handler classified as %s, want Interactive", k)
	}
}

// TestStateEntryRoundTrip asserts a StateEntry built from Extract survives JSON
// round-trip unchanged — in particular Options must survive for Selection, the
// field whose absence caused the original -tui bug.
func TestStateEntryRoundTrip(t *testing.T) {
	h := selectionSample{}
	kind, _ := Classify(h)
	meta := Extract(h)
	in := StateEntry{
		HandlerName: meta.Name,
		HandlerType: int(kind),
		Label:       meta.Label,
		Value:       meta.Value,
		Options:     meta.Options,
		Shortcuts:   meta.Shortcuts,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out StateEntry
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Options) != 2 || out.Options[0]["L"] != "Large" {
		t.Fatalf("Options did not survive round-trip: %#v", out.Options)
	}
	if out.HandlerType != int(KindSelection) {
		t.Fatalf("HandlerType round-trip = %d, want %d", out.HandlerType, int(KindSelection))
	}
}
