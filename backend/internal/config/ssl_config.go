package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// SSLConfig PostgreSQL SSL配置
type SSLConfig struct {
	// Mode SSL模式: disable, allow, prefer, require, verify-ca, verify-full
	Mode string `json:"mode" yaml:"mode"`
	
	// 证书文件路径
	CAFile     string `json:"ca_file" yaml:"ca_file"`         // CA证书
	CertFile   string `json:"cert_file" yaml:"cert_file"`     // 客户端证书
	KeyFile    string `json:"key_file" yaml:"key_file"`       // 客户端私钥
	CRLFile    string `json:"crl_file" yaml:"crl_file"`       // 证书吊销列表
	
	// 连接参数
	ServerName string `json:"server_name" yaml:"server_name"` // 服务器名称验证
	
	// 环境特定配置
	Environment string `json:"environment" yaml:"environment"`
}

// SSLMode SSL模式常量
const (
	SSLModeDisable    = "disable"     // 不使用SSL
	SSLModeAllow      = "allow"       // 尝试非SSL，如果失败则尝试SSL
	SSLModePrefer     = "prefer"      // 尝试SSL，如果失败则尝试非SSL（默认）
	SSLModeRequire    = "require"     // 要求SSL，不验证证书
	SSLModeVerifyCA   = "verify-ca"   // 要求SSL，验证服务器证书由可信CA签发
	SSLModeVerifyFull = "verify-full" // 要求SSL，验证服务器证书并检查主机名
)

// DefaultSSLConfigs 不同环境的默认SSL配置
var DefaultSSLConfigs = map[string]*SSLConfig{
	"development": {
		Mode:        SSLModeDisable,
		Environment: "development",
	},
	"test": {
		Mode:        SSLModeDisable,
		Environment: "test",
	},
	"staging": {
		Mode:        SSLModeRequire,
		Environment: "staging",
	},
	"production": {
		Mode:        SSLModeVerifyFull,
		CAFile:      "/etc/ssl/certs/postgresql-ca.crt",
		CertFile:    "/etc/ssl/certs/postgresql-client.crt",
		KeyFile:     "/etc/ssl/private/postgresql-client.key",
		Environment: "production",
	},
}

// NewSSLConfig 创建SSL配置
func NewSSLConfig(environment string) *SSLConfig {
	if config, exists := DefaultSSLConfigs[environment]; exists {
		return config
	}
	return DefaultSSLConfigs["development"]
}

// LoadFromEnv 从环境变量加载SSL配置
func (s *SSLConfig) LoadFromEnv() {
	if mode := os.Getenv("DB_SSL_MODE"); mode != "" {
		s.Mode = mode
	}
	if caFile := os.Getenv("DB_SSL_CA_FILE"); caFile != "" {
		s.CAFile = caFile
	}
	if certFile := os.Getenv("DB_SSL_CERT_FILE"); certFile != "" {
		s.CertFile = certFile
	}
	if keyFile := os.Getenv("DB_SSL_KEY_FILE"); keyFile != "" {
		s.KeyFile = keyFile
	}
	if crlFile := os.Getenv("DB_SSL_CRL_FILE"); crlFile != "" {
		s.CRLFile = crlFile
	}
	if serverName := os.Getenv("DB_SSL_SERVER_NAME"); serverName != "" {
		s.ServerName = serverName
	}
}

// Validate 验证SSL配置
func (s *SSLConfig) Validate() error {
	// 验证SSL模式
	validModes := []string{SSLModeDisable, SSLModeAllow, SSLModePrefer, 
		SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull}
	isValid := false
	for _, mode := range validModes {
		if s.Mode == mode {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid SSL mode: %s", s.Mode)
	}
	
	// 如果需要证书验证，检查证书文件
	if s.Mode == SSLModeVerifyCA || s.Mode == SSLModeVerifyFull {
		if s.CAFile == "" {
			return fmt.Errorf("CA file is required for SSL mode %s", s.Mode)
		}
		if _, err := os.Stat(s.CAFile); err != nil {
			return fmt.Errorf("CA file not found: %s", s.CAFile)
		}
	}
	
	// 如果提供了客户端证书，验证证书和私钥
	if s.CertFile != "" || s.KeyFile != "" {
		if s.CertFile == "" || s.KeyFile == "" {
			return fmt.Errorf("both cert file and key file must be provided")
		}
		if _, err := os.Stat(s.CertFile); err != nil {
			return fmt.Errorf("cert file not found: %s", s.CertFile)
		}
		if _, err := os.Stat(s.KeyFile); err != nil {
			return fmt.Errorf("key file not found: %s", s.KeyFile)
		}
	}
	
	return nil
}

// BuildDSNParams 构建DSN参数字符串
func (s *SSLConfig) BuildDSNParams() string {
	params := []string{fmt.Sprintf("sslmode=%s", s.Mode)}
	
	if s.CAFile != "" {
		params = append(params, fmt.Sprintf("sslrootcert=%s", s.CAFile))
	}
	if s.CertFile != "" {
		params = append(params, fmt.Sprintf("sslcert=%s", s.CertFile))
	}
	if s.KeyFile != "" {
		params = append(params, fmt.Sprintf("sslkey=%s", s.KeyFile))
	}
	if s.CRLFile != "" {
		params = append(params, fmt.Sprintf("sslcrl=%s", s.CRLFile))
	}
	
	return strings.Join(params, " ")
}

// CreateTLSConfig 创建TLS配置（用于某些驱动）
func (s *SSLConfig) CreateTLSConfig() (*tls.Config, error) {
	if s.Mode == SSLModeDisable {
		return nil, nil
	}
	
	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.Mode == SSLModeRequire,
	}
	
	// 设置服务器名称
	if s.ServerName != "" {
		tlsConfig.ServerName = s.ServerName
	}
	
	// 加载CA证书
	if s.CAFile != "" {
		caCert, err := ioutil.ReadFile(s.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}
	
	// 加载客户端证书
	if s.CertFile != "" && s.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	
	return tlsConfig, nil
}

// SSLManager SSL配置管理器
type SSLManager struct {
	configs map[string]*SSLConfig
}

// NewSSLManager 创建SSL管理器
func NewSSLManager() *SSLManager {
	return &SSLManager{
		configs: make(map[string]*SSLConfig),
	}
}

// LoadConfig 加载SSL配置
func (m *SSLManager) LoadConfig(name string, config *SSLConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid SSL config for %s: %w", name, err)
	}
	m.configs[name] = config
	return nil
}

// GetConfig 获取SSL配置
func (m *SSLManager) GetConfig(name string) (*SSLConfig, error) {
	if config, exists := m.configs[name]; exists {
		return config, nil
	}
	return nil, fmt.Errorf("SSL config %s not found", name)
}

// LoadDefaultConfigs 加载默认配置
func (m *SSLManager) LoadDefaultConfigs() error {
	for env, config := range DefaultSSLConfigs {
		if err := m.LoadConfig(env, config); err != nil {
			return err
		}
	}
	return nil
}

// GenerateCertificates 生成自签名证书（仅用于开发/测试）
func GenerateSelfSignedCertificates(outputDir string) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// TODO: 实现自签名证书生成逻辑
	// 这里应该使用crypto/x509和crypto/rsa包生成证书
	
	return fmt.Errorf("self-signed certificate generation not implemented")
}

// VerifyCertificateChain 验证证书链
func VerifyCertificateChain(certFile, caFile string) error {
	// 读取证书
	certPEM, err := ioutil.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}
	
	// 读取CA证书
	caPEM, err := ioutil.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	
	// 解析证书
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}
	
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	// 创建CA池
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEM) {
		return fmt.Errorf("failed to parse CA certificate")
	}
	
	// 验证证书
	opts := x509.VerifyOptions{
		Roots: caPool,
	}
	
	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}
	
	return nil
}

// GetRecommendedSSLMode 根据环境获取推荐的SSL模式
func GetRecommendedSSLMode(environment string) string {
	switch environment {
	case "production":
		return SSLModeVerifyFull
	case "staging":
		return SSLModeRequire
	case "test":
		return SSLModeDisable
	default:
		return SSLModeDisable
	}
}

// SSLHealthCheck SSL连接健康检查
type SSLHealthCheck struct {
	Enabled           bool
	CertificateValid  bool
	CertificateExpiry time.Time
	CAValid           bool
	TLSVersion        string
	CipherSuite       string
	LastCheckTime     time.Time
	Error             error
}

// CheckSSLHealth 检查SSL连接健康状态
func CheckSSLHealth(config *SSLConfig) (*SSLHealthCheck, error) {
	health := &SSLHealthCheck{
		Enabled:       config.Mode != SSLModeDisable,
		LastCheckTime: time.Now(),
	}
	
	if !health.Enabled {
		return health, nil
	}
	
	// 检查证书有效性
	if config.CertFile != "" {
		certPEM, err := ioutil.ReadFile(config.CertFile)
		if err != nil {
			health.Error = fmt.Errorf("failed to read certificate: %w", err)
			return health, health.Error
		}
		
		block, _ := pem.Decode(certPEM)
		if block != nil {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				health.CertificateValid = true
				health.CertificateExpiry = cert.NotAfter
				
				// 检查证书是否即将过期（30天内）
				if time.Until(cert.NotAfter) < 30*24*time.Hour {
					health.Error = fmt.Errorf("certificate expiring soon: %v", cert.NotAfter)
				}
			}
		}
	}
	
	// 检查CA证书
	if config.CAFile != "" {
		if _, err := os.Stat(config.CAFile); err == nil {
			health.CAValid = true
		}
	}
	
	return health, nil
}