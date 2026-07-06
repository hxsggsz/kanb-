package git

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestDiffCommandArgs(t *testing.T) {
	cmd := &DiffCommand{DiffArgs: DiffArgs{Show: false, Args: []string{"--staged"}}}
	args := cmd.Args()
	if len(args) < 2 || args[0] != "diff" {
		t.Errorf("expected diff command, got %v", args)
	}
}

func TestDiffShowCommandArgs(t *testing.T) {
	cmd := &DiffCommand{DiffArgs: DiffArgs{Show: true, Args: []string{"HEAD"}}}
	args := cmd.Args()
	if len(args) < 2 || args[0] != "show" {
		t.Errorf("expected show command, got %v", args)
	}
}

func TestDiffCommandParse(t *testing.T) {
	output := `diff --git a/a.go b/a.go
index abc..def 100644
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-a
+b
`
	cmd := &DiffCommand{
		DiffArgs: DiffArgs{Show: false},
		Parser:  NewUnifiedParser(),
		Aligner: &UnifiedAligner{},
	}
	result, err := cmd.Parse(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if result[0].NewPath != "a.go" {
		t.Errorf("expected a.go, got %s", result[0].NewPath)
	}
	if len(result[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result[0].Hunks))
	}
	if len(result[0].Hunks[0].Lines) != 2 {
		t.Fatalf("expected 2 aligned lines, got %d", len(result[0].Hunks[0].Lines))
	}
}

func TestDiffCommandViaExecute(t *testing.T) {
	output := `diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1 +1 @@
-old
+new
`
	runner := &MockRunner{Output: output}
	cmd := &DiffCommand{
		Parser:  NewUnifiedParser(),
		Aligner: &UnifiedAligner{},
	}
	result, err := Execute(context.Background(), runner, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if len(result[0].Hunks[0].Lines) != 2 {
		t.Fatalf("expected 2 aligned lines, got %d", len(result[0].Hunks[0].Lines))
	}
}

func TestDiffWithGit(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {}\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")
	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n")

	runner := NewGitRunner(dir)
	cmd := &DiffCommand{
		Parser:  NewUnifiedParser(),
		Aligner: &UnifiedAligner{},
	}
	diffs, err := Execute(context.Background(), runner, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected diffs")
	}
}

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %s", args, out)
		}
	}
	return dir
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s", args, out)
	}
}

func writeFile(t *testing.T, dir, path, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/"+path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRepoRootCommand(t *testing.T) {
	cmd := &RepoRootCommand{}
	args := cmd.Args()
	if len(args) != 2 || args[0] != "rev-parse" || args[1] != "--show-toplevel" {
		t.Errorf("unexpected args: %v", args)
	}
	result, err := cmd.Parse("/home/user/repo\n")
	if err != nil {
		t.Fatal(err)
	}
	if result != "/home/user/repo" {
		t.Errorf("expected '/home/user/repo', got %q", result)
	}
}

func TestRepoRootCommandViaExecute(t *testing.T) {
	runner := &MockRunner{Output: "/home/user/repo\n"}
	cmd := &RepoRootCommand{}
	result, err := Execute(context.Background(), runner, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if result != "/home/user/repo" {
		t.Errorf("expected '/home/user/repo', got %q", result)
	}
}

func TestRepoRootIntegration(t *testing.T) {
	dir := setupTestRepo(t)
	runner := NewGitRunner(dir)
	result, err := Execute(context.Background(), runner, &RepoRootCommand{})
	if err != nil {
		t.Fatal(err)
	}
	if result != dir {
		t.Errorf("expected %s, got %s", dir, result)
	}
}
