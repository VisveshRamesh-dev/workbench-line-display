package main

import (
	"regexp"
	"testing"
)

// TestExtractCodeRealLabel locks the server's part-code extraction against the
// real Zebra DS3678 ISO-15434 payloads the client scans. serial-capture strips
// the non-printable separators (RS/GS/EOT) before forwarding, so the server
// sees the printable line shown here; extractCode must pull the 5-digit group
// after the "P" data identifier.
func TestExtractCodeRealLabel(t *testing.T) {
	codeRegex = regexp.MustCompile(`P(\d{5})`)
	cases := map[string]string{
		// the four real MQ4I door-trim labels from the client's DS3678
		"[)>06VTEIHP82301GB000MQCSB00QT2609100101AB9AXG0148C:M:": "82301",
		"[)>06VTEIHP82302GB000MQCSB00QT2609100101AB9AXG0149C:M:": "82302",
		"[)>06VTEIHP83301GB000MQCSB00QT2609100101AB9AXG0147C:M:": "83301",
		"[)>06VTEIHP83302GB000MQCSB00QT2609100101AB9AXG0146C:M:": "83302",
		// bare part-number field
		"P82302GB000MQC": "82302",
		// pattern matches, but part is unknown -> lookup (not extract) rejects it
		"[)>06VTEIHP99999ZZ000MQCSB00QT2609100101AB9AXG9999C:M:": "99999",
		// no P-prefixed code at all
		"HELLO123": "",
	}
	for in, want := range cases {
		if got := extractCode(in); got != want {
			t.Errorf("extractCode(%q) = %q, want %q", in, got, want)
		}
	}
}
