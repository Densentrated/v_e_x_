package git

import (
	"fmt"
	"os"
	"path/filepath"
	"vex-backend/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneRepo clones a git repository and returns a list of all files in the repo
// repoURL should be the full URL to the git repository
func CloneRepo(repoURL string) ([]string, error) {
	clonePath := filepath.Join(config.Config.CloneFolder, filepath.Base(repoURL))

	// Remove the directory if it already exists
	if _, err := os.Stat(clonePath); err == nil {
		if err := os.RemoveAll(clonePath); err != nil {
			return nil, fmt.Errorf("failed to remove existing clone directory: %w", err)
		}
	}

	// Clone the repository
	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL: repoURL,
		Auth: &http.BasicAuth{
			Username: config.Config.GitUser,
			Password: config.Config.GitPAT,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get all files in the cloned repository
	files, err := getAllFiles(clonePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get files from cloned repository: %w", err)
	}

	return files, nil
}

// PullRepo pulls updates from a git repository and returns a list of changed files
// repoURL should be the full URL to the git repository
func PullRepo(repoURL string) ([]string, error) {
	clonePath := filepath.Join(config.Config.CloneFolder, filepath.Base(repoURL))

	// Check if the repository exists
	if _, err := os.Stat(clonePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not found at %s", clonePath)
	}

	// Open the existing repository
	repo, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Pull the latest changes
	err = worktree.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: config.Config.GitUser,
			Password: config.Config.GitPAT,
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("failed to pull repository: %w", err)
	}

	// Get all files in the repository (changed files would be the current state)
	files, err := getAllFiles(clonePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get files from repository: %w", err)
	}

	return files, nil
}

// GetFiles clones the repository if it doesn't exist, or pulls if it does
// Returns the list of changed files (or all files if newly cloned)
// repoURL should be the full URL to the git repository
func GetFiles(repoURL string) ([]string, error) {
	clonePath := filepath.Join(config.Config.CloneFolder, filepath.Base(repoURL))

	// Check if the repository already exists
	if _, err := os.Stat(clonePath); os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		return CloneRepo(repoURL)
	}

	// Repository exists, pull the latest changes
	return PullRepo(repoURL)
}

// getAllFiles returns a list of all files in the repository (excluding .git directory)
func getAllFiles(repoPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Add files only (not directories)
		if !info.IsDir() {
			// Get the relative path from the repo root
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
