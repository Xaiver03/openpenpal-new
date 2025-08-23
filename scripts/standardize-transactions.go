package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

type TransactionPattern struct {
	Pattern     *regexp.Regexp
	Replacement string
	Description string
}

type FileChange struct {
	FilePath    string
	Changes     []LineChange
	NeedsImport bool
}

type LineChange struct {
	LineNum     int
	OriginalLine string
	NewLine     string
	Pattern     string
}

func main() {
	fmt.Printf("%s🔄 OpenPenPal Transaction Standardization Tool%s\n", colorBlue, colorReset)
	fmt.Printf("Replacing direct db.Begin() calls with standardized TransactionHelper\n\n")

	projectRoot := "../"
	servicesDir := filepath.Join(projectRoot, "backend/internal/services")

	patterns := []TransactionPattern{
		{
			Pattern:     regexp.MustCompile(`(\s*)tx\s*:=\s*s\.db\.WithContext\(ctx\)\.Begin\(\)`),
			Replacement: `$1err := s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {`,
			Description: "Replace ctx-aware transaction",
		},
		{
			Pattern:     regexp.MustCompile(`(\s*)tx\s*:=\s*s\.db\.Begin\(\)`),
			Replacement: `$1err := s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {`,
			Description: "Replace simple transaction",
		},
		{
			Pattern:     regexp.MustCompile(`(\s*)tx\.Rollback\(\)`),
			Replacement: `$1return err // Transaction will auto-rollback`,
			Description: "Replace manual rollback",
		},
		{
			Pattern:     regexp.MustCompile(`(\s*)tx\.Commit\(\)\.Error`),
			Replacement: `$1return nil // Transaction will auto-commit`,
			Description: "Replace manual commit",
		},
		{
			Pattern:     regexp.MustCompile(`(\s*)return\s+tx\.Commit\(\)\.Error`),
			Replacement: `$1return nil // Transaction will auto-commit`,
			Description: "Replace return commit",
		},
	}

	changes, err := processDirectory(servicesDir, patterns)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if len(changes) == 0 {
		fmt.Printf("%s✅ No files need transaction standardization%s\n", colorGreen, colorReset)
		return
	}

	// Display preview
	fmt.Printf("%s📋 Preview of changes:%s\n", colorYellow, colorReset)
	for _, fileChange := range changes {
		fmt.Printf("\n%s📄 %s%s\n", colorCyan, fileChange.FilePath, colorReset)
		if fileChange.NeedsImport {
			fmt.Printf("  %s+ Import: TransactionHelper%s\n", colorGreen, colorReset)
		}
		for _, change := range fileChange.Changes {
			fmt.Printf("  %sLine %d:%s\n", colorPurple, change.LineNum, colorReset)
			fmt.Printf("    %s- %s%s\n", colorRed, strings.TrimSpace(change.OriginalLine), colorReset)
			fmt.Printf("    %s+ %s%s\n", colorGreen, strings.TrimSpace(change.NewLine), colorReset)
		}
	}

	// Ask for confirmation
	fmt.Printf("\n%sApply these changes? (y/n): %s", colorYellow, colorReset)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		fmt.Printf("%sOperation cancelled%s\n", colorYellow, colorReset)
		return
	}

	// Apply changes
	for _, fileChange := range changes {
		if err := applyChanges(fileChange); err != nil {
			fmt.Printf("%sError applying changes to %s: %v%s\n", colorRed, fileChange.FilePath, err, colorReset)
			continue
		}
		fmt.Printf("%s✅ Updated %s%s\n", colorGreen, fileChange.FilePath, colorReset)
	}

	// Generate summary report
	generateReport(changes)
	
	fmt.Printf("\n%s🎉 Transaction standardization completed!%s\n", colorGreen, colorReset)
	fmt.Printf("%sNext steps:%s\n", colorBlue, colorReset)
	fmt.Printf("1. Add TransactionHelper to service constructors\n")
	fmt.Printf("2. Update error handling in affected functions\n")
	fmt.Printf("3. Test the changes thoroughly\n")
	fmt.Printf("4. Review the generated report: transaction_migration_report.md\n")
}

func processDirectory(dir string, patterns []TransactionPattern) ([]FileChange, error) {
	var allChanges []FileChange

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		changes, err := processFile(path, patterns)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", path, err)
		}

		if len(changes.Changes) > 0 {
			allChanges = append(allChanges, changes)
		}

		return nil
	})

	return allChanges, err
}

func processFile(filePath string, patterns []TransactionPattern) (FileChange, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileChange{}, err
	}
	defer file.Close()

	var lines []string
	var changes []LineChange
	needsImport := false

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lines = append(lines, line)

		for _, pattern := range patterns {
			if pattern.Pattern.MatchString(line) {
				newLine := pattern.Pattern.ReplaceAllString(line, pattern.Replacement)
				changes = append(changes, LineChange{
					LineNum:     lineNum,
					OriginalLine: line,
					NewLine:     newLine,
					Pattern:     pattern.Description,
				})
				needsImport = true
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return FileChange{}, err
	}

	return FileChange{
		FilePath:    filePath,
		Changes:     changes,
		NeedsImport: needsImport,
	}, nil
}

func applyChanges(fileChange FileChange) error {
	file, err := os.Open(fileChange.FilePath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Apply changes in reverse order to maintain line numbers
	for i := len(fileChange.Changes) - 1; i >= 0; i-- {
		change := fileChange.Changes[i]
		if change.LineNum <= len(lines) {
			lines[change.LineNum-1] = change.NewLine
		}
	}

	// Add import if needed
	if fileChange.NeedsImport {
		lines = addTransactionHelperImport(lines)
	}

	// Write back to file
	output := strings.Join(lines, "\n")
	return os.WriteFile(fileChange.FilePath, []byte(output), 0644)
}

func addTransactionHelperImport(lines []string) []string {
	// Check if already imported
	for _, line := range lines {
		if strings.Contains(line, "TransactionHelper") {
			return lines // Already imported
		}
	}

	// Find imports section and add comment
	for i, line := range lines {
		if strings.Contains(line, "import (") {
			// Add comment before imports
			comment := "\t// Note: Add TransactionHelper field to service struct and initialize in constructor"
			lines = append(lines[:i+1], append([]string{comment}, lines[i+1:]...)...)
			break
		}
	}

	return lines
}

func generateReport(changes []FileChange) {
	report := `# Transaction Standardization Migration Report

## Summary

This report documents the automatic migration from direct database transaction calls to standardized TransactionHelper usage.

## Changes Applied

`

	totalFiles := len(changes)
	totalChanges := 0

	for _, fileChange := range changes {
		totalChanges += len(fileChange.Changes)
		report += fmt.Sprintf("### %s\n\n", fileChange.FilePath)
		report += fmt.Sprintf("- **Changes**: %d\n", len(fileChange.Changes))
		report += fmt.Sprintf("- **Needs Import**: %t\n\n", fileChange.NeedsImport)

		for _, change := range fileChange.Changes {
			report += fmt.Sprintf("**Line %d**: %s\n", change.LineNum, change.Pattern)
			report += fmt.Sprintf("```go\n// Before:\n%s\n\n// After:\n%s\n```\n\n", 
				strings.TrimSpace(change.OriginalLine), 
				strings.TrimSpace(change.NewLine))
		}
	}

	report += fmt.Sprintf(`## Statistics

- **Total Files Modified**: %d
- **Total Changes Applied**: %d

## Required Manual Steps

1. **Add TransactionHelper to Service Structs**:
   ```go
   type YourService struct {
       db                *gorm.DB
       transactionHelper *services.TransactionHelper  // Add this
       // ... other fields
   }
   ```

2. **Initialize TransactionHelper in Constructors**:
   ```go
   func NewYourService(db *gorm.DB) *YourService {
       return &YourService{
           db:                db,
           transactionHelper: services.NewTransactionHelper(db),  // Add this
           // ... other fields
       }
   }
   ```

3. **Update Function Signatures**:
   - Add context.Context parameter where missing
   - Update error handling for new transaction pattern

4. **Test All Modified Functions**:
   - Verify transaction boundaries work correctly
   - Test error conditions and rollback behavior
   - Ensure performance is not negatively impacted

## Benefits of Standardization

1. **Consistent Error Handling**: Automatic rollback on errors
2. **Panic Safety**: Automatic rollback on panics
3. **Context Awareness**: Proper context propagation
4. **Nested Transactions**: Support for savepoints
5. **Performance Monitoring**: Transaction statistics and monitoring
6. **Isolation Level Control**: Different transaction types for different scenarios

## Migration Checklist

- [ ] Review all modified files
- [ ] Add TransactionHelper to all affected services
- [ ] Update service constructors
- [ ] Test critical transaction scenarios
- [ ] Monitor transaction performance
- [ ] Update documentation

---

*Generated on %s*
`, totalFiles, totalChanges, "2025-01-15")

	err := os.WriteFile("transaction_migration_report.md", []byte(report), 0644)
	if err != nil {
		fmt.Printf("%sWarning: Could not generate report: %v%s\n", colorYellow, err, colorReset)
	} else {
		fmt.Printf("%s📝 Migration report generated: transaction_migration_report.md%s\n", colorGreen, colorReset)
	}
}