import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const authHeader = request.headers.get('Authorization');
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json({
        success: false,
        error: 'Missing or invalid authorization header'
      }, { status: 401 });
    }
    
    const token = authHeader.substring(7);
    
    if (!token || !token.startsWith('mock_')) {
      return NextResponse.json({
        success: false,
        error: 'Invalid token'
      }, { status: 401 });
    }
    
    // Extract username from token
    const tokenParts = token.split('_');
    if (tokenParts.length < 2) {
      return NextResponse.json({
        success: false,
        error: 'Invalid token format'
      }, { status: 401 });
    }
    
    const username = tokenParts[1];
    
    // Mock permissions based on user roles
    const userPermissions: Record<string, string[]> = {
      'admin': [
        'MANAGE_USERS', 'VIEW_ANALYTICS', 'MODERATE_CONTENT',
        'MANAGE_SCHOOLS', 'MANAGE_EXHIBITIONS', 'SYSTEM_CONFIG',
        'AUDIT_SUBMISSIONS', 'HANDLE_REPORTS'
      ],
      'courier': [
        'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
        'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE'
      ],
      'senior': [
        'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
        'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE',
        'VIEW_REPORTS'
      ],
      'coordinator': [
        'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
        'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE',
        'MANAGE_COURIERS', 'ASSIGN_TASKS', 'VIEW_REPORTS'
      ]
    };
    
    const permissions = userPermissions[username] || [];
    
    return NextResponse.json({
      success: true,
      data: {
        permissions
      }
    });
    
  } catch (error) {
    console.error('Get user permissions error:', error);
    return NextResponse.json({
      success: false,
      error: 'Internal server error'
    }, { status: 500 });
  }
}