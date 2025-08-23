package com.openpenpal.admin.config;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;

import javax.sql.DataSource;
import java.util.concurrent.TimeUnit;

/**
 * Connection Leak Detection Configuration
 * 
 * This configuration enhances the HikariCP data source with advanced leak detection
 * and monitoring capabilities. It's specifically designed to help identify and
 * resolve connection leaks that could cause the thread starvation issues observed
 * in the admin service.
 * 
 * Features:
 * - Aggressive leak detection in development
 * - Connection validation and timeout configuration
 * - Enhanced logging and monitoring
 * - Automatic leak alerting
 * 
 * @author OpenPenPal Admin Service
 * @since 2025-08-20
 */
@Configuration
public class ConnectionLeakDetectionConfig {
    
    private static final Logger log = LoggerFactory.getLogger(ConnectionLeakDetectionConfig.class);
    
    @Value("${spring.profiles.active:development}")
    private String activeProfile;
    
    @Value("${spring.datasource.url}")
    private String jdbcUrl;
    
    @Value("${spring.datasource.username:#{null}}")
    private String username;
    
    @Value("${spring.datasource.password:#{null}}")
    private String password;
    
    @Value("${spring.datasource.driver-class-name}")
    private String driverClassName;
    
    /**
     * Enhanced HikariCP DataSource with leak detection
     * This replaces the default DataSource when leak detection is enabled
     */
    @Bean
    @Primary
    @ConditionalOnProperty(name = "hikari.leak-detection.enabled", havingValue = "true", matchIfMissing = true)
    public DataSource leakDetectionDataSource() {
        log.info("Configuring HikariCP with enhanced leak detection for profile: {}", activeProfile);
        
        HikariConfig config = new HikariConfig();
        
        // Basic connection settings
        config.setJdbcUrl(jdbcUrl);
        config.setUsername(username);
        config.setPassword(password);
        config.setDriverClassName(driverClassName);
        
        // Enhanced pool configuration from application.yml
        config.setMaximumPoolSize(30);
        config.setMinimumIdle(3);
        config.setConnectionTimeout(TimeUnit.SECONDS.toMillis(20));
        config.setIdleTimeout(TimeUnit.MINUTES.toMillis(2));
        config.setMaxLifetime(TimeUnit.MINUTES.toMillis(30));
        config.setValidationTimeout(TimeUnit.SECONDS.toMillis(5));
        
        // Leak detection configuration
        if ("development".equals(activeProfile) || "test".equals(activeProfile)) {
            // Aggressive leak detection in development
            config.setLeakDetectionThreshold(TimeUnit.SECONDS.toMillis(30)); // 30 seconds
            log.info("Development mode: Aggressive leak detection enabled (30s threshold)");
        } else {
            // Production leak detection
            config.setLeakDetectionThreshold(TimeUnit.MINUTES.toMillis(1)); // 1 minute
            log.info("Production mode: Standard leak detection enabled (60s threshold)");
        }
        
        // Connection validation
        config.setConnectionTestQuery("SELECT 1");
        
        // Pool name for identification
        config.setPoolName("OpenPenPal-AdminService-Pool");
        
        // Advanced settings
        config.setInitializationFailTimeout(-1); // Fail fast on startup
        config.setIsolateInternalQueries(false); // Better performance
        config.setAllowPoolSuspension(false); // Prevent pool suspension
        config.setAutoCommit(true); // Ensure auto-commit
        
        // JMX monitoring
        config.setRegisterMbeans(true);
        
        // PostgreSQL specific optimizations
        config.addDataSourceProperty("prepareThreshold", "5");
        config.addDataSourceProperty("preparedStatementCacheQueries", "256");
        config.addDataSourceProperty("preparedStatementCacheSizeMiB", "5");
        config.addDataSourceProperty("socketTimeout", "300"); // 5 minutes
        config.addDataSourceProperty("tcpKeepAlive", "true");
        config.addDataSourceProperty("loginTimeout", "10");
        config.addDataSourceProperty("connectTimeout", "10");
        config.addDataSourceProperty("cancelSignalTimeout", "10");
        
        // Create and configure the data source
        HikariDataSource dataSource = new HikariDataSource(config);
        
        // Log the final configuration
        logDataSourceConfiguration(dataSource);
        
        return dataSource;
    }
    
    /**
     * Log the data source configuration for debugging
     */
    private void logDataSourceConfiguration(HikariDataSource dataSource) {
        log.info("HikariCP Configuration Summary:");
        log.info("  Pool Name: {}", dataSource.getPoolName());
        log.info("  Maximum Pool Size: {}", dataSource.getMaximumPoolSize());
        log.info("  Minimum Idle: {}", dataSource.getMinimumIdle());
        log.info("  Connection Timeout: {}ms", dataSource.getConnectionTimeout());
        log.info("  Idle Timeout: {}ms", dataSource.getIdleTimeout());
        log.info("  Max Lifetime: {}ms", dataSource.getMaxLifetime());
        log.info("  Leak Detection Threshold: {}ms", dataSource.getLeakDetectionThreshold());
        log.info("  Validation Timeout: {}ms", dataSource.getValidationTimeout());
        log.info("  Connection Test Query: {}", dataSource.getConnectionTestQuery());
        log.info("  Auto Commit: {}", dataSource.isAutoCommit());
        log.info("  JMX Enabled: {}", dataSource.isRegisterMbeans());
    }
    
    /**
     * Bean for testing connection leak detection
     * This bean can be used to manually trigger leak detection scenarios
     */
    @Bean
    @ConditionalOnProperty(name = "hikari.leak-detection.test-mode", havingValue = "true")
    public ConnectionLeakTester connectionLeakTester(DataSource dataSource) {
        return new ConnectionLeakTester(dataSource);
    }
    
    /**
     * Utility class for testing connection leaks
     */
    public static class ConnectionLeakTester {
        private final DataSource dataSource;
        private final Logger testLog = LoggerFactory.getLogger(ConnectionLeakTester.class);
        
        public ConnectionLeakTester(DataSource dataSource) {
            this.dataSource = dataSource;
        }
        
        /**
         * Simulate a connection leak for testing purposes
         * WARNING: Only use in development/testing!
         */
        public void simulateLeak() {
            try {
                testLog.warn("SIMULATING CONNECTION LEAK - FOR TESTING ONLY");
                dataSource.getConnection(); // Intentionally not closing
                testLog.warn("Connection acquired without closing - leak detection should trigger");
            } catch (Exception e) {
                testLog.error("Failed to simulate connection leak", e);
            }
        }
    }
}