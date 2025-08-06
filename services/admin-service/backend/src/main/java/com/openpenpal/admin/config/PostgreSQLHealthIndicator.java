package com.openpenpal.admin.config;

import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.stereotype.Component;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.util.HashMap;
import java.util.Map;

/**
 * PostgreSQL健康检查指示器
 * 确保数据库连接稳定可靠
 */
@Component
public class PostgreSQLHealthIndicator implements HealthIndicator {

    private final DataSource dataSource;

    public PostgreSQLHealthIndicator(DataSource dataSource) {
        this.dataSource = dataSource;
    }

    @Override
    public Health health() {
        try (Connection connection = dataSource.getConnection()) {
            Map<String, Object> details = new HashMap<>();
            
            // 1. 检查连接是否有效
            if (!connection.isValid(2)) {
                return Health.down()
                    .withDetail("error", "Connection is not valid")
                    .build();
            }
            
            // 2. 执行测试查询
            try (PreparedStatement ps = connection.prepareStatement("SELECT version(), current_database(), current_user")) {
                try (ResultSet rs = ps.executeQuery()) {
                    if (rs.next()) {
                        details.put("version", rs.getString(1));
                        details.put("database", rs.getString(2));
                        details.put("user", rs.getString(3));
                    }
                }
            }
            
            // 3. 检查关键表
            String[] requiredTables = {"users", "letters", "couriers"};
            for (String table : requiredTables) {
                boolean tableExists = checkTableExists(connection, table);
                details.put("table_" + table, tableExists ? "exists" : "missing");
            }
            
            // 4. 检查连接池状态
            details.put("connection_pool", "active");
            details.put("auto_commit", connection.getAutoCommit());
            details.put("transaction_isolation", connection.getTransactionIsolation());
            
            return Health.up()
                .withDetails(details)
                .build();
                
        } catch (Exception e) {
            return Health.down()
                .withDetail("error", e.getClass().getName() + ": " + e.getMessage())
                .build();
        }
    }
    
    private boolean checkTableExists(Connection connection, String tableName) {
        try (PreparedStatement ps = connection.prepareStatement(
                "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)")) {
            ps.setString(1, tableName);
            try (ResultSet rs = ps.executeQuery()) {
                return rs.next() && rs.getBoolean(1);
            }
        } catch (Exception e) {
            return false;
        }
    }
}