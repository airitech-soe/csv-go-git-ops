package main

import (
	"encoding/base64"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ScenarioOperation struct {
	FilePath      string
	OperationType string
	CommitMessage string
	FileContent   string
	LineNumber    int
}

func main() {
	var repoPath, scenarioPath, logPath, githubUsername, githubToken string
	
	flag.StringVar(&repoPath, "repo", "", "Path to git repository")
	flag.StringVar(&scenarioPath, "scenario", "", "Path to scenario CSV file")
	flag.StringVar(&logPath, "log", "execution_o.log", "Path to log file")
	flag.StringVar(&githubUsername, "username", "", "GitHub username")
	flag.StringVar(&githubToken, "token", "", "GitHub personal access token")
	flag.Parse()
	
	if repoPath == "" || scenarioPath == "" || githubUsername == "" || githubToken == "" {
		fmt.Println("Usage: git_scenario_execute --repo /path/to/repo --scenario /path/to/scenario.csv --username <github_username> --token <github_token> [--log /path/to/log]")
		os.Exit(1)
	}
	
	// Setup logging
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	
	logger := log.New(logFile, "", 0)
	
	// Log execution start
	logger.Printf("[%s] === Scenario execution started ===", time.Now().Format("2006-01-02 15:04:05"))
	logger.Printf("[%s] Repository: %s", time.Now().Format("2006-01-02 15:04:05"), repoPath)
	logger.Printf("[%s] Scenario file: %s", time.Now().Format("2006-01-02 15:04:05"), scenarioPath)
	logger.Printf("[%s] GitHub username: %s", time.Now().Format("2006-01-02 15:04:05"), githubUsername)
	
	// Read scenario CSV
	operations, err := readScenarioCSV(scenarioPath)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to read scenario file: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		fmt.Printf("Error reading scenario file: %v\n", err)
		os.Exit(1)
	}
	
	logger.Printf("[%s] Total operations to execute: %d", time.Now().Format("2006-01-02 15:04:05"), len(operations))
	
	// Change to repository directory
	err = os.Chdir(repoPath)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to change to repository directory: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		fmt.Printf("Error changing to repository directory: %v\n", err)
		os.Exit(1)
	}
	
	// Configure git credentials (after changing to repo directory)
	err = configureGitCredentials(githubUsername, githubToken, logger)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to configure git credentials: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		fmt.Printf("Error configuring git credentials: %v\n", err)
		os.Exit(1)
	}
	
	// Configure git credentials (after changing to repo directory)
	err = configureGitCredentials(githubUsername, githubToken, logger)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to configure git credentials: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		fmt.Printf("Error configuring git credentials: %v\n", err)
		os.Exit(1)
	}
	
	// Execute operations
	successCount := 0
	for _, op := range operations {
		logger.Printf("[%s] --- Executing line %d ---", time.Now().Format("2006-01-02 15:04:05"), op.LineNumber)
		
		success := executeOperation(op, logger, scenarioPath)
		if success {
			successCount++
			logger.Printf("[%s] Line %d completed successfully", time.Now().Format("2006-01-02 15:04:05"), op.LineNumber)
		} else {
			logger.Printf("[%s] Line %d failed", time.Now().Format("2006-01-02 15:04:05"), op.LineNumber)
		}
		
		// Add a small delay between operations
		time.Sleep(100 * time.Millisecond)
	}
	
	logger.Printf("[%s] === Scenario execution completed ===", time.Now().Format("2006-01-02 15:04:05"))
	logger.Printf("[%s] Success: %d/%d operations", time.Now().Format("2006-01-02 15:04:05"), successCount, len(operations))
	
	fmt.Printf("Execution completed. Success: %d/%d operations\n", successCount, len(operations))
	fmt.Printf("Check log file for details: %s\n", logPath)
}

func readScenarioCSV(filename string) ([]ScenarioOperation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	// Make CSV parsing more flexible
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.TrimLeadingSpace = true
	
	var operations []ScenarioOperation
	lineNumber := 1
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV parsing error at line %d: %v", lineNumber, err)
		}
		
		// Skip empty lines
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			lineNumber++
			continue
		}
		
		// Validate minimum required fields
		if len(record) < 3 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected at least 3 columns, got %d columns. Record: %v", lineNumber, len(record), record)
		}
		
		// Trim whitespace from all fields
		for i := range record {
			record[i] = strings.TrimSpace(record[i])
		}
		
		op := ScenarioOperation{
			FilePath:      record[0],
			OperationType: record[1],
			CommitMessage: record[2],
			LineNumber:    lineNumber,
		}
		
		// Validate operation type
		if op.OperationType != "create" && op.OperationType != "update" {
			return nil, fmt.Errorf("invalid operation type '%s' at line %d: must be 'create' or 'update'", op.OperationType, lineNumber)
		}
		
		// Add file content if available (for update operations)
		if len(record) > 3 && strings.TrimSpace(record[3]) != "" {
			op.FileContent = record[3]
		} else if op.OperationType == "update" {
			// Default content for update operations if not specified
			op.FileContent = "test data"
		}
		
		operations = append(operations, op)
		lineNumber++
	}
	
	return operations, nil
}

func executeOperation(op ScenarioOperation, logger *log.Logger, scenarioFile string) bool {
	logger.Printf("[%s] Operation: %s on %s", time.Now().Format("2006-01-02 15:04:05"), op.OperationType, op.FilePath)
	
	// Step 1: Pull
	if !executeGitCommand("pull", logger, scenarioFile, op.LineNumber) {
		return false
	}
	
	// Step 2: Execute the operation
	var success bool
	switch op.OperationType {
	case "create":
		success = executeCreateOperation(op, logger, scenarioFile)
	case "update":
		success = executeUpdateOperation(op, logger, scenarioFile)
	default:
		logger.Printf("[%s] ERROR: Unknown operation type: %s (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), op.OperationType, scenarioFile, op.LineNumber)
		return false
	}
	
	if !success {
		return false
	}
	
	// Step 3: Add and commit
	if !executeGitCommand(fmt.Sprintf("add %s", op.FilePath), logger, scenarioFile, op.LineNumber) {
		return false
	}
	
	// Check if there are any changes to commit
	if !hasChangesToCommit(logger, scenarioFile, op.LineNumber) {
		logger.Printf("[%s] No changes to commit for %s, skipping commit", time.Now().Format("2006-01-02 15:04:05"), op.FilePath)
		return true
	}
	
	if !executeGitCommand(fmt.Sprintf("commit -m %q", op.CommitMessage), logger, scenarioFile, op.LineNumber) {
		return false
	}
	
	// Step 4: Push
	if !executeGitCommand("push", logger, scenarioFile, op.LineNumber) {
		return false
	}
	
	return true
}

func configureGitCredentials(username, token string, logger *log.Logger) error {
	logger.Printf("[%s] Configuring git credentials with HTTP Basic Auth...", time.Now().Format("2006-01-02 15:04:05"))
	
	// Set git config for the current repository (local config)
	cmd := exec.Command("git", "config", "--local", "user.name", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("[%s] ERROR: Git config user.name output: %s", time.Now().Format("2006-01-02 15:04:05"), string(output))
		return fmt.Errorf("failed to set git user.name: %v", err)
	}
	
	// Set default email
	email := username + "@users.noreply.github.com"
	cmd = exec.Command("git", "config", "--local", "user.email", email)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Printf("[%s] ERROR: Git config user.email output: %s", time.Now().Format("2006-01-02 15:04:05"), string(output))
		return fmt.Errorf("failed to set git user.email: %v", err)
	}
	
	logger.Printf("[%s] Set git user.name: %s", time.Now().Format("2006-01-02 15:04:05"), username)
	logger.Printf("[%s] Set git user.email: %s", time.Now().Format("2006-01-02 15:04:05"), email)
	
	// Configure HTTP Basic Auth for GitHub
	// Set credential helper to store credentials
	cmd = exec.Command("git", "config", "--local", "credential.helper", "store")
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Printf("[%s] ERROR: Git config credential.helper output: %s", time.Now().Format("2006-01-02 15:04:05"), string(output))
		return fmt.Errorf("failed to set credential helper: %v", err)
	}
	
	// Configure HTTP Basic Auth specifically for github.com
	cmd = exec.Command("git", "config", "--local", "http.https://github.com/.extraheader", fmt.Sprintf("Authorization: Basic %s", encodeBasicAuth(username, token)))
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Printf("[%s] ERROR: Git config http auth output: %s", time.Now().Format("2006-01-02 15:04:05"), string(output))
		return fmt.Errorf("failed to set HTTP basic auth: %v", err)
	}
	
	// Alternative approach: Set credential.username and use askpass helper
	cmd = exec.Command("git", "config", "--local", "credential.https://github.com.username", username)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Printf("[%s] WARNING: Failed to set credential username: %v", time.Now().Format("2006-01-02 15:04:05"), err)
	}
	
	// Ensure we're using HTTPS URL (not SSH)
	err = ensureHTTPSRemote(logger)
	if err != nil {
		logger.Printf("[%s] WARNING: Failed to ensure HTTPS remote: %v", time.Now().Format("2006-01-02 15:04:05"), err)
	}
	
	// Create credential file for git credential store
	err = createCredentialFile(username, token, logger)
	if err != nil {
		logger.Printf("[%s] WARNING: Failed to create credential file: %v", time.Now().Format("2006-01-02 15:04:05"), err)
	}
	
	logger.Printf("[%s] Git HTTP Basic Auth configured successfully", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func encodeBasicAuth(username, token string) string {
	auth := username + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func ensureHTTPSRemote(logger *log.Logger) error {
	// Set the specific GitHub repository URL
	targetURL := "https://github.com/airitech-soe/csv-go-git-ops.git"
	
	// Get current remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		logger.Printf("[%s] No remote origin found, adding it...", time.Now().Format("2006-01-02 15:04:05"))
		// Add the remote if it doesn't exist
		cmd = exec.Command("git", "remote", "add", "origin", targetURL)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add remote origin: %v, output: %s", err, string(output))
		}
		logger.Printf("[%s] Added remote origin: %s", time.Now().Format("2006-01-02 15:04:05"), targetURL)
		return nil
	}
	
	remoteURL := strings.TrimSpace(string(output))
	logger.Printf("[%s] Current remote URL: %s", time.Now().Format("2006-01-02 15:04:05"), remoteURL)
	
	// Always set to our target URL to ensure consistency
	if remoteURL != targetURL {
		cmd = exec.Command("git", "remote", "set-url", "origin", targetURL)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to set remote URL: %v, output: %s", err, string(output))
		}
		logger.Printf("[%s] Updated remote URL to: %s", time.Now().Format("2006-01-02 15:04:05"), targetURL)
	} else {
		logger.Printf("[%s] Remote URL is already correct", time.Now().Format("2006-01-02 15:04:05"))
	}
	
	return nil
}

func createCredentialFile(username, token string, logger *log.Logger) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	
	credentialFile := filepath.Join(homeDir, ".git-credentials")
	// Use the specific repository URL
	credentialEntry := fmt.Sprintf("https://%s:%s@github.com/airitech-soe/csv-go-git-ops.git\n", username, token)
	
	// Check if file exists and if our entry is already there
	if _, err := os.Stat(credentialFile); err == nil {
		content, err := os.ReadFile(credentialFile)
		if err == nil && strings.Contains(string(content), "github.com/airitech-soe/csv-go-git-ops") {
			logger.Printf("[%s] Credential file already contains entry for csv-go-git-ops repository", time.Now().Format("2006-01-02 15:04:05"))
			return nil
		}
	}
	
	// Append our credentials to the file
	file, err := os.OpenFile(credentialFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open credential file: %v", err)
	}
	defer file.Close()
	
	_, err = file.WriteString(credentialEntry)
	if err != nil {
		return fmt.Errorf("failed to write to credential file: %v", err)
	}
	
	logger.Printf("[%s] Created/updated git credential file for repository: https://github.com/airitech-soe/csv-go-git-ops.git", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func executeCreateOperation(op ScenarioOperation, logger *log.Logger, scenarioFile string) bool {
	// Create directory if it doesn't exist
	dir := filepath.Dir(op.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to create directory %s: %v (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), dir, err, scenarioFile, op.LineNumber)
		return false
	}
	
	// Create empty file
	file, err := os.Create(op.FilePath)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to create file %s: %v (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), op.FilePath, err, scenarioFile, op.LineNumber)
		return false
	}
	defer file.Close()
	
	logger.Printf("[%s] Created file: %s", time.Now().Format("2006-01-02 15:04:05"), op.FilePath)
	return true
}

func executeUpdateOperation(op ScenarioOperation, logger *log.Logger, scenarioFile string) bool {
	// Check if file exists
	if _, err := os.Stat(op.FilePath); os.IsNotExist(err) {
		logger.Printf("[%s] ERROR: File does not exist for update: %s (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), op.FilePath, scenarioFile, op.LineNumber)
		return false
	}
	
	// Read current content to check if update is needed
	currentContent, err := os.ReadFile(op.FilePath)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to read current file content %s: %v (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), op.FilePath, err, scenarioFile, op.LineNumber)
		return false
	}
	
	// Check if content is already the same
	if string(currentContent) == op.FileContent {
		logger.Printf("[%s] File %s already has the same content, no update needed", 
			time.Now().Format("2006-01-02 15:04:05"), op.FilePath)
		return true
	}
	
	// Write content to file
	err = os.WriteFile(op.FilePath, []byte(op.FileContent), 0644)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to update file %s: %v (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), op.FilePath, err, scenarioFile, op.LineNumber)
		return false
	}
	
	logger.Printf("[%s] Updated file: %s with content: %s", time.Now().Format("2006-01-02 15:04:05"), op.FilePath, op.FileContent)
	return true
}

func hasChangesToCommit(logger *log.Logger, scenarioFile string, lineNumber int) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	
	// git diff --cached --quiet returns:
	// - exit code 0 if no changes (no differences)
	// - exit code 1 if there are changes
	// - other exit codes for errors
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Exit code 1 means there are staged changes
			return true
		}
		// Other exit codes indicate an error
		logger.Printf("[%s] WARNING: Error checking for staged changes: %v (scenario: %s, line: %d)", 
			time.Now().Format("2006-01-02 15:04:05"), err, scenarioFile, lineNumber)
		return true // Assume there are changes to be safe
	}
	
	// Exit code 0 means no staged changes
	return false
}

func executeGitCommand(gitCmd string, logger *log.Logger, scenarioFile string, lineNumber int) bool {
	// Parse the command more carefully to handle quotes properly
	parts := parseGitCommand(gitCmd)
	cmd := exec.Command("git", parts...)
	
	logger.Printf("[%s] Executing: git %s", time.Now().Format("2006-01-02 15:04:05"), gitCmd)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Special handling for common git errors
		outputStr := string(output)
		if strings.Contains(outputStr, "nothing to commit") || strings.Contains(outputStr, "no changes added to commit") {
			logger.Printf("[%s] INFO: No changes to commit - this is expected in some cases", time.Now().Format("2006-01-02 15:04:05"))
			return true
		}
		
		logger.Printf("[%s] ERROR: Git command failed: git %s", time.Now().Format("2006-01-02 15:04:05"), gitCmd)
		logger.Printf("[%s] ERROR: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		logger.Printf("[%s] ERROR: Output: %s", time.Now().Format("2006-01-02 15:04:05"), outputStr)
		logger.Printf("[%s] ERROR: Scenario file: %s, Line: %d", time.Now().Format("2006-01-02 15:04:05"), scenarioFile, lineNumber)
		return false
	}
	
	if len(output) > 0 {
		logger.Printf("[%s] Git output: %s", time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(string(output)))
	}
	
	return true
}

func parseGitCommand(gitCmd string) []string {
	// Handle commands with quoted arguments properly
	var parts []string
	var current strings.Builder
	inQuotes := false
	escaped := false
	
	for _, char := range gitCmd {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}
		
		if char == '\\' {
			escaped = true
			continue
		}
		
		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
			} else {
				inQuotes = false
			}
			continue
		}
		
		if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
	}
	
	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
}