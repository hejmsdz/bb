package main

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hejmsdz/bb/prs"
)

func CheckLocalDir(pr prs.PullRequest, m rootModel) (string, tea.Cmd) {
	localDir, exists := m.localRepos[pr.Repo]
	if !exists {
		return "", NewErrorToast("Configure local repository path for " + pr.Repo + " first")
	}
	fileInfo, err := os.Stat(localDir)
	if err != nil || !fileInfo.IsDir() {
		return "", NewErrorToast(localDir + " is not a valid directory")
	}
	return localDir, nil
}

func RunGitCommand(localDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = localDir
	outBytes, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(outBytes))

	return out, err
}

func Checkout(pr prs.PullRequest, m rootModel) tea.Cmd {
	localDir, teaCmd := CheckLocalDir(pr, m)
	if localDir == "" {
		return teaCmd
	}
	out, err := RunGitCommand(localDir, "checkout", pr.Branch)
	return NewToast(out, err == nil)
}

func PullOrigin(pr prs.PullRequest, m rootModel) tea.Cmd {
	localDir, teaCmd := CheckLocalDir(pr, m)
	var (
		out string
		err error
	)
	if localDir == "" {
		return teaCmd
	}
	out, _ = RunGitCommand(localDir, "status", "--short")
	isDirty := out != ""
	if isDirty {
		RunGitCommand(localDir, "stash")
	}
	out, err = RunGitCommand(localDir, "checkout", pr.Branch)
	if err != nil {
		return NewErrorToast(out)
	}
	out, err = RunGitCommand(localDir, "pull", "origin", pr.TargetBranch, "--no-edit")
	if err != nil {
		return NewErrorToast(out)
	}
	out, err = RunGitCommand(localDir, "push", "--no-verify")
	if err != nil {
		return NewErrorToast(out)
	}
	out, err = RunGitCommand(localDir, "checkout", "-")
	if err != nil {
		return NewErrorToast(out)
	}
	if isDirty {
		_, err = RunGitCommand(localDir, "stash", "pop")
	}

	return NewToast("Pulled & pushed", err == nil)
}
