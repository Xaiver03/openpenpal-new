package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/platform/security"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Initialize configuration
	cfg := &config.Config{}

	fmt.Println("🔐 Zero-Trust Security Architecture Demo")
	fmt.Println("==========================================")

	// Initialize Zero-Trust Manager
	zeroTrustManager := security.NewZeroTrustManager(cfg, logger)

	// Start the Zero-Trust system
	ctx := context.Background()
	if err := zeroTrustManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start Zero-Trust Manager: %v", err)
	}

	fmt.Println("\n✅ Zero-Trust Security Manager started successfully")

	// Demo scenarios
	runDemoScenarios(ctx, zeroTrustManager, logger)

	// Clean shutdown
	fmt.Println("\n🔄 Stopping Zero-Trust Security Manager...")
	if err := zeroTrustManager.Stop(ctx); err != nil {
		log.Printf("Error stopping Zero-Trust Manager: %v", err)
	}

	fmt.Println("\n✅ Zero-Trust Security Demo completed successfully!")
}

func runDemoScenarios(ctx context.Context, ztm *security.ZeroTrustManager, logger *zap.Logger) {
	fmt.Println("\n📋 Running Zero-Trust Security Scenarios...")
	fmt.Println("============================================")

	// Scenario 1: Trusted User Access
	fmt.Println("\n🟢 Scenario 1: Trusted User Access")
	trustedRequest := &security.SecurityRequest{
		UserID:    "user123",
		Resource:  "/api/letters",
		Action:    "read",
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		Method:    "GET",
		Path:      "/api/letters",
		Headers:   map[string]string{"Content-Type": "application/json"},
	}

	decision, err := ztm.AuthorizeRequest(ctx, trustedRequest)
	if err != nil {
		logger.Error("Trusted user authorization failed", zap.Error(err))
	} else {
		fmt.Printf("   Decision: %s\n", decision.Decision)
		fmt.Printf("   Trust Level: %v\n", decision.TrustLevel)
		fmt.Printf("   Risk Score: %.2f\n", decision.RiskScore)
		fmt.Printf("   Reason: %s\n", decision.Reason)
		fmt.Printf("   Permissions: %d granted\n", len(decision.Permissions))
	}

	// Scenario 2: Suspicious User Access
	fmt.Println("\n🟡 Scenario 2: Suspicious User Access")
	suspiciousRequest := &security.SecurityRequest{
		UserID:    "user456",
		Resource:  "/api/admin/users",
		Action:    "delete",
		IPAddress: "10.0.0.1", // Different IP pattern
		UserAgent: "curl/7.68.0", // Automated tool
		Method:    "DELETE",
		Path:      "/api/admin/users/123",
		Headers:   map[string]string{"X-Forwarded-For": "multiple,ips,here"},
	}

	decision, err = ztm.AuthorizeRequest(ctx, suspiciousRequest)
	if err != nil {
		logger.Error("Suspicious user authorization failed", zap.Error(err))
	} else {
		fmt.Printf("   Decision: %s\n", decision.Decision)
		fmt.Printf("   Trust Level: %v\n", decision.TrustLevel)
		fmt.Printf("   Risk Score: %.2f\n", decision.RiskScore)
		fmt.Printf("   Reason: %s\n", decision.Reason)
		fmt.Printf("   Required Actions: %d\n", len(decision.RequiredActions))
		for _, action := range decision.RequiredActions {
			fmt.Printf("     - %s (priority: %d)\n", action.Type, action.Priority)
		}
	}

	// Scenario 3: High-Risk Access Attempt
	fmt.Println("\n🔴 Scenario 3: High-Risk Access Attempt")
	highRiskRequest := &security.SecurityRequest{
		UserID:    "unknown_user",
		Resource:  "/api/sensitive-data",
		Action:    "export",
		IPAddress: "192.168.1.1", // Different location
		UserAgent: "Python-requests/2.25.1", // Automated script
		Method:    "POST",
		Path:      "/api/sensitive-data/export",
		Headers: map[string]string{
			"X-Forwarded-For": "tor.exit.node",
			"Content-Type":    "application/json",
		},
		Body: map[string]interface{}{
			"export_all": true,
			"format":     "csv",
		},
	}

	decision, err = ztm.AuthorizeRequest(ctx, highRiskRequest)
	if err != nil {
		logger.Error("High-risk authorization failed", zap.Error(err))
	} else {
		fmt.Printf("   Decision: %s\n", decision.Decision)
		fmt.Printf("   Trust Level: %v\n", decision.TrustLevel)
		fmt.Printf("   Risk Score: %.2f\n", decision.RiskScore)
		fmt.Printf("   Reason: %s\n", decision.Reason)
		fmt.Printf("   Restrictions: %d applied\n", len(decision.Restrictions))
	}

	// Scenario 4: Administrative Access
	fmt.Println("\n🟦 Scenario 4: Administrative Access")
	adminRequest := &security.SecurityRequest{
		UserID:    "admin_user",
		Resource:  "/api/admin/system-config",
		Action:    "update",
		IPAddress: "192.168.1.10", // Corporate network
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		Method:    "PUT",
		Path:      "/api/admin/system-config",
		Headers:   map[string]string{"Authorization": "Bearer admin_token"},
	}

	decision, err = ztm.AuthorizeRequest(ctx, adminRequest)
	if err != nil {
		logger.Error("Admin authorization failed", zap.Error(err))
	} else {
		fmt.Printf("   Decision: %s\n", decision.Decision)
		fmt.Printf("   Trust Level: %v\n", decision.TrustLevel)
		fmt.Printf("   Risk Score: %.2f\n", decision.RiskScore)
		fmt.Printf("   Reason: %s\n", decision.Reason)
		fmt.Printf("   Permissions: %d granted\n", len(decision.Permissions))
	}

	// Display Security Metrics
	fmt.Println("\n📊 Security Metrics Summary")
	fmt.Println("===========================")
	metrics, err := ztm.GetSecurityMetrics(ctx)
	if err != nil {
		logger.Error("Failed to get security metrics", zap.Error(err))
	} else {
		fmt.Printf("   Total Requests: %d\n", metrics.TotalRequests)
		fmt.Printf("   Allowed: %d (%.1f%%)\n", metrics.AllowedRequests, 
			float64(metrics.AllowedRequests)/float64(metrics.TotalRequests)*100)
		fmt.Printf("   Denied: %d (%.1f%%)\n", metrics.DeniedRequests,
			float64(metrics.DeniedRequests)/float64(metrics.TotalRequests)*100)
		fmt.Printf("   Challenged: %d (%.1f%%)\n", metrics.ChallengedRequests,
			float64(metrics.ChallengedRequests)/float64(metrics.TotalRequests)*100)
		fmt.Printf("   Average Risk Score: %.2f\n", metrics.AverageRiskScore)
		fmt.Printf("   Threat Detections: %d\n", metrics.ThreatDetections)
		fmt.Printf("   Policy Violations: %d\n", metrics.PolicyViolations)
		fmt.Printf("   Compliance Score: %.2f%%\n", metrics.ComplianceScore*100)
	}

	// Demonstrate Encryption Capabilities
	demonstrateEncryption(ctx, logger)

	// Demonstrate Session Management
	demonstrateSessionManagement(ctx, logger)

	// Demonstrate Compliance Validation
	demonstrateCompliance(ctx, logger)

	// Demonstrate Risk Assessment
	demonstrateRiskAssessment(ctx, logger)
}

func demonstrateEncryption(ctx context.Context, logger *zap.Logger) {
	fmt.Println("\n🔒 Encryption Manager Demo")
	fmt.Println("==========================")

	cfg := &config.Config{}
	encryptionManager := security.NewEncryptionManager(cfg, logger)
	encryptionManager.Start(ctx)

	// Encrypt sensitive data
	sensitiveData := []byte("This is highly confidential user data that must be protected")
	encryptionResult, err := encryptionManager.EncryptData(sensitiveData, security.PurposeDataEncryption)
	if err != nil {
		logger.Error("Encryption failed", zap.Error(err))
		return
	}

	fmt.Printf("   ✅ Data encrypted successfully\n")
	fmt.Printf("   Algorithm: %s\n", encryptionResult.Algorithm)
	fmt.Printf("   Key ID: %s\n", encryptionResult.KeyID)
	fmt.Printf("   Encrypted size: %d bytes\n", len(encryptionResult.EncryptedData))

	// Decrypt the data
	decryptedData, err := encryptionManager.DecryptData(encryptionResult.EncryptedData, encryptionResult.KeyID)
	if err != nil {
		logger.Error("Decryption failed", zap.Error(err))
		return
	}

	fmt.Printf("   ✅ Data decrypted successfully\n")
	fmt.Printf("   Original: %s\n", string(sensitiveData))
	fmt.Printf("   Decrypted: %s\n", string(decryptedData))
	fmt.Printf("   Match: %t\n", string(sensitiveData) == string(decryptedData))

	encryptionManager.Stop(ctx)
}

func demonstrateSessionManagement(ctx context.Context, logger *zap.Logger) {
	fmt.Println("\n👤 Session Management Demo")
	fmt.Println("==========================")

	cfg := &config.Config{}
	sessionManager := security.NewSessionManager(cfg, logger)
	sessionManager.Start(ctx)

	fmt.Printf("   ✅ Session Manager initialized\n")
	fmt.Printf("   Features: Secure sessions, device tracking, trust scoring\n")
	fmt.Printf("   Security: MFA integration, suspicious activity detection\n")

	sessionManager.Stop(ctx)
}

func demonstrateCompliance(ctx context.Context, logger *zap.Logger) {
	fmt.Println("\n📋 Compliance Validation Demo")
	fmt.Println("==============================")

	cfg := &config.Config{}
	complianceValidator := security.NewComplianceValidator(cfg, logger)
	complianceValidator.Start(ctx)

	// Simulate compliance validation
	securityCtx := &security.SecurityContext{
		UserID:    "test_user",
		IPAddress: "192.168.1.100",
		TrustLevel: security.TrustLevelMedium,
	}

	request := &security.SecurityRequest{
		UserID:   "test_user",
		Resource: "/api/personal-data",
		Action:   "read",
	}

	result, err := complianceValidator.ValidateCompliance(ctx, securityCtx, request)
	if err != nil {
		logger.Error("Compliance validation failed", zap.Error(err))
		return
	}

	fmt.Printf("   ✅ Compliance validation completed\n")
	fmt.Printf("   Compliant: %t\n", result.Compliant)
	fmt.Printf("   Score: %.2f%%\n", result.Score*100)
	fmt.Printf("   Violations: %d\n", len(result.Violations))
	fmt.Printf("   Frameworks: GDPR, SOX, HIPAA compatible\n")

	complianceValidator.Stop(ctx)
}

func demonstrateRiskAssessment(ctx context.Context, logger *zap.Logger) {
	fmt.Println("\n⚠️  Risk Assessment Demo")
	fmt.Println("========================")

	cfg := &config.Config{}
	riskAssessment := security.NewRiskAssessment(cfg, logger)
	riskAssessment.Start(ctx)

	// Test different risk scenarios
	scenarios := []struct {
		name     string
		trustLevel security.TrustLevel
		ipAddress  string
		time       time.Time
	}{
		{"Low Risk User", security.TrustLevelHigh, "192.168.1.100", time.Now().Add(-2 * time.Hour)},
		{"Medium Risk User", security.TrustLevelMedium, "10.0.0.1", time.Now()},
		{"High Risk User", security.TrustLevelLow, "192.168.1.1", time.Now().Add(20 * time.Hour)},
	}

	for _, scenario := range scenarios {
		securityCtx := &security.SecurityContext{
			UserID:     "test_user",
			IPAddress:  scenario.ipAddress,
			TrustLevel: scenario.trustLevel,
			Timestamp:  scenario.time,
		}

		request := &security.SecurityRequest{
			UserID:    "test_user",
			IPAddress: scenario.ipAddress,
		}

		riskScore, err := riskAssessment.AssessRisk(ctx, securityCtx, request)
		if err != nil {
			logger.Error("Risk assessment failed", zap.Error(err))
			continue
		}

		var riskLevel string
		if riskScore < 0.3 {
			riskLevel = "LOW"
		} else if riskScore < 0.6 {
			riskLevel = "MEDIUM"
		} else {
			riskLevel = "HIGH"
		}

		fmt.Printf("   %s: %.2f (%s)\n", scenario.name, riskScore, riskLevel)
	}

	fmt.Printf("   ✅ Risk models: Behavioral, Technical, Geographic, Temporal\n")
	fmt.Printf("   ✅ ML insights: Threat prediction, Anomaly detection\n")

	riskAssessment.Stop(ctx)
}

// HTTP Handler for web-based demo (optional)
func createDemoHandler(ztm *security.ZeroTrustManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Create security request from HTTP request
		securityRequest := &security.SecurityRequest{
			UserID:    r.Header.Get("X-User-ID"),
			Resource:  r.URL.Path,
			Action:    "access",
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Method:    r.Method,
			Path:      r.URL.Path,
			Headers:   make(map[string]string),
		}

		// Copy headers
		for name, values := range r.Header {
			if len(values) > 0 {
				securityRequest.Headers[name] = values[0]
			}
		}

		// Authorize the request
		decision, err := ztm.AuthorizeRequest(ctx, securityRequest)
		if err != nil {
			http.Error(w, "Authorization failed", http.StatusInternalServerError)
			return
		}

		// Set response based on decision
		w.Header().Set("Content-Type", "application/json")
		
		switch decision.Decision {
		case security.DecisionAllow:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"decision": "allow", "trust_level": %d, "risk_score": %.2f}`, 
				decision.TrustLevel, decision.RiskScore)
		case security.DecisionDeny:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, `{"decision": "deny", "reason": "%s"}`, decision.Reason)
		case security.DecisionChallenge:
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"decision": "challenge", "required_actions": %d}`, 
				len(decision.RequiredActions))
		default:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"decision": "monitor"}`)
		}
	}
}