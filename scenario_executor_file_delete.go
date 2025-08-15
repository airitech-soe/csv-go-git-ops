package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func main() {
	repoPath := flag.String("repo", "", "Path to the local Git repository")
	scenarioPath := flag.String("scenario", "", "Path to the delete scenario CSV file")
	username := flag.String("username", "", "GitHub username")
	token := flag.String("token", "", "GitHub personal access token")
	flag.Parse()

	if *repoPath == "" || *scenarioPath == "" || *username == "" || *token == "" {
		log.Fatal("All flags --repo, --scenario, --username, and --token are required.")
	}

	// Open local Git repository
	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		log.Fatalf("Failed to open repository at %s: %v", *repoPath, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Failed to get worktree: %v", err)
	}

	// Open scenario CSV
	f, err := os.Open(*scenarioPath)
	if err != nil {
		log.Fatalf("Failed to open scenario CSV: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Track whether any files were successfully deleted
	filesDeleted := 0

	for i, rec := range records {
		if len(rec) < 3 {
			log.Printf("Skipping malformed line %d", i+1)
			continue
		}

		path, opType := rec[0], rec[1]

		if opType != "delete" {
			log.Printf("Skipping non-delete op at line %d", i+1)
			continue
		}

		// Construct absolute path if needed
		fullPath := fmt.Sprintf("%s/%s", *repoPath, path)

		// Delete the file
		err := os.Remove(fullPath)
		if err != nil {
			log.Printf("Failed to delete %s: %v", fullPath, err)
			continue
		}

		// Remove from git index
		_, err = worktree.Remove(path)
		if err != nil {
			log.Printf("Failed to remove from Git index: %v", err)
			continue
		}

		log.Printf("Marked for deletion: %s", path)
		filesDeleted++
	}

	// Skip commit if nothing was deleted
	if filesDeleted == 0 {
		log.Println("No files were deleted. Skipping commit and push.")
		return
	}

	// Commit deletion
	commitMsg := fmt.Sprintf("Deleted %d file(s) as per scenario", filesDeleted)
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

	// Push changes
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: *username, // can be anything except empty
			Password: *token,
		},
	})
	if err != nil {
		log.Fatalf("Failed to push: %v", err)
	}

	log.Println("All changes pushed to remote successfully.")
}