package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func main() {
	repoPath := flag.String("repo", "", "Path to the local Git repository")
	scenarioPath := flag.String("scenario", "", "Path to the folder delete scenario CSV file")
	username := flag.String("username", "", "GitHub username")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	if *repoPath == "" || *scenarioPath == "" || *username == "" || *token == "" {
		log.Fatal("All flags --repo, --scenario, --username, and --token are required.")
	}

	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Failed to get worktree: %v", err)
	}

	file, err := os.Open(*scenarioPath)
	if err != nil {
		log.Fatalf("Failed to open scenario CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read scenario CSV: %v", err)
	}

	foldersDeleted := 0

	for i, rec := range records {
		if len(rec) < 2 {
			log.Printf("Skipping malformed line %d", i+1)
			continue
		}

		relativePath, opType := rec[0], rec[1]
		if opType != "delete" {
			log.Printf("Skipping non-delete op at line %d", i+1)
			continue
		}

		// Use repoPath to construct full filesystem path
		fullPath := filepath.Join(*repoPath, relativePath)

		// Check if the folder exists
		if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			log.Printf("Folder not found (skipped): %s", fullPath)
			continue
		}

		// First, remove all files in the folder from the Git index
		err = worktree.RemoveGlob(filepath.Join(relativePath, "*"))
		if err != nil {
			log.Printf("Failed to remove from Git index: %v", err)
			continue
		}
		// Then delete from filesystem
		err := os.RemoveAll(fullPath)
		if err != nil {
			log.Printf("Failed to delete folder %s: %v", fullPath, err)
			continue
		}

		log.Printf("Deleted folder: %s", relativePath)
		foldersDeleted++
	}

	if foldersDeleted == 0 {
		log.Println("No folders deleted. Skipping commit and push.")
		return
	}

	commitMsg := fmt.Sprintf("Deleted %d folder(s) as per scenario", foldersDeleted)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  *username,
			Email: fmt.Sprintf("%s@example.com", *username),
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: *username, // this can be anything except empty
			Password: *token,
		},
	})
	if err != nil {
		log.Fatalf("Failed to push: %v", err)
	}

	log.Println("All folder deletions committed and pushed successfully.")
}