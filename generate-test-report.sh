#!/bin/bash

# 测试报告生成器

set -e

echo "📊 生成测试报告..."

# 创建报告目录
mkdir -p test-reports

# 生成时间戳
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
REPORT_FILE="test-reports/test-report-${TIMESTAMP}.html"

# 开始生成 HTML 报告
cat > $REPORT_FILE << 'HTML_START'
<!DOCTYPE html>
<html>
<head>
    <title>OpenPenPal 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        .summary { background: #f0f0f0; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .pass { color: green; }
        .fail { color: red; }
        .coverage { margin: 20px 0; }
        .coverage-bar { width: 300px; height: 20px; background: #ddd; border-radius: 10px; overflow: hidden; }
        .coverage-fill { height: 100%; background: #4CAF50; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #f0f0f0; }
    </style>
</head>
<body>
    <h1>OpenPenPal 测试报告</h1>
    <p>生成时间: <script>document.write(new Date().toLocaleString())</script></p>
HTML_START

# 收集后端测试结果
if [ -d "backend" ]; then
    echo "<h2>后端测试结果</h2>" >> $REPORT_FILE
    echo "<div class='summary'>" >> $REPORT_FILE
    cd backend
    go test ./... -json | go-test-report >> ../$REPORT_FILE 2>/dev/null || echo "<p>后端测试数据暂无</p>" >> ../$REPORT_FILE
    cd ..
    echo "</div>" >> $REPORT_FILE
fi

# 收集前端测试结果
if [ -d "frontend" ]; then
    echo "<h2>前端测试结果</h2>" >> $REPORT_FILE
    echo "<div class='summary'>" >> $REPORT_FILE
    if [ -f "frontend/coverage/coverage-summary.json" ]; then
        node -e "
        const coverage = require('./frontend/coverage/coverage-summary.json');
        const total = coverage.total;
        console.log('<table>');
        console.log('<tr><th>类型</th><th>覆盖率</th><th>覆盖/总数</th></tr>');
        ['lines', 'statements', 'functions', 'branches'].forEach(type => {
            const data = total[type];
            const pct = data.pct;
            const color = pct >= 80 ? 'pass' : pct >= 60 ? 'warning' : 'fail';
            console.log(\`<tr><td>\${type}</td><td class='\${color}'>\${pct}%</td><td>\${data.covered}/\${data.total}</td></tr>\`);
        });
        console.log('</table>');
        " >> $REPORT_FILE
    else
        echo "<p>前端测试数据暂无</p>" >> $REPORT_FILE
    fi
    echo "</div>" >> $REPORT_FILE
fi

# 结束 HTML
cat >> $REPORT_FILE << 'HTML_END'
    <h2>测试建议</h2>
    <ul>
        <li>确保所有关键路径都有测试覆盖</li>
        <li>为新功能编写测试用例</li>
        <li>定期运行测试确保代码质量</li>
        <li>目标：80% 以上的测试覆盖率</li>
    </ul>
</body>
</html>
HTML_END

echo "✅ 测试报告已生成: $REPORT_FILE"
open $REPORT_FILE 2>/dev/null || echo "请手动打开报告文件查看"
