package com.openpenpal.admin;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.sql.DataSource;
import java.sql.Connection;
import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;

@SpringBootApplication
@RestController
public class AdminServiceApplication {

    @Autowired
    private DataSource dataSource;

    public static void main(String[] args) {
        System.out.println("🚀 [SOTA] OpenPenPal Admin Service 启动中...");
        SpringApplication.run(AdminServiceApplication.class, args);
        System.out.println("✅ [SOTA] Admin Service 启动完成！");
    }

    @GetMapping("/")
    public Map<String, Object> home() {
        Map<String, Object> response = new HashMap<>();
        response.put("service", "OpenPenPal Admin Service");
        response.put("version", "1.0.0-SOTA");
        response.put("status", "running");
        response.put("timestamp", LocalDateTime.now().toString());
        response.put("database", getDatabaseInfo());
        return response;
    }

    @GetMapping("/health")
    public Map<String, Object> health() {
        Map<String, Object> response = new HashMap<>();
        
        // SOTA: 智能健康检查
        boolean dbHealthy = checkDatabaseHealth();
        
        response.put("status", dbHealthy ? "UP" : "DEGRADED");
        response.put("service", "admin-service");
        response.put("timestamp", LocalDateTime.now().toString());
        response.put("database", getDatabaseInfo());
        response.put("database_healthy", dbHealthy);
        
        return response;
    }

    @GetMapping("/status")
    public Map<String, Object> detailedStatus() {
        Map<String, Object> response = new HashMap<>();
        response.put("service", "OpenPenPal Admin Service");
        response.put("version", "1.0.0-SOTA");
        response.put("uptime", getUptime());
        response.put("database", getDatabaseInfo());
        response.put("features", new String[]{"智能数据库自适应", "优雅降级", "实时健康监控"});
        return response;
    }

    private boolean checkDatabaseHealth() {
        try (Connection conn = dataSource.getConnection()) {
            return conn.isValid(2);
        } catch (Exception e) {
            return false;
        }
    }

    private Map<String, Object> getDatabaseInfo() {
        Map<String, Object> dbInfo = new HashMap<>();
        try (Connection conn = dataSource.getConnection()) {
            String url = conn.getMetaData().getURL();
            String dbProduct = conn.getMetaData().getDatabaseProductName();
            String dbVersion = conn.getMetaData().getDatabaseProductVersion();
            
            dbInfo.put("url", url);
            dbInfo.put("product", dbProduct);
            dbInfo.put("version", dbVersion);
            dbInfo.put("healthy", true);
            
            if (url.contains("h2")) {
                dbInfo.put("mode", "降级模式 (H2内存数据库)");
            } else if (url.contains("postgresql")) {
                dbInfo.put("mode", "生产模式 (PostgreSQL)");
            }
        } catch (Exception e) {
            dbInfo.put("error", e.getMessage());
            dbInfo.put("healthy", false);
        }
        return dbInfo;
    }

    private String getUptime() {
        return "已运行";
    }
}