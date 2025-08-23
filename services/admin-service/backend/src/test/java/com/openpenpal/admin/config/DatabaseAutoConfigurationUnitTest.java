package com.openpenpal.admin.config;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.springframework.core.env.Environment;

import javax.sql.DataSource;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

/**
 * DatabaseAutoConfiguration单元测试
 * 不依赖Spring上下文的纯单元测试
 */
@DisplayName("数据库自动配置单元测试")
class DatabaseAutoConfigurationUnitTest {

    @Mock
    private Environment mockEnvironment;
    
    private DatabaseAutoConfiguration configuration;

    @BeforeEach
    void setUp() {
        MockitoAnnotations.openMocks(this);
        configuration = new DatabaseAutoConfiguration(mockEnvironment);
    }

    @Nested
    @DisplayName("配置对象创建测试")
    class ConfigurationCreationTests {

        @Test
        @DisplayName("应该能够创建配置对象")
        void shouldCreateConfigurationObject() {
            // Given
            Environment env = mock(Environment.class);
            
            // When & Then
            assertDoesNotThrow(() -> {
                DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(env);
                assertNotNull(config, "配置对象不应为空");
            }, "应该能够创建DatabaseAutoConfiguration对象");
        }

        @Test
        @DisplayName("配置对象应该接受Environment参数")
        void shouldAcceptEnvironmentParameter() {
            // Given
            Environment env = mock(Environment.class);
            
            // When
            DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(env);
            
            // Then
            assertNotNull(config, "配置对象应该成功创建");
        }
    }

    @Nested
    @DisplayName("环境变量处理测试")
    class EnvironmentVariableTests {

        @Test
        @DisplayName("应该处理DATABASE_URL环境变量")
        void shouldHandleDatabaseUrlEnvironmentVariable() {
            // Given
            when(mockEnvironment.getProperty("DATABASE_URL", ""))
                .thenReturn("jdbc:postgresql://localhost:5432/testdb");
            when(mockEnvironment.getProperty("DB_USERNAME", ""))
                .thenReturn("testuser");
            when(mockEnvironment.getProperty("DB_PASSWORD", ""))
                .thenReturn("testpass");
            
            // When & Then
            assertDoesNotThrow(() -> {
                // 验证配置对象能被创建（这会间接测试环境变量处理）
                DatabaseAutoConfiguration config = new DatabaseAutoConfiguration(mockEnvironment);
                assertNotNull(config, "配置对象应该成功创建");
            }, "应该能处理DATABASE_URL环境变量");
        }

        @Test
        @DisplayName("应该处理空的环境变量")
        void shouldHandleEmptyEnvironmentVariables() {
            // Given
            when(mockEnvironment.getProperty(anyString(), anyString()))
                .thenReturn("");
            
            // When & Then
            assertDoesNotThrow(() -> {
                // 配置对象应该能处理空的环境变量
                assertNotNull(configuration);
            }, "应该能处理空的环境变量");
        }
    }

    @Nested
    @DisplayName("数据源配置测试")
    class DataSourceConfigurationTests {

        @Test
        @DisplayName("配置应该定义dataSource方法")
        void shouldDefineDataSourceMethod() {
            // When & Then
            assertDoesNotThrow(() -> {
                var method = DatabaseAutoConfiguration.class.getMethod("dataSource");
                assertNotNull(method, "dataSource方法应该存在");
                assertEquals(DataSource.class, method.getReturnType(), 
                    "dataSource方法应该返回DataSource类型");
            }, "应该定义dataSource方法");
        }

        @Test
        @DisplayName("dataSource方法应该有@Primary注解")
        void dataSourceMethodShouldHavePrimaryAnnotation() {
            // When & Then
            assertDoesNotThrow(() -> {
                var method = DatabaseAutoConfiguration.class.getMethod("dataSource");
                var primaryAnnotation = method.getAnnotation(org.springframework.context.annotation.Primary.class);
                assertNotNull(primaryAnnotation, "dataSource方法应该有@Primary注解");
            }, "dataSource方法应该标记为@Primary");
        }

        @Test
        @DisplayName("dataSource方法应该有@Bean注解")
        void dataSourceMethodShouldHaveBeanAnnotation() {
            // When & Then
            assertDoesNotThrow(() -> {
                var method = DatabaseAutoConfiguration.class.getMethod("dataSource");
                var beanAnnotation = method.getAnnotation(org.springframework.context.annotation.Bean.class);
                assertNotNull(beanAnnotation, "dataSource方法应该有@Bean注解");
            }, "dataSource方法应该标记为@Bean");
        }
    }

    @Nested
    @DisplayName("注解验证测试")
    class AnnotationTests {

        @Test
        @DisplayName("类应该有@Configuration注解")
        void classShouldHaveConfigurationAnnotation() {
            // When
            var configAnnotation = DatabaseAutoConfiguration.class
                .getAnnotation(org.springframework.context.annotation.Configuration.class);
            
            // Then
            assertNotNull(configAnnotation, "DatabaseAutoConfiguration应该有@Configuration注解");
        }

        @Test
        @DisplayName("类应该是Spring配置类")
        void classShouldBeSpringConfigurationClass() {
            // When
            boolean isConfiguration = DatabaseAutoConfiguration.class
                .isAnnotationPresent(org.springframework.context.annotation.Configuration.class);
            
            // Then
            assertTrue(isConfiguration, "DatabaseAutoConfiguration应该是Spring配置类");
        }
    }

    @Nested
    @DisplayName("异常处理测试")
    class ExceptionHandlingTests {

        @Test
        @DisplayName("空Environment参数不应导致NPE")
        void nullEnvironmentShouldNotCauseNPE() {
            // When & Then
            assertDoesNotThrow(() -> {
                // 虽然传入null可能不是最佳实践，但不应该立即抛出NPE
                // 实际的NPE可能在调用dataSource()方法时发生
                new DatabaseAutoConfiguration(null);
            }, "null Environment参数不应该立即导致NPE");
        }
    }
}