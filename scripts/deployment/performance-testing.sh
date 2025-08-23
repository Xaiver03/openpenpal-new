#!/bin/bash
# OpenPenPal 性能测试套件
# 包含压力测试、基准测试、性能回归检测等
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="$SCRIPT_DIR/performance-results"

# 创建结果目录
mkdir -p "$RESULTS_DIR"/{k6,jmeter,lighthouse,benchmarks,reports}

# 测试配置
TEST_CONFIGS=(
    "smoke:10:30s:轻量级冒烟测试"
    "load:100:5m:正常负载测试"  
    "stress:500:10m:压力测试"
    "spike:1000:2m:突发测试"
    "volume:50:30m:容量测试"
)

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}🔍 检查性能测试依赖...${NC}"
    
    # 检查 k6
    if ! command -v k6 >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  k6 未安装，正在安装...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install k6 || echo -e "${RED}❌ k6 安装失败，请手动安装${NC}"
        else
            echo -e "${RED}❌ 请手动安装 k6: https://k6.io/docs/getting-started/installation${NC}"
        fi
    fi
    
    # 检查 curl
    if ! command -v curl >/dev/null 2>&1; then
        echo -e "${RED}❌ curl 未安装${NC}"
        exit 1
    fi
    
    # 检查 Node.js (for Lighthouse)
    if command -v node >/dev/null 2>&1; then
        if ! command -v lighthouse >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠️  Lighthouse 未安装，正在安装...${NC}"
            npm install -g lighthouse || echo -e "${RED}❌ Lighthouse 安装失败${NC}"
        fi
    fi
    
    # 检查 Apache Bench
    if ! command -v ab >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Apache Bench 未安装${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo "可通过 brew install apache2 安装"
        else
            echo "可通过 apt-get install apache2-utils 安装"
        fi
    fi
    
    echo -e "${GREEN}✅ 依赖检查完成${NC}"
}

# 创建 K6 测试脚本
create_k6_scripts() {
    echo -e "${BLUE}📝 创建 K6 测试脚本...${NC}"
    
    # API 端点测试
    cat > "$SCRIPT_DIR/k6/api-endpoints.js" << 'EOF'
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// 自定义指标
export let errorRate = new Rate('errors');
export let responseTimeTrend = new Trend('response_time', true);

// 测试配置
export let options = {
  scenarios: {
    api_test: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: __ENV.VUS || 10 },
        { duration: __ENV.DURATION || '2m', target: __ENV.VUS || 10 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% 的请求应在 500ms 内完成
    'http_req_failed': ['rate<0.1'],     // 错误率应低于 10%
    'errors': ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

// 测试用户凭据
const TEST_USER = {
  username: 'testuser',
  password: 'testpass123'
};

let authToken = '';

export function setup() {
  // 获取认证 token
  const loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (loginRes.status === 200) {
    authToken = JSON.parse(loginRes.body).token;
  }
  
  return { authToken };
}

export default function(data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': data.authToken ? `Bearer ${data.authToken}` : '',
  };
  
  group('健康检查', () => {
    const res = http.get(`${BASE_URL}/api/health`);
    check(res, {
      '健康检查状态 200': (r) => r.status === 200,
      '响应时间 < 100ms': (r) => r.timings.duration < 100,
    });
    responseTimeTrend.add(res.timings.duration);
    errorRate.add(res.status !== 200);
  });
  
  group('用户相关API', () => {
    // 获取用户信息
    const userRes = http.get(`${BASE_URL}/api/user/profile`, { headers });
    check(userRes, {
      '用户信息状态 200 或 401': (r) => [200, 401].includes(r.status),
    });
    responseTimeTrend.add(userRes.timings.duration);
    errorRate.add(![200, 401].includes(userRes.status));
    
    sleep(0.1);
  });
  
  group('信件相关API', () => {
    // 获取信件列表
    const lettersRes = http.get(`${BASE_URL}/api/letters`, { headers });
    check(lettersRes, {
      '信件列表状态 200 或 401': (r) => [200, 401].includes(r.status),
      '响应时间 < 500ms': (r) => r.timings.duration < 500,
    });
    responseTimeTrend.add(lettersRes.timings.duration);
    errorRate.add(![200, 401].includes(lettersRes.status));
    
    sleep(0.1);
  });
  
  group('快递相关API', () => {
    // 获取快递任务
    const courierRes = http.get(`${BASE_URL}/api/courier/tasks`, { headers });
    check(courierRes, {
      '快递任务状态 200 或 401': (r) => [200, 401].includes(r.status),
    });
    responseTimeTrend.add(courierRes.timings.duration);
    errorRate.add(![200, 401].includes(courierRes.status));
    
    sleep(0.1);
  });
  
  sleep(1);
}

export function teardown(data) {
  // 清理工作
  console.log('测试完成，正在清理...');
}
EOF

    # WebSocket 连接测试
    cat > "$SCRIPT_DIR/k6/websocket-test.js" << 'EOF'
import ws from 'k6/ws';
import { check } from 'k6';

export let options = {
  vus: __ENV.VUS || 10,
  duration: __ENV.DURATION || '30s',
};

const BASE_URL = __ENV.BASE_URL || 'ws://localhost:8080';

export default function () {
  const url = `${BASE_URL}/ws`;
  const params = { tags: { my_tag: 'websocket' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', function open() {
      console.log('WebSocket 连接已建立');
      
      // 发送测试消息
      socket.send(JSON.stringify({
        type: 'ping',
        timestamp: Date.now()
      }));
      
      // 每秒发送一次心跳
      socket.setInterval(function timeout() {
        socket.send(JSON.stringify({
          type: 'heartbeat',
          timestamp: Date.now()
        }));
      }, 1000);
    });

    socket.on('message', function (message) {
      const data = JSON.parse(message);
      check(data, {
        '消息类型正确': (msg) => ['pong', 'heartbeat_ack'].includes(msg.type),
      });
    });

    socket.on('close', function close() {
      console.log('WebSocket 连接已关闭');
    });

    socket.on('error', function (e) {
      if (e.error() != 'websocket: close sent') {
        console.log('WebSocket 错误: ', e.error());
      }
    });

    // 保持连接 30 秒
    socket.setTimeout(function () {
      console.log('关闭 WebSocket 连接');
      socket.close();
    }, 30000);
  });

  check(res, { 'WebSocket 连接状态': (r) => r && r.status === 101 });
}
EOF

    # 数据库压力测试
    cat > "$SCRIPT_DIR/k6/database-stress.js" << 'EOF'
import http from 'k6/http';
import { check, group } from 'k6';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export let options = {
  scenarios: {
    database_write: {
      executor: 'constant-arrival-rate',
      rate: __ENV.RATE || 50, // 每秒请求数
      timeUnit: '1s',
      duration: __ENV.DURATION || '2m',
      preAllocatedVUs: 10,
      maxVUs: 100,
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<1000'],
    'http_req_failed': ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

export default function () {
  const headers = {
    'Content-Type': 'application/json',
  };
  
  group('数据库写入测试', () => {
    // 模拟创建信件
    const letterData = {
      recipient_id: randomIntBetween(1, 1000),
      content: randomString(200, 'abcdefghijklmnopqrstuvwxyz '),
      sender_name: randomString(10),
      recipient_name: randomString(10),
    };
    
    const res = http.post(`${BASE_URL}/api/letters`, JSON.stringify(letterData), { headers });
    check(res, {
      '创建信件状态': (r) => [200, 201, 401].includes(r.status),
    });
  });
  
  group('数据库读取测试', () => {
    // 获取信件列表
    const res = http.get(`${BASE_URL}/api/letters?page=${randomIntBetween(1, 10)}&limit=20`);
    check(res, {
      '获取信件列表状态': (r) => [200, 401].includes(r.status),
    });
  });
}
EOF

    echo -e "${GREEN}✅ K6 测试脚本创建完成${NC}"
}

# 创建前端性能测试脚本
create_frontend_tests() {
    echo -e "${BLUE}📱 创建前端性能测试...${NC}"
    
    # Lighthouse 测试配置
    cat > "$SCRIPT_DIR/lighthouse/lighthouse-config.js" << 'EOF'
module.exports = {
  extends: 'lighthouse:default',
  settings: {
    onlyAudits: [
      'first-contentful-paint',
      'largest-contentful-paint',
      'first-meaningful-paint',
      'speed-index',
      'cumulative-layout-shift',
      'server-response-time',
      'interactive',
      'total-blocking-time',
    ],
  },
  audits: [
    'metrics/first-contentful-paint',
    'metrics/largest-contentful-paint',
    'metrics/first-meaningful-paint',
    'metrics/speed-index',
    'metrics/cumulative-layout-shift',
    'server-response-time',
    'metrics/interactive',
    'metrics/total-blocking-time',
  ],
  categories: {
    performance: {
      title: 'Performance',
      auditRefs: [
        {id: 'first-contentful-paint', weight: 10, group: 'metrics'},
        {id: 'largest-contentful-paint', weight: 25, group: 'metrics'},
        {id: 'first-meaningful-paint', weight: 10, group: 'metrics'},
        {id: 'speed-index', weight: 10, group: 'metrics'},
        {id: 'cumulative-layout-shift', weight: 15, group: 'metrics'},
        {id: 'server-response-time', weight: 5, group: 'load-opportunities'},
        {id: 'interactive', weight: 10, group: 'metrics'},
        {id: 'total-blocking-time', weight: 15, group: 'metrics'},
      ],
    },
  },
};
EOF

    # 前端性能测试脚本
    cat > "$SCRIPT_DIR/frontend-performance.js" << 'EOF'
const lighthouse = require('lighthouse');
const chromeLauncher = require('chrome-launcher');
const fs = require('fs');
const path = require('path');

const urls = [
  'http://localhost:3000',
  'http://localhost:3000/login',
  'http://localhost:3000/mailbox',
  'http://localhost:3000/courier',
];

const config = require('./lighthouse/lighthouse-config.js');

async function runLighthouseTests() {
  const results = [];
  
  for (const url of urls) {
    console.log(`测试页面: ${url}`);
    
    const chrome = await chromeLauncher.launch({chromeFlags: ['--headless']});
    const options = {
      logLevel: 'info',
      output: 'json',
      onlyCategories: ['performance'],
      port: chrome.port,
    };
    
    try {
      const runnerResult = await lighthouse(url, options, config);
      const reportJson = runnerResult.report;
      const report = JSON.parse(reportJson);
      
      results.push({
        url,
        score: report.categories.performance.score * 100,
        metrics: {
          'First Contentful Paint': report.audits['first-contentful-paint'].displayValue,
          'Largest Contentful Paint': report.audits['largest-contentful-paint'].displayValue,
          'Speed Index': report.audits['speed-index'].displayValue,
          'Cumulative Layout Shift': report.audits['cumulative-layout-shift'].displayValue,
          'Total Blocking Time': report.audits['total-blocking-time'].displayValue,
        }
      });
      
      // 保存详细报告
      const reportPath = `./performance-results/lighthouse/report-${url.replace(/[^a-zA-Z0-9]/g, '_')}.json`;
      fs.writeFileSync(reportPath, reportJson);
      
    } catch (error) {
      console.error(`测试 ${url} 失败:`, error.message);
    } finally {
      await chrome.kill();
    }
  }
  
  // 生成汇总报告
  const summary = {
    timestamp: new Date().toISOString(),
    results,
    averageScore: results.reduce((sum, r) => sum + r.score, 0) / results.length
  };
  
  fs.writeFileSync('./performance-results/lighthouse/summary.json', JSON.stringify(summary, null, 2));
  
  console.log('\n前端性能测试完成:');
  results.forEach(result => {
    console.log(`${result.url}: ${result.score.toFixed(1)}/100`);
  });
  console.log(`平均分数: ${summary.averageScore.toFixed(1)}/100`);
}

if (require.main === module) {
  runLighthouseTests().catch(console.error);
}

module.exports = { runLighthouseTests };
EOF

    echo -e "${GREEN}✅ 前端性能测试创建完成${NC}"
}

# 创建基准测试
create_benchmark_tests() {
    echo -e "${BLUE}📊 创建基准测试...${NC}"
    
    # API 基准测试
    cat > "$SCRIPT_DIR/benchmarks/api-benchmark.sh" << 'EOF'
#!/bin/bash
# API 性能基准测试

BASE_URL="${1:-http://localhost:8000}"
RESULTS_FILE="./performance-results/benchmarks/api-benchmark-$(date +%Y%m%d_%H%M%S).txt"

echo "OpenPenPal API 基准测试 - $(date)" > "$RESULTS_FILE"
echo "======================================" >> "$RESULTS_FILE"
echo "测试目标: $BASE_URL" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 测试函数
test_endpoint() {
    local name="$1"
    local url="$2"
    local method="${3:-GET}"
    
    echo "测试: $name" | tee -a "$RESULTS_FILE"
    echo "URL: $url" >> "$RESULTS_FILE"
    
    if command -v ab >/dev/null 2>&1; then
        # Apache Bench 测试
        echo "Apache Bench 结果:" >> "$RESULTS_FILE"
        ab -n 1000 -c 10 -q "$url" >> "$RESULTS_FILE" 2>&1
        echo "" >> "$RESULTS_FILE"
    fi
    
    if command -v curl >/dev/null 2>&1; then
        # cURL 响应时间测试
        echo "cURL 响应时间:" >> "$RESULTS_FILE"
        for i in {1..10}; do
            time=$(curl -w "%{time_total}" -s -o /dev/null "$url")
            echo "请求 $i: ${time}s" >> "$RESULTS_FILE"
        done
        echo "" >> "$RESULTS_FILE"
    fi
    
    echo "----------------------------------------" >> "$RESULTS_FILE"
}

# 执行测试
test_endpoint "健康检查" "$BASE_URL/api/health"
test_endpoint "首页" "$BASE_URL/"
test_endpoint "用户信息" "$BASE_URL/api/user/profile"
test_endpoint "信件列表" "$BASE_URL/api/letters"
test_endpoint "快递任务" "$BASE_URL/api/courier/tasks"

echo "基准测试完成，结果保存到: $RESULTS_FILE"
EOF

    chmod +x "$SCRIPT_DIR/benchmarks/api-benchmark.sh"
    
    # 数据库基准测试
    cat > "$SCRIPT_DIR/benchmarks/database-benchmark.sql" << 'EOF'
-- PostgreSQL 性能基准测试

-- 测试查询性能
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM letters 
WHERE created_at > NOW() - INTERVAL '30 days' 
ORDER BY created_at DESC 
LIMIT 100;

-- 测试连接查询性能
EXPLAIN (ANALYZE, BUFFERS) 
SELECT l.*, u.username 
FROM letters l 
JOIN users u ON l.sender_id = u.id 
WHERE l.status = 'delivered' 
LIMIT 100;

-- 测试聚合查询性能
EXPLAIN (ANALYZE, BUFFERS) 
SELECT DATE(created_at) as date, COUNT(*) as letter_count 
FROM letters 
WHERE created_at > NOW() - INTERVAL '7 days' 
GROUP BY DATE(created_at) 
ORDER BY date;

-- 测试快递任务查询性能
EXPLAIN (ANALYZE, BUFFERS) 
SELECT ct.*, c.username as courier_name 
FROM courier_tasks ct 
JOIN users c ON ct.courier_id = c.id 
WHERE ct.status = 'pending' 
ORDER BY ct.created_at 
LIMIT 50;
EOF

    echo -e "${GREEN}✅ 基准测试创建完成${NC}"
}

# 运行性能测试
run_performance_tests() {
    local test_type="${1:-all}"
    local intensity="${2:-load}"
    
    echo -e "${BLUE}🚀 开始性能测试 (类型: $test_type, 强度: $intensity)${NC}"
    
    # 解析测试强度配置
    local vus duration description
    for config in "${TEST_CONFIGS[@]}"; do
        IFS=':' read -r config_name config_vus config_duration config_desc <<< "$config"
        if [ "$config_name" = "$intensity" ]; then
            vus="$config_vus"
            duration="$config_duration"
            description="$config_desc"
            break
        fi
    done
    
    if [ -z "$vus" ]; then
        echo -e "${RED}❌ 未知的测试强度: $intensity${NC}"
        echo "可用强度: smoke, load, stress, spike, volume"
        return 1
    fi
    
    echo -e "${BLUE}📋 测试配置: $description (${vus} 用户, ${duration})${NC}"
    
    # 创建测试结果目录
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local test_dir="$RESULTS_DIR/$test_type-$intensity-$timestamp"
    mkdir -p "$test_dir"
    
    # K6 API 测试
    if [ "$test_type" = "all" ] || [ "$test_type" = "api" ]; then
        echo -e "${YELLOW}🔌 运行 API 性能测试...${NC}"
        if command -v k6 >/dev/null 2>&1; then
            k6 run --vus "$vus" --duration "$duration" \
                   --out json="$test_dir/api-test-results.json" \
                   "$SCRIPT_DIR/k6/api-endpoints.js"
        else
            echo -e "${RED}❌ k6 未安装，跳过 API 测试${NC}"
        fi
    fi
    
    # WebSocket 测试  
    if [ "$test_type" = "all" ] || [ "$test_type" = "websocket" ]; then
        echo -e "${YELLOW}🔗 运行 WebSocket 性能测试...${NC}"
        if command -v k6 >/dev/null 2>&1; then
            k6 run --vus "$((vus/2))" --duration "$duration" \
                   --out json="$test_dir/websocket-test-results.json" \
                   "$SCRIPT_DIR/k6/websocket-test.js"
        else
            echo -e "${RED}❌ k6 未安装，跳过 WebSocket 测试${NC}"
        fi
    fi
    
    # 数据库压力测试
    if [ "$test_type" = "all" ] || [ "$test_type" = "database" ]; then
        echo -e "${YELLOW}💾 运行数据库压力测试...${NC}"
        if command -v k6 >/dev/null 2>&1; then
            k6 run --env RATE="$((vus*2))" --env DURATION="$duration" \
                   --out json="$test_dir/database-stress-results.json" \
                   "$SCRIPT_DIR/k6/database-stress.js"
        else
            echo -e "${RED}❌ k6 未安装，跳过数据库测试${NC}"
        fi
    fi
    
    # 前端性能测试
    if [ "$test_type" = "all" ] || [ "$test_type" = "frontend" ]; then
        echo -e "${YELLOW}🎨 运行前端性能测试...${NC}"
        if command -v node >/dev/null 2>&1 && [ -f "$SCRIPT_DIR/frontend-performance.js" ]; then
            cd "$SCRIPT_DIR"
            node frontend-performance.js > "$test_dir/frontend-test.log" 2>&1
            cp -r performance-results/lighthouse/* "$test_dir/" 2>/dev/null || true
        else
            echo -e "${RED}❌ Node.js 或 Lighthouse 未安装，跳过前端测试${NC}"
        fi
    fi
    
    # 基准测试
    if [ "$test_type" = "all" ] || [ "$test_type" = "benchmark" ]; then
        echo -e "${YELLOW}📊 运行基准测试...${NC}"
        if [ -f "$SCRIPT_DIR/benchmarks/api-benchmark.sh" ]; then
            "$SCRIPT_DIR/benchmarks/api-benchmark.sh" > "$test_dir/benchmark-results.txt"
        fi
    fi
    
    # 生成测试报告
    generate_test_report "$test_dir" "$test_type" "$intensity" "$description"
    
    echo -e "${GREEN}✅ 性能测试完成，结果保存在: $test_dir${NC}"
}

# 生成测试报告
generate_test_report() {
    local test_dir="$1"
    local test_type="$2"
    local intensity="$3"
    local description="$4"
    
    echo -e "${BLUE}📋 生成测试报告...${NC}"
    
    cat > "$test_dir/report.md" << EOF
# OpenPenPal 性能测试报告

## 测试概述
- **测试时间**: $(date)
- **测试类型**: $test_type
- **测试强度**: $intensity ($description)
- **测试环境**: $(uname -a)

## 测试结果摘要

### K6 API 测试
EOF

    # 分析 K6 结果
    if [ -f "$test_dir/api-test-results.json" ]; then
        echo "分析 API 测试结果..." >> "$test_dir/report.md"
        
        # 提取关键指标
        if command -v jq >/dev/null 2>&1; then
            local avg_duration=$(jq -r '.metrics.http_req_duration.values.avg' "$test_dir/api-test-results.json" 2>/dev/null || echo "N/A")
            local p95_duration=$(jq -r '.metrics.http_req_duration.values."p(95)"' "$test_dir/api-test-results.json" 2>/dev/null || echo "N/A")
            local error_rate=$(jq -r '.metrics.http_req_failed.values.rate' "$test_dir/api-test-results.json" 2>/dev/null || echo "N/A")
            local total_requests=$(jq -r '.metrics.http_reqs.values.count' "$test_dir/api-test-results.json" 2>/dev/null || echo "N/A")
            
            cat >> "$test_dir/report.md" << EOF

- **平均响应时间**: ${avg_duration}ms
- **95% 分位数响应时间**: ${p95_duration}ms
- **错误率**: ${error_rate}
- **总请求数**: $total_requests

EOF
        fi
    else
        echo "API 测试未执行或失败" >> "$test_dir/report.md"
    fi
    
    # 添加建议
    cat >> "$test_dir/report.md" << EOF

## 性能建议

### 响应时间优化
- 如果 95% 响应时间 > 500ms，考虑优化数据库查询
- 如果平均响应时间 > 200ms，检查网络延迟和服务器负载

### 错误率优化  
- 如果错误率 > 5%，检查服务器资源和错误日志
- 如果出现大量 5xx 错误，检查服务器配置和依赖服务

### 并发处理优化
- 如果在高并发下性能下降明显，考虑增加服务器实例
- 检查数据库连接池配置和缓存策略

## 详细测试文件
$(ls -la "$test_dir" | grep -v "^d" | awk '{print "- " $9}')

---
*报告生成时间: $(date)*
EOF
    
    echo -e "${GREEN}✅ 测试报告生成完成: $test_dir/report.md${NC}"
}

# 性能基线管理
manage_baselines() {
    local action="${1:-list}"
    local baseline_dir="$RESULTS_DIR/baselines"
    
    mkdir -p "$baseline_dir"
    
    case "$action" in
        "create")
            echo -e "${BLUE}📏 创建性能基线...${NC}"
            
            # 运行基线测试
            run_performance_tests "all" "load"
            
            # 保存基线
            local latest_result=$(ls -t "$RESULTS_DIR"/all-load-* | head -1)
            if [ -n "$latest_result" ]; then
                cp -r "$latest_result" "$baseline_dir/baseline-$(date +%Y%m%d)"
                echo -e "${GREEN}✅ 基线创建完成${NC}"
            fi
            ;;
        "compare")
            echo -e "${BLUE}📊 比较当前性能与基线...${NC}"
            
            local latest_baseline=$(ls -t "$baseline_dir"/baseline-* | head -1)
            local latest_result=$(ls -t "$RESULTS_DIR"/all-load-* | head -1)
            
            if [ -n "$latest_baseline" ] && [ -n "$latest_result" ]; then
                echo "基线: $latest_baseline"
                echo "当前: $latest_result"
                
                # 简单比较 (需要 jq)
                if command -v jq >/dev/null 2>&1; then
                    local baseline_avg=$(jq -r '.metrics.http_req_duration.values.avg' "$latest_baseline/api-test-results.json" 2>/dev/null || echo "0")
                    local current_avg=$(jq -r '.metrics.http_req_duration.values.avg' "$latest_result/api-test-results.json" 2>/dev/null || echo "0")
                    
                    if [ "$baseline_avg" != "0" ] && [ "$current_avg" != "0" ]; then
                        local diff=$(echo "scale=2; ($current_avg - $baseline_avg) / $baseline_avg * 100" | bc -l 2>/dev/null || echo "0")
                        echo "平均响应时间变化: ${diff}%"
                        
                        if (( $(echo "$diff > 10" | bc -l) )); then
                            echo -e "${RED}⚠️  性能回归: 响应时间增加超过 10%${NC}"
                        elif (( $(echo "$diff < -10" | bc -l) )); then
                            echo -e "${GREEN}🎉 性能提升: 响应时间减少超过 10%${NC}"
                        else
                            echo -e "${GREEN}✅ 性能稳定${NC}"
                        fi
                    fi
                fi
            else
                echo -e "${YELLOW}⚠️  未找到基线或测试结果${NC}"
            fi
            ;;
        "list")
            echo -e "${BLUE}📋 可用基线:${NC}"
            ls -la "$baseline_dir" 2>/dev/null || echo "暂无基线"
            ;;
        *)
            echo "用法: $0 baseline {create|compare|list}"
            ;;
    esac
}

# 主函数
main() {
    case "${1:-}" in
        "setup")
            check_dependencies
            mkdir -p "$SCRIPT_DIR"/{k6,lighthouse,benchmarks}
            create_k6_scripts
            create_frontend_tests
            create_benchmark_tests
            echo -e "${GREEN}✅ 性能测试环境设置完成${NC}"
            ;;
        "test")
            local test_type="${2:-all}"
            local intensity="${3:-load}"
            run_performance_tests "$test_type" "$intensity"
            ;;
        "baseline")
            manage_baselines "${2:-list}"
            ;;
        "report")
            local test_dir="${2:-$(ls -t "$RESULTS_DIR"/all-* | head -1)}"
            if [ -n "$test_dir" ] && [ -d "$test_dir" ]; then
                echo -e "${BLUE}📊 查看测试报告: $test_dir/report.md${NC}"
                cat "$test_dir/report.md" 2>/dev/null || echo "报告文件不存在"
            else
                echo -e "${RED}❌ 未找到测试结果${NC}"
            fi
            ;;
        "clean")
            echo -e "${YELLOW}⚠️  确认清理所有测试结果? (y/N)${NC}"
            read -r response
            if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
                rm -rf "$RESULTS_DIR"
                echo -e "${GREEN}✅ 测试结果已清理${NC}"
            fi
            ;;
        *)
            echo -e "${BLUE}OpenPenPal 性能测试套件${NC}"
            echo ""
            echo "用法: $0 {setup|test|baseline|report|clean}"
            echo ""
            echo "命令:"
            echo "  setup                    - 设置测试环境"
            echo "  test [type] [intensity]  - 运行性能测试"
            echo "  baseline {create|compare|list} - 管理性能基线"
            echo "  report [test_dir]        - 查看测试报告"
            echo "  clean                    - 清理测试结果"
            echo ""
            echo "测试类型:"
            echo "  all                      - 全部测试"
            echo "  api                      - API 性能测试"
            echo "  websocket                - WebSocket 测试"
            echo "  database                 - 数据库压力测试"
            echo "  frontend                 - 前端性能测试"
            echo "  benchmark                - 基准测试"
            echo ""
            echo "测试强度:"
            for config in "${TEST_CONFIGS[@]}"; do
                IFS=':' read -r name vus duration desc <<< "$config"
                printf "  %-10s - %s (%s 用户, %s)\n" "$name" "$desc" "$vus" "$duration"
            done
            echo ""
            echo "示例:"
            echo "  $0 setup                 # 设置测试环境"
            echo "  $0 test api smoke        # 运行 API 冒烟测试"
            echo "  $0 test all stress       # 运行全面压力测试"
            echo "  $0 baseline create       # 创建性能基线"
            echo "  $0 baseline compare      # 比较当前性能与基线"
            ;;
    esac
}

main "$@"