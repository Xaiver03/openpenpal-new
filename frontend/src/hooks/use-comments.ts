// Custom hook for comment management with state and real-time updates
// 评论管理自定义Hook - 包含状态管理和实时更新

import { useState, useCallback, useEffect, useMemo } from 'react';
import { commentApi } from '@/lib/api/comment';
import type {
  Comment,
  CommentCreateRequest,
  CommentUpdateRequest,
  CommentListQuery,
  CommentListResponse,
  CommentStatsResponse,
  UseCommentsOptions,
  UseCommentsReturn,
  CommentError,
} from '@/types/comment';

/**
 * Custom hook for comprehensive comment management
 */
export function useComments(options: UseCommentsOptions): UseCommentsReturn {
  const {
    letter_id,
    initial_query = {},
    auto_load = true,
    enable_real_time = false,
  } = options;

  // State management
  const [comments, setComments] = useState<Comment[]>([]);
  const [stats, setStats] = useState<CommentStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [current_query, setCurrentQuery] = useState<CommentListQuery>({
    page: 1,
    limit: 20,
    sort_by: 'created_at',
    order: 'desc',
    only_top_level: false,
    ...initial_query,
  });
  const [pagination, setPagination] = useState<CommentListResponse['pagination'] | null>(null);

  // Computed state
  const has_more = useMemo(() => {
    if (!pagination) return false;
    return pagination.page < pagination.pages;
  }, [pagination]);

  // Error handling
  const handle_error = useCallback((err: any) => {
    const error_message = err instanceof Error ? err.message : 'An unknown error occurred';
    setError(error_message);
    console.error('Comment operation failed:', err);
  }, []);

  // Clear error state
  const clear_error = useCallback(() => {
    setError(null);
  }, []);

  // Load comments with query parameters
  const load_comments = useCallback(async (query: Partial<CommentListQuery> = {}) => {
    setLoading(true);
    clear_error();

    try {
      const merged_query = { ...current_query, ...query };
      const response = await commentApi.getCommentsByLetter(letter_id, merged_query);
      
      if (merged_query.page === 1) {
        // Replace comments for first page or new query
        setComments(response.comments);
      } else {
        // Append comments for subsequent pages
        setComments(prev => [...prev, ...response.comments]);
      }
      
      setPagination(response.pagination);
      setCurrentQuery(merged_query);
    } catch (err) {
      handle_error(err);
    } finally {
      setLoading(false);
    }
  }, [letter_id, current_query, clear_error, handle_error]);

  // Load comment statistics
  const load_stats = useCallback(async () => {
    try {
      const stats_data = await commentApi.getCommentStats(letter_id);
      setStats(stats_data);
    } catch (err) {
      // Don't set error state for stats failures, just log
      console.error('Failed to load comment stats:', err);
    }
  }, [letter_id]);

  // Create a new comment
  const create_comment = useCallback(async (data: CommentCreateRequest): Promise<Comment> => {
    clear_error();
    
    try {
      const new_comment = await commentApi.createComment(data);
      
      // Add comment to appropriate position in the list
      setComments(prev => {
        if (data.parent_id) {
          // For replies, find parent and add to replies
          return prev.map(comment => {
            if (comment.id === data.parent_id) {
              return {
                ...comment,
                reply_count: comment.reply_count + 1,
                replies: [...(comment.replies || []), new_comment],
              };
            }
            return comment;
          });
        } else {
          // For top-level comments, add to beginning
          return [new_comment, ...prev];
        }
      });

      // Update stats
      setStats(prev => prev ? { 
        ...prev, 
        comment_count: prev.comment_count + 1 
      } : null);

      return new_comment;
    } catch (err) {
      handle_error(err);
      throw err;
    }
  }, [clear_error, handle_error]);

  // Update an existing comment
  const update_comment = useCallback(async (id: string, data: CommentUpdateRequest): Promise<Comment> => {
    clear_error();

    try {
      const updated_comment = await commentApi.updateComment(id, data);
      
      // Update comment in the list
      setComments(prev => {
        const update_in_tree = (comment_list: Comment[]): Comment[] => {
          return comment_list.map(comment => {
            if (comment.id === id) {
              return { ...comment, ...updated_comment };
            }
            if (comment.replies && comment.replies.length > 0) {
              return {
                ...comment,
                replies: update_in_tree(comment.replies),
              };
            }
            return comment;
          });
        };

        return update_in_tree(prev);
      });

      return updated_comment;
    } catch (err) {
      handle_error(err);
      throw err;
    }
  }, [clear_error, handle_error]);

  // Delete a comment
  const delete_comment = useCallback(async (id: string): Promise<void> => {
    clear_error();

    try {
      await commentApi.deleteComment(id);
      
      // Remove comment from the list
      setComments(prev => {
        const remove_from_tree = (comment_list: Comment[]): Comment[] => {
          return comment_list.filter(comment => {
            if (comment.id === id) {
              return false;
            }
            if (comment.replies && comment.replies.length > 0) {
              comment.replies = remove_from_tree(comment.replies);
            }
            return true;
          });
        };

        return remove_from_tree(prev);
      });

      // Update stats
      setStats(prev => prev ? { 
        ...prev, 
        comment_count: Math.max(0, prev.comment_count - 1)
      } : null);
    } catch (err) {
      handle_error(err);
      throw err;
    }
  }, [clear_error, handle_error]);

  // Like or unlike a comment
  const like_comment = useCallback(async (id: string): Promise<void> => {
    clear_error();

    try {
      const result = await commentApi.likeComment(id);
      
      // Update comment like status and count
      setComments(prev => {
        const update_in_tree = (comment_list: Comment[]): Comment[] => {
          return comment_list.map(comment => {
            if (comment.id === id) {
              return {
                ...comment,
                is_liked: result.is_liked,
                like_count: result.like_count,
              };
            }
            if (comment.replies && comment.replies.length > 0) {
              return {
                ...comment,
                replies: update_in_tree(comment.replies),
              };
            }
            return comment;
          });
        };

        return update_in_tree(prev);
      });
    } catch (err) {
      handle_error(err);
      throw err;
    }
  }, [clear_error, handle_error]);

  // Load replies for a specific comment
  const load_replies = useCallback(async (parent_id: string, query: Partial<CommentListQuery> = {}): Promise<Comment[]> => {
    clear_error();

    try {
      const response = await commentApi.getCommentReplies(parent_id, {
        page: 1,
        limit: 10,
        sort_by: 'created_at',
        order: 'asc',
        ...query,
      });

      // Update parent comment with loaded replies
      setComments(prev => {
        const update_in_tree = (comment_list: Comment[]): Comment[] => {
          return comment_list.map(comment => {
            if (comment.id === parent_id) {
              return {
                ...comment,
                replies: response.comments,
                is_loading_replies: false,
              };
            }
            if (comment.replies && comment.replies.length > 0) {
              return {
                ...comment,
                replies: update_in_tree(comment.replies),
              };
            }
            return comment;
          });
        };

        return update_in_tree(prev);
      });

      return response.comments;
    } catch (err) {
      handle_error(err);
      throw err;
    }
  }, [clear_error, handle_error]);

  // Refresh all data
  const refresh = useCallback(async () => {
    await Promise.all([
      load_comments({ page: 1 }),
      load_stats(),
    ]);
  }, [load_comments, load_stats]);

  // Utility functions
  const get_comment_by_id = useCallback((id: string): Comment | undefined => {
    const find_in_tree = (comment_list: Comment[]): Comment | undefined => {
      for (const comment of comment_list) {
        if (comment.id === id) return comment;
        if (comment.replies && comment.replies.length > 0) {
          const found = find_in_tree(comment.replies);
          if (found) return found;
        }
      }
      return undefined;
    };

    return find_in_tree(comments);
  }, [comments]);

  const get_replies = useCallback((parent_id: string): Comment[] => {
    const parent = get_comment_by_id(parent_id);
    return parent?.replies || [];
  }, [get_comment_by_id]);

  // Auto-load on mount and letter_id change
  useEffect(() => {
    if (auto_load && letter_id) {
      refresh();
    }
  }, [auto_load, letter_id, refresh]);

  // Real-time updates (TODO: implement WebSocket integration)
  useEffect(() => {
    if (!enable_real_time || !letter_id) return;

    // WebSocket integration would go here
    // Listen for comment events and update state accordingly
    
    return () => {
      // Cleanup WebSocket connection
    };
  }, [enable_real_time, letter_id]);

  return {
    comments,
    stats,
    loading,
    error,
    has_more,
    load_comments,
    create_comment,
    update_comment,
    delete_comment,
    like_comment,
    load_replies,
    refresh,
    clear_error,
    get_comment_by_id,
    get_replies,
  };
}

// Simplified hook for basic comment display
export function useCommentList(letter_id: string, options: Partial<UseCommentsOptions> = {}) {
  return useComments({
    letter_id,
    auto_load: true,
    enable_real_time: false,
    ...options,
  });
}

// Hook for comment statistics only
export function useCommentStats(letter_id: string) {
  const [stats, setStats] = useState<CommentStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load_stats = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const stats_data = await commentApi.getCommentStats(letter_id);
      setStats(stats_data);
    } catch (err) {
      const error_message = err instanceof Error ? err.message : 'Failed to load stats';
      setError(error_message);
    } finally {
      setLoading(false);
    }
  }, [letter_id]);

  useEffect(() => {
    if (letter_id) {
      load_stats();
    }
  }, [letter_id, load_stats]);

  return { stats, loading, error, refresh: load_stats };
}

export default useComments;