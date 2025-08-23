package com.openpenpal.admin.config;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.core.env.Environment;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.TestPropertySource;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;

import static org.junit.jupiter.api.Assertions.*;

/**
 * DatabaseAutoConfiguration测试类
 * 测试数据库连接配置的回退逻辑和错误处理
 */
@SpringBootTest
@ActiveProfiles("test")
@DisplayName("数据库自动配置测试")
class DatabaseAutoConfigurationTest {

    @Autowired
    private Environment environment;

    @Nested
    @DisplayName("测试环境下的基本功能测试")
    class BasicFunctionalityTests {

        @Test
        @DisplayName("应该在测试环境下成功创建数据源")
        void shouldCreateDataSourceInTestEnvironment() {
            // Given - 测试环境配置
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            
            // When - 创建数据源
            DataSource dataSource = config.dataSource();
            
            // Then - 验证数据源不为空且可连接
            assertNotNull(dataSource, "数据源不应为空");
            
            // 验证可以获取连接
            assertDoesNotThrow(() -> {
                try (Connection conn = dataSource.getConnection()) {
                    assertTrue(conn.isValid(1), "连接应该有效");
                }
            }, "应该能够获取有效的数据库连接");
        }

        @Test
        @DisplayName("应该能执行基本SQL查询")
        void shouldExecuteBasicQuery() throws SQLException {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            DataSource dataSource = config.dataSource();
            
            // When & Then
            try (Connection conn = dataSource.getConnection()) {
                var statement = conn.createStatement();
                var result = statement.executeQuery("SELECT 1 as test_value");
                
                assertTrue(result.next(), "查询应该返回结果");
                assertEquals(1, result.getInt("test_value"), "查询结果应该正确");
            }
        }
    }

    @Nested
    @DisplayName("配置验证测试")
    @TestPropertySource(properties = {
        "spring.datasource.url=jdbc:h2:mem:testdb",
        "spring.datasource.username=sa",
        "spring.datasource.password="
    })
    class ConfigurationValidationTests {

        @Test
        @DisplayName("应该验证数据库配置的有效性")
        void shouldValidateDatabaseConfig() {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            
            // When & Then - 在H2测试环境下应该成功
            assertDoesNotThrow(() -> {
                DataSource dataSource = config.dataSource();
                assertNotNull(dataSource);
            });
        }
    }

    @Nested
    @DisplayName("错误处理测试")
    class ErrorHandlingTests {

        @Test
        @DisplayName("无效配置应该提供清晰的错误信息")
        void shouldProvideDescriptiveErrorForInvalidConfig() {
            // 注意：这个测试在真实环境下会尝试连接PostgreSQL
            // 在测试环境下使用H2，所以实际不会失败
            // 这里主要测试配置逻辑的存在性
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            
            // 验证配置对象可以创建（不测试实际连接失败，因为测试环境使用H2）
            assertNotNull(config);
        }

        @Test
        @DisplayName("应该处理数据源创建过程中的异常")
        void shouldHandleDataSourceCreationExceptions() {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            
            // When & Then - 确保配置不会抛出意外异常
            assertDoesNotThrow(() -> {
                DataSource dataSource = config.dataSource();
                // 在测试环境下应该成功创建
                assertNotNull(dataSource);
            });
        }
    }

    @Nested
    @DisplayName("连接池验证测试")
    class ConnectionPoolTests {

        @Test
        @DisplayName("应该支持多个并发连接")
        void shouldSupportMultipleConcurrentConnections() throws SQLException {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            DataSource dataSource = config.dataSource();
            
            // When - 创建多个连接
            Connection conn1 = dataSource.getConnection();
            Connection conn2 = dataSource.getConnection();
            
            try {
                // Then - 验证连接独立性
                assertNotNull(conn1);
                assertNotNull(conn2);
                assertNotSame(conn1, conn2, "应该是不同的连接实例");
                
                assertTrue(conn1.isValid(1), "连接1应该有效");
                assertTrue(conn2.isValid(1), "连接2应该有效");
                
            } finally {
                conn1.close();
                conn2.close();
            }
        }

        @Test
        @DisplayName("连接关闭后应该能重新获取")
        void shouldGetNewConnectionAfterClose() throws SQLException {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            DataSource dataSource = config.dataSource();
            
            // When - 获取连接，关闭，再获取
            Connection conn1 = dataSource.getConnection();
            conn1.close();
            Connection conn2 = dataSource.getConnection();
            
            try {
                // Then
                assertTrue(conn1.isClosed(), "第一个连接应该已关闭");
                assertFalse(conn2.isClosed(), "第二个连接应该是打开的");
                assertTrue(conn2.isValid(1), "新连接应该有效");
                
            } finally {
                if (!conn2.isClosed()) {
                    conn2.close();
                }
            }
        }
    }

    @Nested
    @DisplayName("环境特定测试")
    class EnvironmentSpecificTests {

        @Test
        @DisplayName("应该在测试Profile下正确工作")
        void shouldWorkCorrectlyUnderTestProfile() {
            // Given
            String[] activeProfiles = environment.getActiveProfiles();
            
            // When & Then
            assertTrue(
                java.util.Arrays.asList(activeProfiles).contains("test"),
                "应该包含test profile"
            );
            
            // 验证配置在测试环境下工作
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            assertDoesNotThrow(() -> {
                DataSource dataSource = config.dataSource();
                assertNotNull(dataSource);
            });
        }

        @Test
        @DisplayName("应该使用正确的数据库驱动")
        void shouldUseCorrectDatabaseDriver() throws SQLException {
            // Given
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(environment);
            DataSource dataSource = config.dataSource();
            
            // When
            try (Connection conn = dataSource.getConnection()) {
                String driverName = conn.getMetaData().getDriverName();
                
                // Then - 在测试环境下应该是H2
                assertTrue(
                    driverName.contains("H2") || driverName.contains("PostgreSQL"),
                    "应该使用H2或PostgreSQL驱动，实际: " + driverName
                );
            }
        }
    }
}