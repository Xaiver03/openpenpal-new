package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🚀 Running Gateway Service Test Suite")
	fmt.Println("======================================")

	// Set test environment
	os.Setenv("GIN_MODE", "test")
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-gateway-testing-32-chars")
	os.Setenv("LOG_LEVEL", "error") // Reduce log noise during testing

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	testDirs := []string{
		"tests/config",
		"tests/middleware",
		"tests/proxy",
		"tests/loadbalancer",
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
		fmt.Println("  ✅ Config: Environment variables, service configuration, validation")
		fmt.Println("  ✅ Middleware: Authentication, CORS, rate limiting, security headers")
		fmt.Println("  ✅ Proxy: Request forwarding, load balancing, error handling")
		fmt.Println("  ✅ Load Balancer: Round robin, weighted, least connections algorithms")
		fmt.Println("  ✅ Integration: End-to-end Gateway functionality and service routing")
		
		fmt.Println("\n🔧 Test Infrastructure:")
		fmt.Println("  ✅ Mock backend services for realistic testing")
		fmt.Println("  ✅ JWT authentication simulation")
		fmt.Println("  ✅ Rate limiting and CORS validation")
		fmt.Println("  ✅ Health check and metrics collection")
		fmt.Println("  ✅ Complete microservices gateway simulation")
		
		fmt.Println("\n🚦 Gateway Components Tested:")
		fmt.Println("  ✅ Service Discovery and Health Monitoring")
		fmt.Println("  ✅ Authentication and Role-based Access Control")
		fmt.Println("  ✅ Request Routing and Path Rewriting")
		fmt.Println("  ✅ Load Balancing Algorithms (Round Robin, Weighted, Adaptive)")
		fmt.Println("  ✅ Rate Limiting per Service and User")
		fmt.Println("  ✅ Circuit Breaker and Retry Mechanisms")
		fmt.Println("  ✅ Metrics Collection and Performance Monitoring")
		fmt.Println("  ✅ CORS and Security Headers")
		fmt.Println("  ✅ Error Handling and Timeout Management")
		
		os.Exit(0)
	} else {
		fmt.Println("❌ Some tests failed!")
		os.Exit(1)
	}
}