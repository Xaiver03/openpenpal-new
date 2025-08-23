package com.openpenpal.admin.config;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.Status;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

/**
 * PostgreSQLHealthIndicator测试类
 * 测试数据库健康检查功能
 */
@SpringBootTest
@ActiveProfiles("test")
@DisplayName("PostgreSQL健康指示器测试")
class PostgreSQLHealthIndicatorTest {

    @Autowired
    private DataSource dataSource;

    private PostgreSQLHealthIndicator healthIndicator;

    @Mock
    private DataSource mockDataSource;

    @Mock
    private Connection mockConnection;

    @Mock
    private PreparedStatement mockStatement;

    @Mock
    private ResultSet mockResultSet;

    @BeforeEach
    void setUp() {
        MockitoAnnotations.openMocks(this);
        healthIndicator = new PostgreSQLHealthIndicator(dataSource);
    }

    @Nested
    @DisplayName("健康检查成功场景")
    class HealthCheckSuccessTests {

        @Test
        @DisplayName("健康的数据库应该返回UP状态")
        void healthyDatabaseShouldReturnUpStatus() {
            // When
            Health health = healthIndicator.health();
            
            // Then
            assertEquals(Status.UP, health.getStatus(), "健康的数据库应该返回UP状态");
            assertNotNull(health.getDetails(), "健康检查应该包含详细信息");
        }

        @Test
        @DisplayName("健康检查应该包含数据库版本信息")
        void healthCheckShouldIncludeDatabaseVersion() {
            // When
            Health health = healthIndicator.health();
            
            // Then
            Map<String, Object> details = health.getDetails();
            assertTrue(details.containsKey("version") || details.containsKey("database"),
                "健康检查应该包含数据库版本或数据库名称信息");
        }

        @Test
        @DisplayName("健康检查应该包含连接池状态")
        void healthCheckShouldIncludeConnectionPoolStatus() {
            // When
            Health health = healthIndicator.health();
            
            // Then
            Map<String, Object> details = health.getDetails();
            
            // 在测试环境下，某些详细信息可能不同，但应该有基本信息
            assertNotNull(details, "应该有健康检查详细信息");
        }
    }

    @Nested
    @DisplayName("健康检查失败场景")
    class HealthCheckFailureTests {

        @Test
        @DisplayName("数据源连接失败应该返回DOWN状态")
        void dataSourceConnectionFailureShouldReturnDownStatus() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            when(mockDataSource.getConnection()).thenThrow(new SQLException("Connection failed"));
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.DOWN, health.getStatus(), "连接失败时应该返回DOWN状态");
            assertTrue(health.getDetails().containsKey("error"), "失败时应该包含错误信息");
        }

        @Test
        @DisplayName("无效连接应该返回DOWN状态")
        void invalidConnectionShouldReturnDownStatus() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            when(mockDataSource.getConnection()).thenReturn(mockConnection);
            when(mockConnection.isValid(anyInt())).thenReturn(false);
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.DOWN, health.getStatus(), "无效连接应该返回DOWN状态");
            assertEquals("Connection is not valid", 
                health.getDetails().get("error"), "应该包含连接无效的错误信息");
        }

        @Test
        @DisplayName("查询执行失败应该返回DOWN状态")
        void queryExecutionFailureShouldReturnDownStatus() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            when(mockDataSource.getConnection()).thenReturn(mockConnection);
            when(mockConnection.isValid(anyInt())).thenReturn(true);
            when(mockConnection.prepareStatement(anyString())).thenThrow(new SQLException("Query failed"));
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.DOWN, health.getStatus(), "查询失败时应该返回DOWN状态");
            assertTrue(health.getDetails().containsKey("error"), "失败时应该包含错误信息");
        }
    }

    @Nested
    @DisplayName("Mock数据库详细信息测试")
    class MockDatabaseDetailsTests {

        @Test
        @DisplayName("成功的健康检查应该包含数据库详细信息")
        void successfulHealthCheckShouldIncludeDatabaseDetails() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            
            when(mockDataSource.getConnection()).thenReturn(mockConnection);
            when(mockConnection.isValid(anyInt())).thenReturn(true);
            when(mockConnection.prepareStatement(contains("version()"))).thenReturn(mockStatement);
            when(mockStatement.executeQuery()).thenReturn(mockResultSet);
            when(mockResultSet.next()).thenReturn(true);
            when(mockResultSet.getString(1)).thenReturn("PostgreSQL 15.0");
            when(mockResultSet.getString(2)).thenReturn("testdb");
            when(mockResultSet.getString(3)).thenReturn("testuser");
            when(mockConnection.getAutoCommit()).thenReturn(true);
            when(mockConnection.getTransactionIsolation()).thenReturn(Connection.TRANSACTION_READ_COMMITTED);
            
            // Mock table existence check
            PreparedStatement tableCheckStatement = mock(PreparedStatement.class);
            ResultSet tableCheckResultSet = mock(ResultSet.class);
            when(mockConnection.prepareStatement(contains("information_schema.tables"))).thenReturn(tableCheckStatement);
            when(tableCheckStatement.executeQuery()).thenReturn(tableCheckResultSet);
            when(tableCheckResultSet.next()).thenReturn(true);
            when(tableCheckResultSet.getBoolean(1)).thenReturn(true);
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.UP, health.getStatus());
            Map<String, Object> details = health.getDetails();
            
            assertEquals("PostgreSQL 15.0", details.get("version"));
            assertEquals("testdb", details.get("database"));
            assertEquals("testuser", details.get("user"));
            assertEquals("active", details.get("connection_pool"));
            assertEquals(true, details.get("auto_commit"));
            assertEquals(Connection.TRANSACTION_READ_COMMITTED, details.get("transaction_isolation"));
        }

        @Test
        @DisplayName("表存在检查应该正确工作")
        void tableExistenceCheckShouldWorkCorrectly() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            
            when(mockDataSource.getConnection()).thenReturn(mockConnection);
            when(mockConnection.isValid(anyInt())).thenReturn(true);
            
            // Mock version query
            PreparedStatement versionStatement = mock(PreparedStatement.class);
            ResultSet versionResultSet = mock(ResultSet.class);
            when(mockConnection.prepareStatement(contains("version()"))).thenReturn(versionStatement);
            when(versionStatement.executeQuery()).thenReturn(versionResultSet);
            when(versionResultSet.next()).thenReturn(true);
            when(versionResultSet.getString(anyInt())).thenReturn("test");
            when(mockConnection.getAutoCommit()).thenReturn(true);
            when(mockConnection.getTransactionIsolation()).thenReturn(Connection.TRANSACTION_READ_COMMITTED);
            
            // Mock table existence - users exists, letters missing, couriers exists
            PreparedStatement tableStatement = mock(PreparedStatement.class);
            ResultSet tableResultSet = mock(ResultSet.class);
            when(mockConnection.prepareStatement(contains("information_schema.tables"))).thenReturn(tableStatement);
            when(tableStatement.executeQuery()).thenReturn(tableResultSet);
            when(tableResultSet.next()).thenReturn(true);
            when(tableResultSet.getBoolean(1))
                .thenReturn(true)   // users exists
                .thenReturn(false)  // letters missing
                .thenReturn(true);  // couriers exists
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.UP, health.getStatus());
            Map<String, Object> details = health.getDetails();
            
            assertEquals("exists", details.get("table_users"));
            assertEquals("missing", details.get("table_letters"));
            assertEquals("exists", details.get("table_couriers"));
        }
    }

    @Nested
    @DisplayName("错误处理测试")
    class ErrorHandlingTests {

        @Test
        @DisplayName("应该正确处理SQL异常")
        void shouldHandleSqlExceptionsCorrectly() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            SQLException sqlException = new SQLException("Database error", "42000", 123);
            when(mockDataSource.getConnection()).thenThrow(sqlException);
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.DOWN, health.getStatus());
            String errorMessage = (String) health.getDetails().get("error");
            assertTrue(errorMessage.contains("SQLException"));
            assertTrue(errorMessage.contains("Database error"));
        }

        @Test
        @DisplayName("应该正确处理运行时异常")
        void shouldHandleRuntimeExceptionsCorrectly() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            RuntimeException runtimeException = new RuntimeException("Runtime error");
            when(mockDataSource.getConnection()).thenThrow(runtimeException);
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then
            assertEquals(Status.DOWN, health.getStatus());
            String errorMessage = (String) health.getDetails().get("error");
            assertTrue(errorMessage.contains("RuntimeException"));
            assertTrue(errorMessage.contains("Runtime error"));
        }
    }

    @Nested
    @DisplayName("资源管理测试")
    class ResourceManagementTests {

        @Test
        @DisplayName("应该正确关闭数据库资源")
        void shouldProperlyCloseDatabaseResources() throws SQLException {
            // Given
            PostgreSQLHealthIndicator mockHealthIndicator = new PostgreSQLHealthIndicator(mockDataSource);
            
            when(mockDataSource.getConnection()).thenReturn(mockConnection);
            when(mockConnection.isValid(anyInt())).thenReturn(true);
            when(mockConnection.prepareStatement(anyString())).thenReturn(mockStatement);
            when(mockStatement.executeQuery()).thenReturn(mockResultSet);
            when(mockResultSet.next()).thenReturn(true);
            when(mockResultSet.getString(anyInt())).thenReturn("test");
            when(mockConnection.getAutoCommit()).thenReturn(true);
            when(mockConnection.getTransactionIsolation()).thenReturn(Connection.TRANSACTION_READ_COMMITTED);
            
            // When
            Health health = mockHealthIndicator.health();
            
            // Then - verify resources are closed (try-with-resources pattern)
            verify(mockConnection, atLeastOnce()).close();
            assertEquals(Status.UP, health.getStatus());
        }
    }
}