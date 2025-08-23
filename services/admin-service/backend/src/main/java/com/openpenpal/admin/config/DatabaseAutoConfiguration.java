package com.openpenpal.admin.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.boot.jdbc.DataSourceBuilder;
import org.springframework.core.env.Environment;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.DriverManager;
import java.util.Arrays;
import java.util.List;

/**
 * PostgreSQL专用数据库配置
 * 确保只使用PostgreSQL，不降级到H2
 */
@Configuration
public class DatabaseAutoConfiguration {

    private final Environment env;

    public DatabaseAutoConfiguration(Environment env) {
        this.env = env;
    }

    @Bean
    public DataSource dataSource() {
        // PostgreSQL优先级配置列表
        List<DatabaseConfig> postgresConfigs = Arrays.asList(
            // 1. 环境变量配置（最高优先级）
            new DatabaseConfig(
                env.getProperty("DATABASE_URL", ""),
                env.getProperty("DB_USERNAME", ""),
                env.getProperty("DB_PASSWORD", "")
            ),
            // 2. 系统用户配置（与主服务一致）
            new DatabaseConfig(
                "jdbc:postgresql://localhost:5432/openpenpal",
                System.getProperty("user.name", "rocalight"),
                "password"
            ),
            // 3. openpenpal专用用户
            new DatabaseConfig(
                "jdbc:postgresql://localhost:5432/openpenpal",
                "openpenpal",
                "openpenpal"
            ),
            // 4. 标准postgres用户
            new DatabaseConfig(
                "jdbc:postgresql://localhost:5432/openpenpal",
                "postgres",
                "postgres"
            ),
            new DatabaseConfig(
                "jdbc:postgresql://localhost:5432/openpenpal",
                "postgres",
                ""
            )
        );

        // 尝试所有PostgreSQL配置
        for (DatabaseConfig config : postgresConfigs) {
            if (config.isValid() && testConnection(config)) {
                System.out.println("✅ [PostgreSQL] 成功连接数据库: " + config.getUrl());
                System.out.println("   用户: " + config.getUsername());
                System.out.println("   数据源: PostgreSQL (生产模式)");
                
                return DataSourceBuilder.create()
                    .url(config.getUrl())
                    .username(config.getUsername())
                    .password(config.getPassword())
                    .driverClassName("org.postgresql.Driver")
                    .build();
            }
        }

        // 如果所有配置都失败，抛出异常而不是降级
        throw new RuntimeException(
            "❌ 无法连接到PostgreSQL数据库！\n" +
            "   请确保：\n" +
            "   1. PostgreSQL服务正在运行\n" +
            "   2. 数据库'openpenpal'存在\n" +
            "   3. 用户权限正确\n" +
            "   可以运行: ./startup/database-manager.sh ensure"
        );
    }

    private boolean testConnection(DatabaseConfig config) {
        if (!config.isValid()) {
            return false;
        }
        
        try (Connection conn = DriverManager.getConnection(
                config.getUrl(), 
                config.getUsername(), 
                config.getPassword())) {
            
            boolean valid = conn.isValid(3);
            if (valid) {
                // 测试基本查询
                conn.createStatement().executeQuery("SELECT 1");
            }
            return valid;
            
        } catch (Exception e) {
            System.out.println("⚠️  测试连接失败 [" + config.getUsername() + "@" + config.getUrl() + "]: " + e.getMessage());
            return false;
        }
    }

    private static class DatabaseConfig {
        private final String url;
        private final String username;
        private final String password;

        public DatabaseConfig(String url, String username, String password) {
            this.url = url != null ? url : "";
            this.username = username != null ? username : "";
            this.password = password != null ? password : "";
        }

        public boolean isValid() {
            return !url.isEmpty() && 
                   !username.isEmpty() && 
                   url.contains("postgresql");
        }

        public String getUrl() { return url; }
        public String getUsername() { return username; }
        public String getPassword() { return password; }
    }
}