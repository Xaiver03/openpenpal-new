#!/usr/bin/env node

const { spawn, exec } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

/**
 * OpenPenPal JavaScript启动器
 * 通过JS调用.command文件，并集成终端检查结果
 */
class OpenPenPalJSLauncher {
  constructor() {
    this.projectRoot = process.cwd();
    this.commandFile = path.join(this.projectRoot, 'start-openpenpal.command');
    this.results = {
      success: false,
      platform: os.platform(),
      checks: [],
      errors: [],
      warnings: [],
      terminalOutput: [],
      commandProcess: null,
      startTime: new Date(),
      endTime: null
    };
  }

  // 输出彩色日志
  log(message, color = 'reset') {
    const colors = {
      red: '\x1b[31m',
      green: '\x1b[32m',
      yellow: '\x1b[33m',
      blue: '\x1b[34m',
      magenta: '\x1b[35m',
      cyan: '\x1b[36m',
      reset: '\x1b[0m',
      bold: '\x1b[1m'
    };
    
    const timestamp = new Date().toLocaleTimeString();
    const logMessage = `[${timestamp}] ${message}`;
    
    console.log(`${colors[color]}${logMessage}${colors.reset}`);
    this.results.terminalOutput.push({ timestamp, message, color });
  }

  // 添加检查结果
  addCheck(name, status, message, details = null) {
    const check = {
      name,
      status,
      message,
      details,
      timestamp: new Date()
    };
    
    this.results.checks.push(check);
    
    const statusIcon = {
      success: '✅',
      warning: '⚠️',
      error: '❌',
      info: 'ℹ️'
    };
    
    const color = {
      success: 'green',
      warning: 'yellow',
      error: 'red',
      info: 'blue'
    }[status] || 'reset';
    
    this.log(`${statusIcon[status]} ${message}`, color);
    
    if (details) {
      this.log(`    详情: ${details}`, 'cyan');
    }
  }

  // 添加错误
  addError(error) {
    this.results.errors.push({
      message: error,
      timestamp: new Date()
    });
    this.log(`❌ 错误: ${error}`, 'red');
  }

  // 添加警告
  addWarning(warning) {
    this.results.warnings.push({
      message: warning,
      timestamp: new Date()
    });
    this.log(`⚠️  警告: ${warning}`, 'yellow');
  }

  // 检查平台兼容性
  checkPlatform() {
    const platform = os.platform();
    
    if (platform === 'darwin') {
      this.addCheck('platform', 'success', 'macOS平台兼容', `系统: ${os.type()} ${os.release()}`);
      return true;
    } else {
      this.addCheck('platform', 'error', '不支持的平台', `当前平台: ${platform}，需要macOS`);
      return false;
    }
  }

  // 检查.command文件
  checkCommandFile() {
    if (fs.existsSync(this.commandFile)) {
      // 检查文件权限
      try {
        const stats = fs.statSync(this.commandFile);
        const isExecutable = (stats.mode & parseInt('111', 8)) !== 0;
        
        if (isExecutable) {
          this.addCheck('command_file', 'success', '.command文件检查通过', `文件: ${path.basename(this.commandFile)}`);
          return true;
        } else {
          this.addWarning('.command文件没有执行权限，正在修复...');
          fs.chmodSync(this.commandFile, '755');
          this.addCheck('command_file', 'success', '.command文件权限已修复');
          return true;
        }
      } catch (error) {
        this.addCheck('command_file', 'error', '.command文件权限检查失败', error.message);
        return false;
      }
    } else {
      this.addCheck('command_file', 'error', '.command文件不存在', `期望路径: ${this.commandFile}`);
      return false;
    }
  }

  // 预检查系统环境
  async preCheckEnvironment() {
    const checks = [
      {
        name: 'node',
        command: 'node --version',
        description: 'Node.js版本'
      },
      {
        name: 'npm',
        command: 'npm --version',
        description: 'npm版本'
      },
      {
        name: 'git',
        command: 'git --version',
        description: 'Git版本'
      }
    ];

    for (const check of checks) {
      try {
        const result = await this.executeCommand(check.command);
        this.addCheck(check.name, 'success', `${check.description}: ${result.trim()}`);
      } catch (error) {
        if (check.name === 'node' || check.name === 'npm') {
          this.addCheck(check.name, 'error', `${check.description}检查失败`, error.message);
          return false;
        } else {
          this.addCheck(check.name, 'warning', `${check.description}未安装`);
        }
      }
    }
    
    return true;
  }

  // 执行命令并返回结果
  executeCommand(command) {
    return new Promise((resolve, reject) => {
      exec(command, (error, stdout, stderr) => {
        if (error) {
          reject(error);
        } else {
          resolve(stdout);
        }
      });
    });
  }

  // 启动.command文件
  async launchCommandFile() {
    return new Promise((resolve, reject) => {
      this.log('🚀 启动.command脚本...', 'blue');
      
      // 在新的终端窗口中运行.command文件
      const command = `open -a Terminal "${this.commandFile}"`;
      
      exec(command, (error, stdout, stderr) => {
        if (error) {
          this.addError(`启动.command文件失败: ${error.message}`);
          reject(error);
        } else {
          this.addCheck('launch', 'success', '.command文件已在新终端窗口中启动');
          
          // 监控进程（简单版本）
          this.monitorProcess();
          resolve();
        }
      });
    });
  }

  // 监控进程状态
  monitorProcess() {
    this.log('📊 开始监控进程状态...', 'cyan');
    
    // 每5秒检查一次是否有Next.js进程在运行
    const checkInterval = setInterval(() => {
      exec('pgrep -f "next"', (error, stdout) => {
        if (stdout.trim()) {
          this.log('✅ Next.js开发服务器正在运行', 'green');
          
          // 检查端口
          this.checkPorts();
        } else {
          this.log('ℹ️  等待开发服务器启动...', 'blue');
        }
      });
    }, 5000);

    // 30秒后停止监控
    setTimeout(() => {
      clearInterval(checkInterval);
      this.log('📊 监控结束', 'cyan');
    }, 30000);
  }

  // 检查端口状态
  checkPorts() {
    const ports = [3000, 3001, 3002, 3003];
    
    ports.forEach(port => {
      exec(`lsof -Pi :${port} -sTCP:LISTEN`, (error, stdout) => {
        if (!error && stdout.trim()) {
          const lines = stdout.split('\n');
          if (lines.length > 1) {
            const processLine = lines[1];
            const processName = processLine.split(/\s+/)[0];
            
            if (processName.includes('node') || processName.includes('next')) {
              this.addCheck('port_status', 'success', `端口${port}被Next.js占用`, `进程: ${processName}`);
              
              // 尝试打开浏览器
              const url = `http://localhost:${port}`;
              setTimeout(() => {
                exec(`open "${url}"`, (error) => {
                  if (!error) {
                    this.addCheck('browser', 'success', '浏览器已自动打开', url);
                  }
                });
              }, 3000);
            }
          }
        }
      });
    });
  }

  // 生成启动报告
  generateReport() {
    this.results.endTime = new Date();
    const duration = this.results.endTime - this.results.startTime;
    
    const report = {
      ...this.results,
      duration: `${(duration / 1000).toFixed(2)}秒`,
      summary: {
        totalChecks: this.results.checks.length,
        successfulChecks: this.results.checks.filter(c => c.status === 'success').length,
        warnings: this.results.warnings.length,
        errors: this.results.errors.length,
        platform: this.results.platform
      }
    };

    return report;
  }

  // 保存报告到文件
  saveReport(report) {
    const reportFile = path.join(this.projectRoot, 'launch-report.json');
    
    try {
      fs.writeFileSync(reportFile, JSON.stringify(report, null, 2));
      this.log(`📄 启动报告已保存: ${reportFile}`, 'cyan');
    } catch (error) {
      this.addWarning(`无法保存启动报告: ${error.message}`);
    }
  }

  // 显示最终状态
  showFinalStatus() {
    console.log('\n' + '='.repeat(60));
    this.log('📋 启动完成总结', 'bold');
    console.log('='.repeat(60));
    
    const report = this.generateReport();
    
    this.log(`⏱️  总耗时: ${report.duration}`, 'cyan');
    this.log(`✅ 成功检查: ${report.summary.successfulChecks}/${report.summary.totalChecks}`, 'green');
    
    if (report.summary.warnings > 0) {
      this.log(`⚠️  警告: ${report.summary.warnings}个`, 'yellow');
    }
    
    if (report.summary.errors > 0) {
      this.log(`❌ 错误: ${report.summary.errors}个`, 'red');
    }

    console.log('\n🎯 下一步操作:');
    console.log('   • 查看新打开的终端窗口中的开发服务器状态');
    console.log('   • 等待浏览器自动打开或手动访问 http://localhost:3000');
    console.log('   • 按 Ctrl+C 停止开发服务器');
    
    console.log('\n📚 获得帮助:');
    console.log('   • 查看启动报告: cat launch-report.json');
    console.log('   • 阅读文档: docs/启动脚本使用指南.md');
    console.log('   • 运行测试: node test-launch.js');
    
    // 保存报告
    this.saveReport(report);
    
    return report;
  }

  // 主启动流程
  async launch() {
    console.clear();
    
    this.log('╔══════════════════════════════════════════════════════════╗', 'magenta');
    this.log('║  📮 OpenPenPal 信使计划 - JavaScript启动器              ║', 'magenta');
    this.log('║     通过JS调用.command文件并集成终端检查                 ║', 'magenta');
    this.log('╚══════════════════════════════════════════════════════════╝', 'magenta');
    console.log('');

    try {
      // 1. 检查平台兼容性
      this.log('[1/5] 检查平台兼容性...', 'blue');
      if (!this.checkPlatform()) {
        throw new Error('平台不兼容');
      }

      // 2. 检查.command文件
      this.log('\n[2/5] 检查.command文件...', 'blue');
      if (!this.checkCommandFile()) {
        throw new Error('.command文件检查失败');
      }

      // 3. 预检查环境
      this.log('\n[3/5] 预检查系统环境...', 'blue');
      if (!(await this.preCheckEnvironment())) {
        throw new Error('系统环境检查失败');
      }

      // 4. 启动.command文件
      this.log('\n[4/5] 启动.command脚本...', 'blue');
      await this.launchCommandFile();

      // 5. 完成
      this.log('\n[5/5] 启动流程完成', 'blue');
      this.results.success = true;
      
      // 等待一段时间让.command脚本启动
      await new Promise(resolve => setTimeout(resolve, 2000));
      
    } catch (error) {
      this.addError(error.message);
      this.results.success = false;
    }

    // 显示最终状态
    return this.showFinalStatus();
  }
}

// 导出类
module.exports = OpenPenPalJSLauncher;

// 如果直接运行此脚本
if (require.main === module) {
  const launcher = new OpenPenPalJSLauncher();
  
  // 优雅退出处理
  process.on('SIGINT', () => {
    console.log('\n\n收到退出信号...');
    launcher.log('🛑 JavaScript启动器正在退出', 'yellow');
    process.exit(0);
  });

  // 启动
  launcher.launch().then((report) => {
    if (report.summary.errors === 0) {
      launcher.log('🎉 启动成功！请查看新打开的终端窗口', 'green');
    } else {
      launcher.log('⚠️  启动过程中有错误，请查看上述信息', 'yellow');
    }
  }).catch((error) => {
    console.error('❌ 启动器异常:', error.message);
    process.exit(1);
  });
}