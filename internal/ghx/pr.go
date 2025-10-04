package ghx

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PRInfo represents Pull Request information
type PRInfo struct {
	HeadRefName       string `json:"headRefName"`
	HeadOwner         string `json:"headRepositoryOwner"`
	HeadRepo          string `json:"headRepository"`
	IsCrossRepository bool   `json:"isCrossRepository"`
}

// IsGhAvailable checks if GitHub CLI (gh) is installed
func IsGhAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// GetPRInfo retrieves PR information using gh CLI
func GetPRInfo(prNumber int) (*PRInfo, error) {
	if !IsGhAvailable() {
		return nil, fmt.Errorf("GitHub CLI (gh) not found. Please install: https://cli.github.com/")
	}

	// Get PR info with gh pr view
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", prNumber),
		"--json", "headRefName,headRepositoryOwner,headRepository,isCrossRepository")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh pr view failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON
	var result struct {
		HeadRefName string `json:"headRefName"`
		HeadRepositoryOwner struct {
			Login string `json:"login"`
		} `json:"headRepositoryOwner"`
		HeadRepository struct {
			Name string `json:"name"`
		} `json:"headRepository"`
		IsCrossRepository bool `json:"isCrossRepository"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PR info: %w", err)
	}

	return &PRInfo{
		HeadRefName:       result.HeadRefName,
		HeadOwner:         result.HeadRepositoryOwner.Login,
		HeadRepo:          result.HeadRepository.Name,
		IsCrossRepository: result.IsCrossRepository,
	}, nil
}

// FetchPRBranch fetches the PR branch and creates a local branch
func FetchPRBranch(remote, remoteBranch, localBranch string) error {
	// git fetch <remote> <remoteBranch>:<localBranch>
	cmd := exec.Command("git", "fetch", remote,
		fmt.Sprintf("%s:%s", remoteBranch, localBranch))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If branch already exists, try to update
		if strings.Contains(string(output), "already exists") {
			// Update existing branch
			updateCmd := exec.Command("git", "fetch", remote, remoteBranch)
			if updateErr := updateCmd.Run(); updateErr != nil {
				return fmt.Errorf("failed to update branch: %w", updateErr)
			}

			// Reset local branch to match remote (when not checked out)
			resetCmd := exec.Command("git", "branch", "-f", localBranch,
				fmt.Sprintf("%s/%s", remote, remoteBranch))
			if resetErr := resetCmd.Run(); resetErr != nil {
				return fmt.Errorf("failed to reset branch: %w", resetErr)
			}
			return nil
		}
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetCurrentRemote gets the current remote name (usually "origin")
func GetCurrentRemote() (string, error) {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remotes: %w", err)
	}

	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(remotes) == 0 {
		return "", fmt.Errorf("no remotes configured")
	}

	// Return "origin" if it exists
	for _, remote := range remotes {
		if remote == "origin" {
			return "origin", nil
		}
	}

	// Otherwise return first remote
	return remotes[0], nil
}

// RemoteExists checks if a remote exists
func RemoteExists(remote string) bool {
	cmd := exec.Command("git", "remote", "get-url", remote)
	return cmd.Run() == nil
}

// GetOriginURL gets the URL of origin remote
func GetOriginURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get origin remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsSSHURL checks if the URL is SSH format
func IsSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://")
}

// AddRemote adds a new remote with URL format matching origin
func AddRemote(name, owner, repo string) error {
	var url string

	// Get origin URL format
	originURL, err := GetOriginURL()
	if err == nil && IsSSHURL(originURL) {
		// SSH format
		url = fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	} else {
		// HTTPS format (default)
		url = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	}

	cmd := exec.Command("git", "remote", "add", name, url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}
	return nil
}

// RemoveRemote removes a remote
func RemoveRemote(name string) error {
	cmd := exec.Command("git", "remote", "remove", name)
	return cmd.Run()
}
