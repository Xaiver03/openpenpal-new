/**
 * Operation Log Types - SOTA Clean Type Definitions
 * Tracking system operations for audit and monitoring
 */

export interface OperationLog {
  id: string;
  userId: string;
  operation_type: string;
  resource_type: string;
  resource_id: string;
  description: string;
  ip_address: string;
  user_agent: string;
  status: 'success' | 'failed' | 'pending';
  metadata?: Record<string, any>;
  createdAt: string;
}

export interface OperationLogFilter {
  user_id?: string;
  operation_type?: string;
  resource_type?: string;
  status?: 'success' | 'failed' | 'pending';
  start_date?: string;
  end_date?: string;
  page?: number;
  limit?: number;
}

export interface OperationLogStats {
  total_operations: number;
  success_rate: number;
  failed_operations: number;
  operations_by_type: Record<string, number>;
  operations_by_date: Array<{
    date: string;
    count: number;
  }>;
}

export interface OperationLogResponse {
  logs: OperationLog[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export interface CreateOperationLogRequest {
  operation_type: string;
  resource_type: string;
  resource_id: string;
  description: string;
  metadata?: Record<string, any>;
}