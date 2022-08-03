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

func RunGitCommand(localDir string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = localDir
	outBytes, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(outBytes))
	return out, err == nil
}

func Checkout(pr prs.PullRequest, m rootModel) tea.Cmd {
	return func() tea.Msg {
		ch := m.async.GetChannel()
		localDir, teaCmd := CheckLocalDir(pr, m)
		if localDir == "" {
			ch <- teaCmd
			return nil
		}
		out, ok := RunGitCommand(localDir, "checkout", pr.Branch)
		ch <- NewToast(out, ok)
		return nil
	}
}

func PullOrigin(pr prs.PullRequest, m rootModel) tea.Cmd {
	return func() tea.Msg {
		localDir, teaCmd := CheckLocalDir(pr, m)
		ch := m.async.GetChannel()
		if localDir == "" {
			ch <- teaCmd
			return nil
		}
		changes, _ := RunGitCommand(localDir, "status", "--short")
		if changes != "" {
			RunGitCommand(localDir, "stash")
			defer RunGitCommand(localDir, "stash", "pop")
		}

		gitCommands := [][]string{
			{"checkout", pr.Branch},
			{"pull", "origin", pr.TargetBranch, "--no-edit"},
			{"push", "--no-verify"},
			{"checkout", "-"},
		}

		for _, args := range gitCommands {
			out, ok := RunGitCommand(localDir, args...)
			ch <- NewToast(out, ok)
			if !ok {
				return nil
			}
		}

		ch <- NewToast("Pulled & pushed", true)
		return nil
	}
}
