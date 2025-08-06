#!/usr/bin/env node

const net = require('net');
const { execSync } = require('child_process');

/**
 * 检查端口是否被占用
 * @param {number} port 端口号
 * @returns {Promise<boolean>} 是否被占用
 */
function checkPortInUse(port) {
  return new Promise((resolve) => {
    const server = net.createServer();
    
    server.listen(port, () => {
      server.close(() => {
        resolve(false); // 端口未被占用
      });
    });
    
    server.on('error', () => {
      resolve(true); // 端口被占用
    });
  });
}

/**
 * 获取占用端口的进程信息
 * @param {number} port 端口号
 * @returns {string|null} 进程信息
 */
function getPortProcess(port) {
  try {
    // macOS/Linux 使用 lsof
    const result = execSync(`lsof -ti:${port}`, { encoding: 'utf8', stdio: 'pipe' });
    const pid = result.trim();
    
    if (pid) {
      try {
        const processInfo = execSync(`ps -p ${pid} -o pid,ppid,comm`, { encoding: 'utf8', stdio: 'pipe' });
        return processInfo.trim();
      } catch {
        return `PID: ${pid}`;
      }
    }
  } catch (error) {
    // Windows 使用 netstat
    try {
      const result = execSync(`netstat -ano | findstr :${port}`, { encoding: 'utf8', stdio: 'pipe' });
      return result.trim();
    } catch {
      return null;
    }
  }
  return null;
}

/**
 * 查找可用端口
 * @param {number} startPort 起始端口
 * @param {number} maxTries 最大尝试次数
 * @returns {Promise<number>} 可用端口
 */
async function findAvailablePort(startPort = 3000, maxTries = 10) {
  for (let i = 0; i < maxTries; i++) {
    const port = startPort + i;
    const inUse = await checkPortInUse(port);
    
    if (!inUse) {
      return port;
    }
  }
  
  throw new Error(`无法找到可用端口 (尝试范围: ${startPort}-${startPort + maxTries - 1})`);
}

/**
 * 检查端口状态并返回详细信息
 * @param {number} port 要检查的端口
 * @returns {Promise<Object>} 端口状态信息
 */
async function checkPortStatus(port) {
  const inUse = await checkPortInUse(port);
  
  const result = {
    port,
    available: !inUse,
    processInfo: null
  };
  
  if (inUse) {
    result.processInfo = getPortProcess(port);
  }
  
  return result;
}

module.exports = {
  checkPortInUse,
  getPortProcess,
  findAvailablePort,
  checkPortStatus
};

// 如果直接运行此脚本
if (require.main === module) {
  const port = parseInt(process.argv[2]) || 3000;
  
  checkPortStatus(port).then(status => {
    console.log(JSON.stringify(status, null, 2));
  }).catch(error => {
    console.error('错误:', error.message);
    process.exit(1);
  });
}