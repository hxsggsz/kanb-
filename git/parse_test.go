package git

import (
	"testing"
)

func TestParseSimpleDiff(t *testing.T) {
	input := `diff --git a/hello.go b/hello.go
index abc123..def456 100644
--- a/hello.go
+++ b/hello.go
@@ -1,4 +1,5 @@
 package main
 
-func hello() {
-       fmt.Println("hello")
+func hello(name string) {
+       fmt.Println("hello", name)
 }
`

	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.NewPath != "hello.go" {
		t.Errorf("expected hello.go, got %s", f.NewPath)
	}
	if f.Status != "M" {
		t.Errorf("expected M, got %s", f.Status)
	}
	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}
	h := f.Hunks[0]
	if h.OldStart != 1 || h.OldCount != 4 {
		t.Errorf("old start/count: %d,%d", h.OldStart, h.OldCount)
	}
	if h.NewStart != 1 || h.NewCount != 5 {
		t.Errorf("new start/count: %d,%d", h.NewStart, h.NewCount)
	}
	if len(h.Lines) != 7 {
		t.Fatalf("expected 7 lines, got %d", len(h.Lines))
	}
	expected := []struct {
		typ LineType
		old int
		new int
	}{
		{LineContext, 1, 1},
		{LineContext, 2, 2},
		{LineDeleted, 3, lineNumNone},
		{LineDeleted, 4, lineNumNone},
		{LineAdded, lineNumNone, 3},
		{LineAdded, lineNumNone, 4},
		{LineContext, 5, 5},
	}
	for i, exp := range expected {
		got := h.Lines[i]
		if got.Type != exp.typ || got.OldLineNum != exp.old || got.NewLineNum != exp.new {
			t.Errorf("line %d: got type=%d old=%d new=%d, want type=%d old=%d new=%d",
				i, got.Type, got.OldLineNum, got.NewLineNum, exp.typ, exp.old, exp.new)
		}
	}
}

func TestParseNewFile(t *testing.T) {
	input := `diff --git a/new.go b/new.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+
+func newFunc() {}
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if !f.IsNew {
		t.Error("expected IsNew")
	}
	if f.Status != "A" {
		t.Errorf("expected A, got %s", f.Status)
	}
}

func TestParseDeletedFile(t *testing.T) {
	input := `diff --git a/old.go b/old.go
deleted file mode 100644
index abc1234..0000000
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package main
-
-func oldFunc() {
-       fmt.Println("bye")
-}
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if !f.IsDelete {
		t.Error("expected IsDelete")
	}
	if f.Status != "D" {
		t.Errorf("expected D, got %s", f.Status)
	}
	if len(f.Hunks[0].Lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(f.Hunks[0].Lines))
	}
}

func TestParseBinaryFile(t *testing.T) {
	input := `diff --git a/image.png b/image.png
index abc123..def456 100644
Binary files a/image.png and b/image.png differ
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsBinary {
		t.Error("expected IsBinary")
	}
}

func TestParseEmptyDiff(t *testing.T) {
	files, err := parseRawDiff([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestParseMultipleFiles(t *testing.T) {
	input := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-a
+b
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-x
+y
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].NewPath != "a.go" || files[1].NewPath != "b.go" {
		t.Errorf("wrong file paths: %s, %s", files[0].NewPath, files[1].NewPath)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
