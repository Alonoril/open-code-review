package session

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestCodexRecorderPersistsCorrelatedReadOnlyRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repository := t.TempDir()
	recorder, err := OpenCodexRecorder(repository, "run-123", "sha256:bundle")
	if err != nil {
		t.Fatalf("OpenCodexRecorder() error = %v", err)
	}
	if err := recorder.Record("prepare", CodexEvent{
		Files: 3, Warnings: 1, Partial: true, DurationMS: 25,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := recorder.Finalize(CodexEvent{Findings: 2, ValidationValid: boolPointer(true)}); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}

	info, err := os.Stat(recorder.Path())
	if err != nil {
		t.Fatalf("stat session: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("session mode = %o, want 600", info.Mode().Perm())
	}
	file, err := os.Open(recorder.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var records []map[string]any
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	if len(records) != 3 {
		t.Fatalf("records = %d, want start, event, end", len(records))
	}
	if records[0]["controlPlane"] != "codex-owned" ||
		records[0]["bundleId"] != "sha256:bundle" ||
		records[0]["tokenUsage"] != "not_available" {
		t.Fatalf("session start = %+v", records[0])
	}
	if records[1]["event"] != "prepare" || records[2]["type"] != "session_end" {
		t.Fatalf("records = %+v", records)
	}
}

func boolPointer(value bool) *bool {
	return &value
}
