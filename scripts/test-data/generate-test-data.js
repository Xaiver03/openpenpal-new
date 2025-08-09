#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 生成测试用户数据
function generateTestUsers(count = 10) {
  const roles = ['user', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'admin'];
  const users = [];
  
  for (let i = 0; i < count; i++) {
    users.push({
      id: `test-user-${i}`,
      username: `testuser${i}`,
      email: `test${i}@example.com`,
      password: 'Test123!',
      nickname: `测试用户${i}`,
      role: roles[i % roles.length],
      schoolCode: 'BJDX',
      isActive: true,
    });
  }
  
  return users;
}

// 生成测试信件数据
function generateTestLetters(count = 20) {
  const letters = [];
  const statuses = ['draft', 'generated', 'collected', 'in_transit', 'delivered'];
  
  for (let i = 0; i < count; i++) {
    letters.push({
      id: `test-letter-${i}`,
      title: `测试信件 ${i}`,
      content: `这是第 ${i} 封测试信件的内容...`,
      status: statuses[i % statuses.length],
      senderOPCode: `PK5F${String(i % 10).padStart(2, '0')}`,
      recipientOPCode: `PK3D${String(i % 10).padStart(2, '0')}`,
      createdAt: new Date(Date.now() - i * 86400000).toISOString(),
    });
  }
  
  return letters;
}

// 生成测试任务数据
function generateTestTasks(count = 15) {
  const tasks = [];
  const priorities = ['normal', 'urgent'];
  const statuses = ['pending', 'accepted', 'collected', 'in_transit', 'delivered'];
  
  for (let i = 0; i < count; i++) {
    tasks.push({
      id: `test-task-${i}`,
      letterCode: `LC${String(100000 + i).padStart(6, '0')}`,
      title: `投递任务 ${i}`,
      priority: priorities[i % priorities.length],
      status: statuses[i % statuses.length],
      pickupOPCode: `PK5F${String(i % 10).padStart(2, '0')}`,
      deliveryOPCode: `PK3D${String(i % 10).padStart(2, '0')}`,
      reward: 10 + (i % 5) * 5,
    });
  }
  
  return tasks;
}

// 保存测试数据
const testData = {
  users: generateTestUsers(),
  letters: generateTestLetters(),
  tasks: generateTestTasks(),
  timestamp: new Date().toISOString(),
};

// 输出到文件
const outputDir = path.join(__dirname, '../../test-data');
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

fs.writeFileSync(
  path.join(outputDir, 'test-data.json'),
  JSON.stringify(testData, null, 2)
);

console.log('✅ 测试数据已生成：test-data/test-data.json');
console.log(`📊 生成了 ${testData.users.length} 个用户，${testData.letters.length} 封信件，${testData.tasks.length} 个任务`);
