package main

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

func (a *App) GitDiff(repoPath string) (string, error) {
	return a.eng.Git.Diff(repoPath)
}

func (a *App) GitLog(repoPath string, limit int) (string, error) {
	return a.eng.Git.Log(repoPath, limit)
}
