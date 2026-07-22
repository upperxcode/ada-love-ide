package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (a *App) GitInit(repoPath string) (string, error) {
	return a.eng.Git.Init(repoPath)
}

func (a *App) GitStatus(repoPath string) (string, error) {
	return a.eng.Git.Status(repoPath)
}

func (a *App) GitBranchList(repoPath string) (string, error) {
	return a.eng.Git.BranchList(repoPath)
}

func (a *App) GitBranchCreate(repoPath, branchName string) (string, error) {
	return a.eng.Git.BranchCreate(repoPath, branchName)
}

func (a *App) GitBranchCheckout(repoPath, branchName string) (string, error) {
	return a.eng.Git.BranchCheckout(repoPath, branchName)
}

func (a *App) GitAdd(repoPath, filePath string) (string, error) {
	return a.eng.Git.Add(repoPath, filePath)
}

func (a *App) GitCommit(repoPath, message string) (string, error) {
	return a.eng.Git.Commit(repoPath, message)
}

func (a *App) GitPull(repoPath, remote, branch, authToken string) (string, error) {
	return a.eng.Git.Pull(repoPath, remote, branch, authToken)
}

func (a *App) GitPush(repoPath, remote, branch, authToken string) (string, error) {
	return a.eng.Git.Push(repoPath, remote, branch, authToken)
}

func (a *App) GitRemoteAdd(repoPath, name, url string) (string, error) {
	return a.eng.Git.RemoteAdd(repoPath, name, url)
}

func (a *App) GitRemoteList(repoPath string) (string, error) {
	return a.eng.Git.RemoteList(repoPath)
}

func (a *App) GitInferCommitMessage(repoPath string) (string, error) {
	diff, err := a.eng.Git.Diff(repoPath)
	if err != nil {
		diff = ""
	}
	status, err := a.eng.Git.Status(repoPath)
	if err != nil {
		status = ""
	}
	changedFiles := ""
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Working tree clean" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			changedFiles += "\n- " + parts[len(parts)-1]
		}
	}
	if changedFiles == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	diffLines := strings.Split(diff, "\n")
	if len(diffLines) > 100 {
		diff = strings.Join(diffLines[:100], "\n") + "\n... (diff truncated to 100 lines)"
	}

	prompt := fmt.Sprintf(`Generate a short git commit message (max 72 chars, imperative mood) for these changes:

Files changed:%s

Diff:
%s

Return ONLY the commit message, nothing else.`, changedFiles, diff)

	if a.eng.Chat == nil {
		return a.eng.Git.InferCommitMessage(repoPath)
	}

	tinyProvider, tinyModel, _ := a.eng.DB.GetFixedModel("tinybrain")
	model := ""
	if tinyModel != "" {
		if tinyProvider != "" {
			model = tinyProvider + "/" + tinyModel
		} else {
			model = tinyModel
		}
	}
	if model == "" {
		return a.eng.Git.InferCommitMessage(repoPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg, err := a.eng.Chat.InferQuick(ctx, prompt, model)
	if err != nil {
		fmt.Printf("[GitInferCommitMessage] LLM error: %v, falling back to heuristic\n", err)
		return a.eng.Git.InferCommitMessage(repoPath)
	}
	return msg, nil
}

func (a *App) GitDiff(repoPath string) (string, error) {
	return a.eng.Git.Diff(repoPath)
}

func (a *App) GitLog(repoPath string, limit int) (string, error) {
	return a.eng.Git.Log(repoPath, limit)
}
