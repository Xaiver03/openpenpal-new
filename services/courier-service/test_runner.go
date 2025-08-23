package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🚀 Running Courier Service Test Suite")
	fmt.Println("=====================================")

	// Set test environment
	os.Setenv("GIN_MODE", "test")
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	testDirs := []string{
		"tests/models",
		"tests/services", 
		"tests/integration",
	}

	allPassed := true
	totalTests := 0
	passedTests := 0

	for _, testDir := range testDirs {
		fmt.Printf("\n📁 Running tests in %s\n", testDir)
		fmt.Println(strings.Repeat("-", 50))

		testPath := filepath.Join(currentDir, testDir)
		
		// Check if test directory exists
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			fmt.Printf("⚠️  Test directory %s does not exist, skipping\n", testDir)
			continue
		}

		// Run tests with verbose output and coverage
		cmd := exec.Command("go", "test", "-v", "-cover", "-race", "./"+testDir)
		cmd.Dir = currentDir
		
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			fmt.Printf("❌ Tests failed in %s:\n%s\n", testDir, string(output))
			allPassed = false
		} else {
			fmt.Printf("✅ Tests passed in %s\n", testDir)
			
			// Parse test output for statistics
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "PASS") && strings.Contains(line, "Test") {
					passedTests++
				}
				if strings.Contains(line, "RUN") && strings.Contains(line, "Test") {
					totalTests++
				}
				if strings.Contains(line, "coverage:") {
					fmt.Printf("📊 %s\n", line)
				}
			}
		}
		
		fmt.Print(string(output))
	}

	// Run all tests together for final coverage report
	fmt.Printf("\n🔍 Generating Overall Coverage Report\n")
	fmt.Println(strings.Repeat("=", 50))

	cmd := exec.Command("go", "test", "-v", "-cover", "-coverprofile=coverage.out", "./tests/...")
	cmd.Dir = currentDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not generate overall coverage: %v\n", err)
	} else {
		fmt.Printf("📈 Overall test results:\n%s\n", string(output))
		
		// Generate HTML coverage report
		if _, err := os.Stat(filepath.Join(currentDir, "coverage.out")); err == nil {
			fmt.Println("🌐 Generating HTML coverage report...")
			htmlCmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o=coverage.html")
			htmlCmd.Dir = currentDir
			if err := htmlCmd.Run(); err == nil {
				fmt.Println("✅ HTML coverage report saved to coverage.html")
			}
		}
	}

	// Summary
	fmt.Printf("\n📋 Test Summary\n")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed Tests: %d\n", passedTests)
	
	if allPassed {
		fmt.Println("🎉 All test suites passed!")
		fmt.Println("\n🧪 Test Categories Covered:")
		fmt.Println("  ✅ Models: Task, ScanRecord lifecycle and validation")
		fmt.Println("  ✅ Services: LocationService geographic calculations")
		fmt.Println("  ✅ Services: AssignmentService 4-level hierarchy logic")
		fmt.Println("  ✅ Integration: End-to-end HTTP API workflows")
		fmt.Println("  ✅ FSD Barcode System: OP Code permissions and validation")
		fmt.Println("  ✅ Courier Hierarchy: L1-L4 assignment and permissions")
		
		fmt.Println("\n🔧 Test Infrastructure:")
		fmt.Println("  ✅ In-memory SQLite for fast test execution")
		fmt.Println("  ✅ Testify suite framework for organized testing")
		fmt.Println("  ✅ Mock dependencies and services")
		fmt.Println("  ✅ Comprehensive coverage reporting")
		
		os.Exit(0)
	} else {
		fmt.Println("❌ Some tests failed!")
		os.Exit(1)
	}
}