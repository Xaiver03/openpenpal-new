#!/usr/bin/env node

/**
 * 前端API修复验证脚本
 * 验证 AdminService 中的 API 路径修复
 */

const fs = require('fs');
const path = require('path');

console.log('🔍 验证前端 AdminService API 路径修复...\n');

// 读取前端 AdminService 文件
const adminServicePath = '../frontend/src/lib/services/admin-service.ts';
const fullPath = path.resolve(__dirname, adminServicePath);

if (!fs.existsSync(fullPath)) {
  console.error('❌ AdminService 文件未找到:', fullPath);
  process.exit(1);
}

const adminServiceContent = fs.readFileSync(fullPath, 'utf8');

// 验证检查项目
const validationChecks = [
  {
    name: '用户管理 API 路径',
    pattern: /\/api\/v1\/admin\/users/g,
    minOccurrences: 3,
    description: '检查用户管理相关的API是否使用正确的路径前缀'
  },
  {
    name: '信件管理 API 路径', 
    pattern: /\/api\/v1\/admin\/letters/g,
    minOccurrences: 2,
    description: '检查信件管理相关的API是否使用正确的路径前缀'
  },
  {
    name: '信使管理 API 路径',
    pattern: /\/api\/v1\/admin\/couriers/g,
    minOccurrences: 1,
    description: '检查信使管理相关的API是否使用正确的路径前缀'
  },
  {
    name: '仪表板 API 路径',
    pattern: /\/api\/v1\/admin\/dashboard/g,
    minOccurrences: 1,
    description: '检查仪表板API是否使用正确的路径前缀'
  },
  {
    name: '系统设置 API 路径',
    pattern: /\/api\/v1\/admin\/settings/g,
    minOccurrences: 1,
    description: '检查系统设置API是否使用正确的路径前缀'
  },
  {
    name: '无遗留的错误路径',
    pattern: /\/admin\/(?!api)/g,
    maxOccurrences: 0,
    description: '确保没有遗留的错误API路径（缺少/api/v1前缀）'
  }
];

let totalChecks = 0;
let passedChecks = 0;

console.log('📋 执行验证检查:\n');

validationChecks.forEach(check => {
  totalChecks++;
  const matches = adminServiceContent.match(check.pattern) || [];
  const occurrences = matches.length;
  
  let passed = false;
  if (check.minOccurrences !== undefined) {
    passed = occurrences >= check.minOccurrences;
  } else if (check.maxOccurrences !== undefined) {
    passed = occurrences <= check.maxOccurrences;
  }
  
  if (passed) {
    console.log(`✅ ${check.name}`);
    console.log(`   ↳ 找到 ${occurrences} 个匹配项 ${check.description}`);
    passedChecks++;
  } else {
    console.log(`❌ ${check.name}`);
    console.log(`   ↳ 找到 ${occurrences} 个匹配项，${check.description}`);
    if (check.minOccurrences !== undefined) {
      console.log(`   ↳ 期望至少 ${check.minOccurrences} 个`);
    }
    if (check.maxOccurrences !== undefined) {
      console.log(`   ↳ 期望最多 ${check.maxOccurrences} 个`);
    }
  }
  console.log();
});

// 附加检查：统计所有API方法
console.log('📊 AdminService API 方法统计:');
const apiMethods = [
  'getDashboardStats',
  'getUsers', 
  'updateUser',
  'deleteUser',
  'getLetters',
  'moderateLetter',
  'getCouriers',
  'getSettings',
  'updateSettings'
];

let definedMethods = 0;
apiMethods.forEach(method => {
  if (adminServiceContent.includes(`static async ${method}`)) {
    console.log(`   ✅ ${method}()`);
    definedMethods++;
  } else {
    console.log(`   ❌ ${method}() - 未找到`);
  }
});

console.log(`\n📈 方法定义完整性: ${definedMethods}/${apiMethods.length} (${((definedMethods/apiMethods.length)*100).toFixed(1)}%)`);

// 最终结果
console.log('\n🎯 验证结果总结:');
console.log(`路径检查: ${passedChecks}/${totalChecks} 通过`);
console.log(`成功率: ${((passedChecks/totalChecks) * 100).toFixed(1)}%`);

if (passedChecks === totalChecks && definedMethods === apiMethods.length) {
  console.log('\n🎉 前端 AdminService 修复验证完全通过！');
  console.log('✨ 所有 API 路径都已正确修复为 /api/v1/admin/* 格式');
  console.log('✨ 所有必要的 API 方法都已定义');
} else {
  console.log('\n⚠️  部分验证失败，需要进一步检查:');
  if (passedChecks < totalChecks) {
    console.log('   - API 路径可能还有未修复的问题');
  }
  if (definedMethods < apiMethods.length) {
    console.log('   - 部分 API 方法定义缺失');
  }
}

console.log('\n🔗 相关文件路径:');
console.log(`   AdminService: ${adminServicePath}`);
