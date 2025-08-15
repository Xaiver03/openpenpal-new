/**
 * useCommentsSOTA - SOTA Comment System Hook
 * SOTA评论系统钩子 - 支持多目标评论系统的完整React Hook
 */

import { useState, useCallback, useEffect, useRef } from 'react'
import { apiClient } from '@/lib/api-client'
import type {
  CommentSOTA,
  CommentStatsSOTA,
  CommentListQuerySOTA,
  CommentListResponseSOTA,
  CommentCreateRequestSOTA,
  CommentUpdateRequestSOTA,
  CommentReportRequestSOTA,
  CommentModerationRequestSOTA,
  UseCommentsSOTAOptions,
  UseCommentsSOTAReturn,
  CommentAPIResponseSOTA
} from '@/types/comment-sota'

// ================================
// SOTA Comment API Service
// ================================

class CommentAPIServiceSOTA {
  private baseURL = '/api/v2'

  // Load comments for a target
  async getComments(
    targetId: string, 
    targetType: string, 
    query: CommentListQuerySOTA
  ): Promise<CommentListResponseSOTA> {
    const params = new URLSearchParams()
    
    Object.entries(query).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        if (Array.isArray(value)) {
          value.forEach(v => params.append(key, String(v)))
        } else {
          params.append(key, String(value))
        }
      }
    })

    const url = `${this.baseURL}/targets/${targetType}/${targetId}/comments?${params}`
    const response = await apiClient.get<CommentAPIResponseSOTA<CommentListResponseSOTA>>(url)
    
    if (!response.data) throw new Error('No response data')
return response.data.data
  }

  // Create a new comment
  async createComment(request: CommentCreateRequestSOTA): Promise<CommentSOTA> {
    const response = await apiClient.post<CommentAPIResponseSOTA<CommentSOTA>>(
      `${this.baseURL}/comments`,
      request
    )
    
    if (!response.data) throw new Error('No response data')
return response.data.data
  }

  // Update an existing comment
  async updateComment(commentId: string, request: CommentUpdateRequestSOTA): Promise<CommentSOTA> {
    const response = await apiClient.put<CommentAPIResponseSOTA<CommentSOTA>>(
      `${this.baseURL}/comments/${commentId}`,
      request
    )
    
    if (!response.data) throw new Error('No response data')
return response.data.data
  }

  // Delete a comment
  async deleteComment(commentId: string): Promise<void> {
    await apiClient.delete(`${this.baseURL}/comments/${commentId}`)
  }

  // Like/unlike a comment
  async likeComment(commentId: string): Promise<{ is_liked: boolean; like_count: number }> {
    const response = await apiClient.post<CommentAPIResponseSOTA<{ is_liked: boolean; like_count: number }>>(
      `${this.baseURL}/comments/${commentId}/like`
    )
    
    if (!response.data) throw new Error('No response data')
return response.data.data
  }

  // Report a comment
  async reportComment(commentId: string, request: CommentReportRequestSOTA): Promise<void> {
    await apiClient.post(
      `${this.baseURL}/comments/${commentId}/report`,
      request
    )
  }

  // Moderate a comment (admin only)
  async moderateComment(commentId: string, request: CommentModerationRequestSOTA): Promise<void> {
    await apiClient.post(
      `${this.baseURL}/comments/${commentId}/moderate`,
      request
    )
  }

  // Get comment statistics
  async getStats(targetId: string, targetType: string): Promise<CommentStatsSOTA> {
    const response = await apiClient.get<CommentAPIResponseSOTA<CommentStatsSOTA>>(
      `${this.baseURL}/targets/${targetType}/${targetId}/comments/stats`
    )
    
    if (!response.data) throw new Error('No response data')
return response.data.data
  }

  // Load replies for a comment
  async getReplies(commentId: string, query?: Partial<CommentListQuerySOTA>): Promise<CommentSOTA[]> {
    const params = new URLSearchParams()
    
    if (query) {
      Object.entries(query).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, String(value))
        }
      })
    }

    const url = `${this.baseURL}/comments/${commentId}/replies?${params}`
    const response = await apiClient.get<CommentAPIResponseSOTA<CommentListResponseSOTA>>(url)
    
    if (!response.data) throw new Error('No response data')
return response.data.data.comments
  }
}

const commentAPI = new CommentAPIServiceSOTA()

// ================================
// SOTA Comment Hook
// ================================

export function useCommentsSOTA(options: UseCommentsSOTAOptions): UseCommentsSOTAReturn {
  const {
    target_id,
    target_type,
    initial_query = {},
    auto_load = true,
    enable_real_time = false,
    enable_polling = false,
    polling_interval = 30000
  } = options

  // State
  const [comments, setComments] = useState<CommentSOTA[]>([])
  const [stats, setStats] = useState<CommentStatsSOTA | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [has_more, setHasMore] = useState(false)
  const [current_query, setCurrentQuery] = useState<CommentListQuerySOTA>({
    target_id,
    target_type,
    ...initial_query
  })

  // Refs
  const polling_ref = useRef<NodeJS.Timeout | null>(null)
  const ws_ref = useRef<WebSocket | null>(null)

  // ================================
  // Core Actions
  // ================================

  // Load comments
  const load_comments = useCallback(async (query?: Partial<CommentListQuerySOTA>) => {
    try {
      setLoading(true)
      setError(null)

      const fullQuery = {
        ...current_query,
        ...query
      }

      const response = await commentAPI.getComments(target_id, target_type, fullQuery)
      
      // If this is a new page load, append to existing comments
      if (query?.page && query.page > 1) {
        setComments(prev => [...prev, ...response.comments])
      } else {
        setComments(response.comments)
      }
      
      setStats(response.stats)
      setHasMore(response.page < response.total_pages)
      setCurrentQuery(fullQuery)

    } catch (err: any) {
      setError(err.message || 'Failed to load comments')
      console.error('Failed to load comments:', err)
    } finally {
      setLoading(false)
    }
  }, [target_id, target_type, current_query])

  // Create comment
  const create_comment = useCallback(async (data: CommentCreateRequestSOTA): Promise<CommentSOTA> => {
    try {
      setError(null)
      
      const newComment = await commentAPI.createComment(data)
      
      // Add to local state
      if (data.parent_id) {
        // Handle reply - add to parent's replies
        setComments(prev => prev.map(comment => {
          if (comment.id === data.parent_id) {
            return {
              ...comment,
              replies: [...(comment.replies || []), newComment],
              reply_count: comment.reply_count + 1
            }
          } else if (comment.replies) {
            return {
              ...comment,
              replies: data.parent_id ? addReplyToTree(comment.replies, data.parent_id, newComment) : comment.replies
            }
          }
          return comment
        }))
      } else {
        // Add new top-level comment
        setComments(prev => [newComment, ...prev])
      }
      
      // Update stats
      setStats(prev => prev ? {
        ...prev,
        total_comments: prev.total_comments + 1,
        active_comments: prev.active_comments + 1
      } : null)

      return newComment
    } catch (err: any) {
      setError(err.message || 'Failed to create comment')
      throw err
    }
  }, [])

  // Update comment
  const update_comment = useCallback(async (id: string, data: CommentUpdateRequestSOTA): Promise<CommentSOTA> => {
    try {
      setError(null)
      
      const updatedComment = await commentAPI.updateComment(id, data)
      
      // Update in local state
      setComments(prev => updateCommentInTree(prev, id, updatedComment))
      
      return updatedComment
    } catch (err: any) {
      setError(err.message || 'Failed to update comment')
      throw err
    }
  }, [])

  // Delete comment
  const delete_comment = useCallback(async (id: string): Promise<void> => {
    try {
      setError(null)
      
      await commentAPI.deleteComment(id)
      
      // Remove from local state
      setComments(prev => removeCommentFromTree(prev, id))
      
      // Update stats
      setStats(prev => prev ? {
        ...prev,
        total_comments: prev.total_comments - 1,
        active_comments: prev.active_comments - 1
      } : null)

    } catch (err: any) {
      setError(err.message || 'Failed to delete comment')
      throw err
    }
  }, [])

  // Like comment
  const like_comment = useCallback(async (id: string): Promise<{ is_liked: boolean; like_count: number }> => {
    try {
      setError(null)
      
      const result = await commentAPI.likeComment(id)
      
      // Update in local state
      setComments(prev => updateCommentInTree(prev, id, {
        is_liked: result.is_liked,
        like_count: result.like_count
      }))
      
      return result
    } catch (err: any) {
      setError(err.message || 'Failed to like comment')
      throw err
    }
  }, [])

  // Report comment
  const report_comment = useCallback(async (id: string, data: CommentReportRequestSOTA): Promise<void> => {
    try {
      setError(null)
      
      await commentAPI.reportComment(id, data)
      
      // Update report count in local state
      setComments(prev => updateCommentInTree(prev, id, {
        report_count: (get_comment_by_id(id)?.report_count || 0) + 1,
        can_report: false
      }))

    } catch (err: any) {
      setError(err.message || 'Failed to report comment')
      throw err
    }
  }, [])

  // Moderate comment
  const moderate_comment = useCallback(async (id: string, data: CommentModerationRequestSOTA): Promise<void> => {
    try {
      setError(null)
      
      await commentAPI.moderateComment(id, data)
      
      // Update in local state
      setComments(prev => updateCommentInTree(prev, id, { status: data.status }))

    } catch (err: any) {
      setError(err.message || 'Failed to moderate comment')
      throw err
    }
  }, [])

  // Load replies
  const load_replies = useCallback(async (parent_id: string, query?: Partial<CommentListQuerySOTA>): Promise<CommentSOTA[]> => {
    try {
      setError(null)
      
      const replies = await commentAPI.getReplies(parent_id, query)
      
      // Update parent comment with loaded replies
      setComments(prev => updateCommentInTree(prev, parent_id, {
        replies: replies,
        is_loading_replies: false
      }))
      
      return replies
    } catch (err: any) {
      setError(err.message || 'Failed to load replies')
      throw err
    }
  }, [])

  // ================================
  // Utility Functions
  // ================================

  const refresh = useCallback(() => {
    return load_comments({ page: 1 })
  }, [load_comments])

  const clear_error = useCallback(() => {
    setError(null)
  }, [])

  const get_comment_by_id = useCallback((id: string): CommentSOTA | undefined => {
    return findCommentInTree(comments, id)
  }, [comments])

  const get_replies = useCallback((parent_id: string): CommentSOTA[] => {
    const parent = get_comment_by_id(parent_id)
    return parent?.replies || []
  }, [get_comment_by_id])

  const build_comment_tree = useCallback((flatComments: CommentSOTA[]): CommentSOTA[] => {
    const commentMap = new Map<string, CommentSOTA>()
    const rootComments: CommentSOTA[] = []
    
    // First pass: create comment map
    flatComments.forEach(comment => {
      commentMap.set(comment.id, { ...comment, replies: [] })
    })
    
    // Second pass: build tree
    flatComments.forEach(comment => {
      const commentWithReplies = commentMap.get(comment.id)!
      
      if (comment.parent_id) {
        const parent = commentMap.get(comment.parent_id)
        if (parent) {
          parent.replies = parent.replies || []
          parent.replies.push(commentWithReplies)
        }
      } else {
        rootComments.push(commentWithReplies)
      }
    })
    
    return rootComments
  }, [])

  const toggle_replies_expanded = useCallback((comment_id: string) => {
    setComments(prev => updateCommentInTree(prev, comment_id, comment => ({
      is_expanded: !comment.is_expanded
    })))
  }, [])

  // ================================
  // Effects
  // ================================

  // Auto load on mount
  useEffect(() => {
    if (auto_load) {
      load_comments()
    }
  }, [auto_load, load_comments])

  // Setup polling
  useEffect(() => {
    if (enable_polling && polling_interval > 0) {
      polling_ref.current = setInterval(() => {
        load_comments()
      }, polling_interval)
      
      return () => {
        if (polling_ref.current) {
          clearInterval(polling_ref.current)
        }
      }
    }
  }, [enable_polling, polling_interval, load_comments])

  // Setup WebSocket for real-time updates
  useEffect(() => {
    if (enable_real_time) {
      // TODO: Implement WebSocket connection for real-time updates
      // This would connect to the backend WebSocket endpoint for comment events
    }
    
    return () => {
      if (ws_ref.current) {
        ws_ref.current.close()
      }
    }
  }, [enable_real_time, target_id, target_type])

  return {
    // Data
    comments,
    stats,
    loading,
    error,
    has_more,
    
    // Actions
    load_comments,
    create_comment,
    update_comment,
    delete_comment,
    like_comment,
    report_comment,
    moderate_comment,
    load_replies,
    
    // Utilities
    refresh,
    clear_error,
    get_comment_by_id,
    get_replies,
    build_comment_tree,
    toggle_replies_expanded
  }
}

// ================================
// Helper Functions
// ================================

// Add reply to comment tree
function addReplyToTree(comments: CommentSOTA[], parent_id: string, reply: CommentSOTA): CommentSOTA[] {
  return comments.map(comment => {
    if (comment.id === parent_id) {
      return {
        ...comment,
        replies: [...(comment.replies || []), reply],
        reply_count: comment.reply_count + 1
      }
    } else if (comment.replies) {
      return {
        ...comment,
        replies: addReplyToTree(comment.replies, parent_id, reply)
      }
    }
    return comment
  })
}

// Update comment in tree
function updateCommentInTree(comments: CommentSOTA[], id: string, updates: Partial<CommentSOTA> | ((comment: CommentSOTA) => Partial<CommentSOTA>)): CommentSOTA[] {
  return comments.map(comment => {
    if (comment.id === id) {
      const updateData = typeof updates === 'function' ? updates(comment) : updates
      return { ...comment, ...updateData }
    } else if (comment.replies) {
      return {
        ...comment,
        replies: updateCommentInTree(comment.replies, id, updates)
      }
    }
    return comment
  })
}

// Remove comment from tree
function removeCommentFromTree(comments: CommentSOTA[], id: string): CommentSOTA[] {
  return comments.filter(comment => {
    if (comment.id === id) {
      return false
    }
    if (comment.replies) {
      comment.replies = removeCommentFromTree(comment.replies, id)
    }
    return true
  })
}

// Find comment in tree
function findCommentInTree(comments: CommentSOTA[], id: string): CommentSOTA | undefined {
  for (const comment of comments) {
    if (comment.id === id) {
      return comment
    }
    if (comment.replies) {
      const found = findCommentInTree(comment.replies, id)
      if (found) {
        return found
      }
    }
  }
  return undefined
}