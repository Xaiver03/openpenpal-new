/**
 * Update frontend files to use camelCase field names
 */

const fs = require('fs');
const path = require('path');
// Using built-in modules only

// Field mappings from snake_case to camelCase
const fieldMappings = {
  // User fields
  'school_code': 'schoolCode',
  'is_active': 'isActive',
  'last_login_at': 'lastLoginAt',
  'created_at': 'createdAt',
  'updated_at': 'updatedAt',
  'sent_letters': 'sentLetters',
  'authored_letters': 'authoredLetters',
  
  // Letter fields
  'user_id': 'userId',
  'author_id': 'authorId',
  'like_count': 'likeCount',
  'recipient_op_code': 'recipientOpCode',
  'sender_op_code': 'senderOpCode',
  'share_count': 'shareCount',
  'view_count': 'viewCount',
  'reply_to': 'replyTo',
  'envelope_id': 'envelopeId',
  'status_logs': 'statusLogs',
  
  // Courier fields
  'courier_id': 'courierId',
  'managed_op_code_prefix': 'managedOpCodePrefix',
  'has_printer': 'hasPrinter',
  'self_intro': 'selfIntro',
  'can_mentor': 'canMentor',
  'weekly_hours': 'weeklyHours',
  'max_daily_tasks': 'maxDailyTasks',
  'transport_method': 'transportMethod',
  'time_slots': 'timeSlots',
  'task_count': 'taskCount',
  'deleted_at': 'deletedAt',
  
  // AI Config fields
  'api_endpoint': 'apiEndpoint',
  'api_key': 'apiKey',
  'max_tokens': 'maxTokens',
  'daily_quota': 'dailyQuota',
  'used_quota': 'usedQuota',
  'quota_reset_at': 'quotaResetAt',
  
  // Museum fields
  'source_type': 'sourceType',
  'source_id': 'sourceId',
  'submitted_by': 'submittedBy',
  'approved_by': 'approvedBy',
  'approved_at': 'approvedAt',
  'origin_op_code': 'originOpCode',
  'submitted_by_user': 'submittedByUser',
  'approved_by_user': 'approvedByUser',
  
  // Museum Entry fields  
  'letter_id': 'letterId',
  'submission_id': 'submissionId',
  'display_title': 'displayTitle',
  'author_display_type': 'authorDisplayType',
  'author_display_name': 'authorDisplayName',
  'curator_type': 'curatorType',
  'curator_id': 'curatorId',
  'moderation_status': 'moderationStatus',
  'bookmark_count': 'bookmarkCount',
  
  // Task fields
  'task_id': 'taskId',
  'pickup_op_code': 'pickupOpCode',
  'delivery_op_code': 'deliveryOpCode',
  'current_op_code': 'currentOpCode',
  'pickup_time': 'pickupTime',
  'delivery_time': 'deliveryTime',
  'completed_at': 'completedAt',
  'estimated_delivery': 'estimatedDelivery',
  
  // Other common fields
  'phone_number': 'phoneNumber',
  'post_code': 'postCode',
  'op_code': 'opCode',
  'qr_code': 'qrCode',
  'qr_code_url': 'qrCodeUrl',
  'full_name': 'fullName',
  'display_name': 'displayName',
  'real_name': 'realName',
  'external_id': 'externalId',
  'refresh_token': 'refreshToken',
  'access_token': 'accessToken',
  'csrf_token': 'csrfToken',
  'expires_at': 'expiresAt',
  'expires_in': 'expiresIn',
  'remember_me': 'rememberMe',
  'is_admin': 'isAdmin',
  'is_verified': 'isVerified',
  'verification_code': 'verificationCode',
  'reset_token': 'resetToken',
  'email_verified': 'emailVerified',
  'phone_verified': 'phoneVerified'
};

// Files/directories to exclude
const excludePatterns = [
  '**/node_modules/**',
  '**/dist/**',
  '**/build/**',
  '**/.next/**',
  '**/coverage/**',
  '**/*.test.*',
  '**/*.spec.*',
  '**/__tests__/**',
  '**/test/**',
  '**/tests/**',
  '**/model-mapping.ts', // Don't update our mapping file
  '**/models-sync.ts'    // Already fixed
];

// Update a single file
function updateFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8');
  let hasChanges = false;
  
  // Skip files that shouldn't be modified
  if (content.includes('// AUTO-GENERATED') || 
      content.includes('// DO NOT EDIT') ||
      content.includes('mock') ||
      filePath.includes('.json')) {
    return false;
  }
  
  // Replace field names in object properties and destructuring
  Object.entries(fieldMappings).forEach(([snakeCase, camelCase]) => {
    // Match patterns like: .school_code, ['school_code'], { school_code }, "school_code":
    const patterns = [
      // Object property access: obj.school_code
      new RegExp(`\\.${snakeCase}(?![a-zA-Z0-9_])`, 'g'),
      // Bracket notation: obj['school_code'] or obj["school_code"]
      new RegExp(`\\[['"]${snakeCase}['"]\\]`, 'g'),
      // Object destructuring: { school_code } or { school_code: alias }
      new RegExp(`([{,]\\s*)${snakeCase}(\\s*[}:,])`, 'g'),
      // Object property definition: school_code: value
      new RegExp(`^(\\s*)${snakeCase}(\\s*:)`, 'gm'),
      // In template literals: ${user.school_code}
      new RegExp(`\\$\\{([^}]*\\.)${snakeCase}(\\s*[}])`, 'g'),
      // JSON-like property: "school_code":
      new RegExp(`"${snakeCase}"(\\s*:)`, 'g')
    ];
    
    patterns.forEach((pattern, index) => {
      const before = content;
      switch(index) {
        case 0: // Property access
          content = content.replace(pattern, `.${camelCase}`);
          break;
        case 1: // Bracket notation
          content = content.replace(pattern, `['${camelCase}']`);
          break;
        case 2: // Object destructuring
          content = content.replace(pattern, `$1${camelCase}$2`);
          break;
        case 3: // Object property definition
          content = content.replace(pattern, `$1${camelCase}$2`);
          break;
        case 4: // Template literals
          content = content.replace(pattern, `\${$1${camelCase}$2`);
          break;
        case 5: // JSON-like
          content = content.replace(pattern, `"${camelCase}"$1`);
          break;
      }
      if (before !== content) hasChanges = true;
    });
  });
  
  if (hasChanges) {
    fs.writeFileSync(filePath, content);
    console.log(`✅ Updated: ${filePath}`);
    return true;
  }
  
  return false;
}

// Recursively find all TypeScript files
function findFiles(dir, extensions, excludes = []) {
  const files = [];
  
  function traverse(currentDir) {
    const entries = fs.readdirSync(currentDir, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(currentDir, entry.name);
      
      // Check excludes
      const shouldExclude = excludes.some(pattern => {
        return entry.name === pattern || 
               fullPath.includes('/' + pattern + '/') ||
               fullPath.includes('/' + pattern);
      });
      
      if (shouldExclude) continue;
      
      if (entry.isDirectory()) {
        traverse(fullPath);
      } else if (extensions.includes(path.extname(entry.name))) {
        files.push(fullPath);
      }
    }
  }
  
  traverse(dir);
  return files;
}

// Main execution
console.log('🔍 Searching for TypeScript/TSX files to update...\n');

const frontendDir = path.join(__dirname, '../frontend/src');
const excludeDirs = ['node_modules', 'dist', 'build', '.next', 'coverage', '__tests__', 'test', 'tests'];
const files = findFiles(frontendDir, ['.ts', '.tsx'], excludeDirs);

let totalFiles = 0;
let updatedFiles = 0;

files.forEach(file => {
  // Skip specific files
  if (file.includes('model-mapping.ts') || 
      file.includes('models-sync.ts') ||
      file.includes('.test.') ||
      file.includes('.spec.')) {
    return;
  }
  
  totalFiles++;
  if (updateFile(file)) {
    updatedFiles++;
  }
});

console.log(`\n📊 Summary:`);
console.log(`   Total files scanned: ${totalFiles}`);
console.log(`   Files updated: ${updatedFiles}`);

if (updatedFiles > 0) {
  console.log('\n✨ Frontend files have been updated to use camelCase field names!');
  console.log('   The backend middleware will automatically transform responses.');
} else {
  console.log('\n✅ No files needed updating - frontend is already using camelCase!');
}