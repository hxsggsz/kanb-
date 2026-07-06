package git

import (
	"os"
	"os/exec"
	"testing"
)

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

func TestDiffWithGit(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {}\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n")

	diffs, err := Diff(dir, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected diffs")
	}
	if diffs[0].NewPath != "hello.go" {
		t.Errorf("expected hello.go, got %s", diffs[0].NewPath)
	}
}

func TestDiffStaged(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "a.go", "package main\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	writeFile(t, dir, "a.go", "package main\n\nfunc f() {}\n")
	gitCmd(t, dir, "add", ".")

	diffs, err := Diff(dir, []string{"--staged"})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected staged diffs")
	}
}

func TestDiffNoChanges(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "a.go", "package main\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	diffs, err := Diff(dir, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestRepoRoot(t *testing.T) {
	dir := setupTestRepo(t)
	root, err := RepoRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestRepoRootOutside(t *testing.T) {
	dir := t.TempDir()
	_, err := RepoRoot(dir)
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}
