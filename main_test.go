package main

import (
	"regexp"
	"testing"
)

// Real DS3678 ISO-15434 payloads must parse down to the 5-digit part code.
func TestExtractCodeRealLabel(t *testing.T) {
	codeRegex = regexp.MustCompile(`P(\d{5})`)
	cases := map[string]string{
		"[)>06VTEIHP82301GB000MQCSB00QT2609100101AB9AXG0148C:M:": "82301",
		"[)>06VTEIHP83302GB000MQCSB00QT2609100101AB9AXG0146C:M:": "83302",
		"P82302GB000MQC": "82302",
		"HELLO123":       "",
	}
	for in, want := range cases {
		if got := extractCode(in); got != want {
			t.Errorf("extractCode(%q) = %q, want %q", in, got, want)
		}
	}
}

// The scan-advanced shift register: each Advance pushes onto position 1 and
// shifts the rest one bench forward; position i shows the product's step (i+1).
func TestPipelineShift(t *testing.T) {
	displayList = []*Display{
		NewDisplay(DisplayConfig{ID: "D1"}),
		NewDisplay(DisplayConfig{ID: "D2"}),
		NewDisplay(DisplayConfig{ID: "D3"}),
	}
	line.init(3)

	steps := func(code string) []Step {
		return []Step{{Description: code + "-s1"}, {Description: code + "-s2"}, {Description: code + "-s3"}}
	}
	// expect verifies the (code, stepIndex) at each of the 3 positions ("" = idle).
	expect := func(label string, want [][2]interface{}) {
		snap := line.Snapshot()
		for i, w := range want {
			code, _ := w[0].(string)
			step, _ := w[1].(int)
			if code == "" {
				if snap[i].State != "idle" {
					t.Fatalf("%s: pos %d = %q step %d, want idle", label, i+1, snap[i].PartCode, snap[i].StepIndex)
				}
				continue
			}
			if snap[i].PartCode != code || snap[i].StepIndex != step {
				t.Fatalf("%s: pos %d = %q step %d, want %q step %d", label, i+1, snap[i].PartCode, snap[i].StepIndex, code, step)
			}
		}
	}

	line.Advance("A", "A", steps("A"))
	expect("after A", [][2]interface{}{{"A", 1}, {"", 0}, {"", 0}})

	line.Advance("B", "B", steps("B"))
	expect("after B", [][2]interface{}{{"B", 1}, {"A", 2}, {"", 0}})

	line.Advance("C", "C", steps("C"))
	expect("after C", [][2]interface{}{{"C", 1}, {"B", 2}, {"A", 3}})

	line.Advance("D", "D", steps("D")) // A falls off the end (complete)
	expect("after D", [][2]interface{}{{"D", 1}, {"C", 2}, {"B", 3}})

	line.Reset()
	expect("after reset", [][2]interface{}{{"", 0}, {"", 0}, {"", 0}})
}

// A part with fewer steps than the line length shows idle once it passes its
// last step, instead of repeating or erroring.
func TestPipelineFewerSteps(t *testing.T) {
	displayList = []*Display{
		NewDisplay(DisplayConfig{ID: "D1"}),
		NewDisplay(DisplayConfig{ID: "D2"}),
		NewDisplay(DisplayConfig{ID: "D3"}),
	}
	line.init(3)
	two := []Step{{Description: "a"}, {Description: "b"}} // only 2 steps
	line.Advance("X", "X", two)
	line.Advance("Y", "Y", two)
	line.Advance("Z", "Z", two) // X now at position 3 but has no step 3 -> idle
	snap := line.Snapshot()
	if snap[2].State != "idle" {
		t.Fatalf("pos 3 with a 2-step part = %q, want idle", snap[2].State)
	}
	if snap[0].PartCode != "Z" || snap[1].PartCode != "Y" {
		t.Fatalf("unexpected pipeline: %q, %q", snap[0].PartCode, snap[1].PartCode)
	}
}
