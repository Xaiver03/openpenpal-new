/**
 * Fix synchronized models to use camelCase
 * Since backend now transforms responses to camelCase
 */

const fs = require('fs');
const path = require('path');

// Convert snake_case to camelCase
function snakeToCamelCase(str) {
  if (str === 'id' || str === 'ok') return str;
  
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
}

// Process TypeScript interface file
function processInterfaceFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8');
  
  // Replace snake_case field names with camelCase
  content = content.replace(/(\s+)(\w+_\w+)([?]?):/g, (match, spaces, fieldName, optional) => {
    const camelCase = snakeToCamelCase(fieldName);
    return `${spaces}${camelCase}${optional}:`;
  });
  
  // Update the header comment
  content = content.replace(
    /Generated on: .+/,
    `Generated on: ${new Date().toISOString()}`
  );
  
  // Add a note about camelCase transformation
  if (!content.includes('camelCase transformation')) {
    content = content.replace(
      '// DO NOT EDIT MANUALLY',
      '// DO NOT EDIT MANUALLY - Use sync-models.js to regenerate\n// Note: Field names use camelCase due to backend transformation middleware'
    );
  }
  
  fs.writeFileSync(filePath, content);
  console.log(`✅ Fixed model sync file: ${filePath}`);
}

// Main
const modelSyncPath = path.join(__dirname, '../frontend/src/types/models-sync.ts');

if (fs.existsSync(modelSyncPath)) {
  processInterfaceFile(modelSyncPath);
  
  // Also create a mapping utility for runtime transformation if needed
  const mappingUtilPath = path.join(__dirname, '../frontend/src/types/model-mapping.ts');
  const mappingUtil = `/**
 * Model field mapping utilities
 * Maps between snake_case (database) and camelCase (frontend)
 */

export const fieldMappings = {
  // User fields
  school_code: 'schoolCode',
  is_active: 'isActive',
  last_login_at: 'lastLoginAt',
  created_at: 'createdAt',
  updated_at: 'updatedAt',
  sent_letters: 'sentLetters',
  authored_letters: 'authoredLetters',
  
  // Letter fields
  user_id: 'userId',
  author_id: 'authorId',
  like_count: 'likeCount',
  recipient_op_code: 'recipientOpCode',
  sender_op_code: 'senderOpCode',
  share_count: 'shareCount',
  view_count: 'viewCount',
  reply_to: 'replyTo',
  envelope_id: 'envelopeId',
  status_logs: 'statusLogs',
  
  // Courier fields
  managed_op_code_prefix: 'managedOpCodePrefix',
  has_printer: 'hasPrinter',
  self_intro: 'selfIntro',
  can_mentor: 'canMentor',
  weekly_hours: 'weeklyHours',
  max_daily_tasks: 'maxDailyTasks',
  transport_method: 'transportMethod',
  time_slots: 'timeSlots',
  task_count: 'taskCount',
  deleted_at: 'deletedAt',
  
  // AI Config fields
  api_endpoint: 'apiEndpoint',
  max_tokens: 'maxTokens',
  daily_quota: 'dailyQuota',
  used_quota: 'usedQuota',
  quota_reset_at: 'quotaResetAt',
  
  // Museum fields
  source_type: 'sourceType',
  source_id: 'sourceId',
  submitted_by: 'submittedBy',
  approved_by: 'approvedBy',
  approved_at: 'approvedAt',
  origin_op_code: 'originOpCode',
  submitted_by_user: 'submittedByUser',
  approved_by_user: 'approvedByUser',
  
  // Museum Entry fields  
  letter_id: 'letterId',
  submission_id: 'submissionId',
  display_title: 'displayTitle',
  author_display_type: 'authorDisplayType',
  author_display_name: 'authorDisplayName',
  curator_type: 'curatorType',
  curator_id: 'curatorId',
  moderation_status: 'moderationStatus',
  bookmark_count: 'bookmarkCount'
} as const;

export type FieldMapping = typeof fieldMappings;
export type SnakeField = keyof FieldMapping;
export type CamelField = FieldMapping[SnakeField];
`;
  
  fs.writeFileSync(mappingUtilPath, mappingUtil);
  console.log(`✅ Created model mapping utility: ${mappingUtilPath}`);
  
} else {
  console.error('❌ Model sync file not found:', modelSyncPath);
}