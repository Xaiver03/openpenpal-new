package com.openpenpal.admin;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Admin Service Application基础测试
 * 验证应用程序类的基本结构
 */
@DisplayName("Admin Service Application 单元测试")
class AdminServiceApplicationTest {

    @Test
    @DisplayName("AdminServiceApplication类应该存在")
    void shouldExistAdminServiceApplicationClass() {
        // 验证AdminServiceApplication类存在
        assertDoesNotThrow(() -> {
            Class<?> appClass = AdminServiceApplication.class;
            assertNotNull(appClass, "AdminServiceApplication类应该存在");
        }, "AdminServiceApplication类应该可以被加载");
    }

    @Test
    @DisplayName("AdminServiceApplication类应该存在主方法")
    void shouldHaveMainMethod() {
        // 验证AdminServiceApplication类存在且有main方法
        assertDoesNotThrow(() -> {
            Class<?> appClass = AdminServiceApplication.class;
            assertNotNull(appClass, "AdminServiceApplication类应该存在");
            
            // 验证main方法存在
            var mainMethod = appClass.getMethod("main", String[].class);
            assertNotNull(mainMethod, "main方法应该存在");
            assertEquals("main", mainMethod.getName(), "方法名应该是main");
            assertEquals(void.class, mainMethod.getReturnType(), "main方法应该返回void");
        }, "AdminServiceApplication应该有有效的main方法");
    }

    @Test
    @DisplayName("AdminServiceApplication类应该有Spring Boot注解")
    void shouldHaveSpringBootAnnotations() {
        // 验证SpringBootApplication注解存在
        Class<?> appClass = AdminServiceApplication.class;
        boolean hasSpringBootApp = appClass.isAnnotationPresent(
            org.springframework.boot.autoconfigure.SpringBootApplication.class);
        
        assertTrue(hasSpringBootApp, "AdminServiceApplication应该有@SpringBootApplication注解");
    }
}