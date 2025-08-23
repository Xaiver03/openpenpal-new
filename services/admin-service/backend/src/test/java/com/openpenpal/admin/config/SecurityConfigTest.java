package com.openpenpal.admin.config;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureWebMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.security.test.context.support.WithAnonymousUser;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.web.cors.CorsConfigurationSource;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;
import static org.springframework.test.web.servlet.result.MockMvcResultHandlers.*;
import static org.hamcrest.Matchers.*;

/**
 * SecurityConfig测试类
 * 测试安全配置的权限控制和CORS设置
 */
@SpringBootTest
@AutoConfigureWebMvc
@ActiveProfiles("test")
@DisplayName("安全配置测试")
class SecurityConfigTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private CorsConfigurationSource corsConfigurationSource;

    @Nested
    @DisplayName("公开端点访问测试")
    class PublicEndpointTests {

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户应该能访问根路径")
        void anonymousUserShouldAccessRoot() throws Exception {
            mockMvc.perform(get("/"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户应该能访问健康检查端点")
        void anonymousUserShouldAccessHealthEndpoint() throws Exception {
            mockMvc.perform(get("/health"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户应该能访问状态端点")
        void anonymousUserShouldAccessStatusEndpoint() throws Exception {
            mockMvc.perform(get("/status"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户应该能访问Actuator端点")
        void anonymousUserShouldAccessActuatorEndpoints() throws Exception {
            mockMvc.perform(get("/actuator/health"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户应该能访问认证端点")
        void anonymousUserShouldAccessAuthEndpoints() throws Exception {
            // 测试认证端点是否允许匿名访问（虽然可能返回404，但不应该是401/403）
            mockMvc.perform(post("/api/admin/auth/login"))
                .andDo(print())
                .andExpect(status().is(not(401)))
                .andExpect(status().is(not(403)));
        }
    }

    @Nested
    @DisplayName("受保护端点访问测试")
    class ProtectedEndpointTests {

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户不应该能访问受保护的API")
        void anonymousUserShouldNotAccessProtectedApi() throws Exception {
            mockMvc.perform(get("/api/admin/users"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }

        @Test
        @WithMockUser(username = "admin", roles = {"ADMIN"})
        @DisplayName("认证用户应该能访问受保护的API")
        void authenticatedUserShouldAccessProtectedApi() throws Exception {
            // 注意：这可能返回404因为端点不存在，但不应该是401/403
            mockMvc.perform(get("/api/admin/users"))
                .andDo(print())
                .andExpect(status().is(not(401)))
                .andExpect(status().is(not(403)));
        }

        @Test
        @WithAnonymousUser
        @DisplayName("匿名用户访问需要认证的端点应该返回401")
        void anonymousUserShouldGet401ForAuthenticatedEndpoints() throws Exception {
            mockMvc.perform(get("/api/admin/dashboard"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }
    }

    @Nested
    @DisplayName("CSRF配置测试")
    class CsrfConfigurationTests {

        @Test
        @WithAnonymousUser
        @DisplayName("CSRF应该被禁用 - POST请求不应要求CSRF令牌")
        void csrfShouldBeDisabled() throws Exception {
            // POST请求到公开端点不应该因为缺少CSRF令牌而失败
            mockMvc.perform(post("/api/admin/auth/login")
                    .contentType("application/json")
                    .content("{}"))
                .andDo(print())
                .andExpect(status().is(not(403))); // 不应该因为CSRF而被拒绝
        }

        @Test
        @WithMockUser
        @DisplayName("已认证用户的POST请求不应要求CSRF令牌")
        void authenticatedPostShouldNotRequireCsrf() throws Exception {
            mockMvc.perform(post("/api/admin/test")
                    .contentType("application/json")
                    .content("{}"))
                .andDo(print())
                .andExpect(status().is(not(403))); // 不应该因为CSRF而被拒绝
        }
    }

    @Nested
    @DisplayName("CORS配置测试")
    class CorsConfigurationTests {

        @Test
        @WithAnonymousUser
        @DisplayName("OPTIONS请求应该被正确处理（CORS预检）")
        void optionsRequestShouldBeHandledForCors() throws Exception {
            mockMvc.perform(options("/api/admin/test")
                    .header("Origin", "http://localhost:3000")
                    .header("Access-Control-Request-Method", "POST"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("跨域请求应该包含Origin头")
        void crossOriginRequestShouldIncludeOriginHeader() throws Exception {
            mockMvc.perform(get("/health")
                    .header("Origin", "http://localhost:3000"))
                .andDo(print())
                .andExpect(status().isOk());
        }
    }

    @Nested
    @DisplayName("HTTP方法安全测试")
    class HttpMethodSecurityTests {

        @Test
        @WithAnonymousUser
        @DisplayName("GET请求到公开端点应该成功")
        void getRequestToPublicEndpointShouldSucceed() throws Exception {
            mockMvc.perform(get("/health"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("POST请求到受保护端点应该要求认证")
        void postRequestToProtectedEndpointShouldRequireAuth() throws Exception {
            mockMvc.perform(post("/api/admin/protected"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("PUT请求到受保护端点应该要求认证")
        void putRequestToProtectedEndpointShouldRequireAuth() throws Exception {
            mockMvc.perform(put("/api/admin/protected"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("DELETE请求到受保护端点应该要求认证")
        void deleteRequestToProtectedEndpointShouldRequireAuth() throws Exception {
            mockMvc.perform(delete("/api/admin/protected"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }
    }

    @Nested
    @DisplayName("路径匹配测试")
    class PathMatchingTests {

        @Test
        @WithAnonymousUser
        @DisplayName("Actuator子路径应该允许匿名访问")
        void actuatorSubPathsShouldAllowAnonymousAccess() throws Exception {
            mockMvc.perform(get("/actuator/info"))
                .andDo(print())
                .andExpect(status().isOk());
        }

        @Test
        @WithAnonymousUser
        @DisplayName("认证子路径应该允许匿名访问")
        void authSubPathsShouldAllowAnonymousAccess() throws Exception {
            mockMvc.perform(post("/api/admin/auth/register")
                    .contentType("application/json")
                    .content("{}"))
                .andDo(print())
                .andExpect(status().is(not(401)))
                .andExpect(status().is(not(403)));
        }

        @Test
        @WithAnonymousUser
        @DisplayName("其他API路径应该要求认证")
        void otherApiPathsShouldRequireAuthentication() throws Exception {
            mockMvc.perform(get("/api/admin/other"))
                .andDo(print())
                .andExpect(status().isUnauthorized());
        }
    }

    @Nested
    @DisplayName("安全头测试")
    class SecurityHeaderTests {

        @Test
        @WithAnonymousUser
        @DisplayName("响应应该包含安全头")
        void responseShouldIncludeSecurityHeaders() throws Exception {
            mockMvc.perform(get("/health"))
                .andDo(print())
                .andExpect(status().isOk())
                // Spring Security默认会添加一些安全头
                .andExpect(header().exists("X-Content-Type-Options"))
                .andExpect(header().exists("X-Frame-Options"));
        }
    }
}