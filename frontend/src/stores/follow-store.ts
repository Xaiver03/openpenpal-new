/**
 * Follow System Store - SOTA State Management
 * 关注系统状态管理 - 集成Zustand的优化状态管理，支持实时更新、缓存、乐观更新
 */

import { create } from 'zustand'
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware'
import { followApi } from '@/lib/api/follow'
import type {
  FollowUser,
  FollowSuggestion,
  FollowActivity,
  FollowNotificationSettings,
  FollowListQuery,
  UserSuggestionsQuery,
  UserSearchQuery,
  FollowState,
  FollowActions,
  FollowError,
} from '@/types/follow'

// ================================
// Store State Interface
// ================================

interface FollowStoreState extends FollowState, FollowActions {
  // Internal state management
  _isInitialized: boolean
  _lastActivity: number
}

// ================================
// Initial State
// ================================

const createInitialState = (): FollowState => ({
  // Current user's follow data
  followers: [],
  following: [],
  follower_count: 0,
  following_count: 0,
  
  // Cached data
  suggestions: [],
  recent_activities: [],
  
  // UI state
  loading: {
    followers: false,
    following: false,
    suggestions: false,
    follow_action: false,
  },
  
  // Last updated timestamps
  last_updated: {
    followers: null,
    following: null,
    suggestions: null,
  },
  
  // Error states
  errors: {
    followers: null,
    following: null,
    suggestions: null,
    follow_action: null,
  },
  
  // Settings
  notification_settings: {
    new_followers: true,
    follow_backs: true,
    suggestions: false,
    activity_digest: false,
    email_notifications: false,
  },
})

// ================================
// SOTA Helper Functions
// ================================

// Cache management
const CACHE_DURATION = 5 * 60 * 1000 // 5 minutes
const isCacheValid = (timestamp: number | null): boolean => {
  return timestamp !== null && Date.now() - timestamp < CACHE_DURATION
}

// Optimistic update helpers
const createOptimisticFollowUser = (user: FollowUser): FollowUser => ({
  ...user,
  followers_count: (user.followers_count || 0) + 1,
  is_following: true,
  followed_at: new Date().toISOString(),
})

const createOptimisticUnfollowUser = (user: FollowUser): FollowUser => ({
  ...user,
  followers_count: Math.max((user.followers_count || 0) - 1, 0),
  is_following: false,
  followed_at: undefined,
})

// Error handling
const createFollowStoreError = (error: any, action: string): string => {
  console.error(`Follow store error in ${action}:`, error)
  
  if (error?.message) return error.message
  if (error?.response?.data?.message) return error.response.data.message
  
  return `Failed to ${action}`
}

// ================================
// Main Follow Store
// ================================

export const useFollowStore = create<FollowStoreState>()(
  devtools(
    persist(
      subscribeWithSelector((set, get) => ({
        // Initial state
        ...createInitialState(),
        _isInitialized: false,
        _lastActivity: Date.now(),
        
        // ================================
        // Core Follow Actions
        // ================================
        
        followUser: async (userId: string, options = {}) => {
          const state = get()
          
          // Set loading state
          set(
            (state) => ({
              loading: { ...state.loading, follow_action: true },
              errors: { ...state.errors, follow_action: null },
            }),
            false,
            'followUser:start'
          )
          
          // Optimistic update for following list
          const optimisticUpdate = () => {
            set(
              (state) => ({
                following: state.following.map(user => 
                  user.id === userId ? createOptimisticFollowUser(user) : user
                ),
                following_count: state.following_count + 1,
                _lastActivity: Date.now(),
              }),
              false,
              'followUser:optimistic'
            )
          }
          
          // Rollback function
          const rollback = () => {
            set(
              (state) => ({
                following: state.following.map(user =>
                  user.id === userId ? createOptimisticUnfollowUser(user) : user
                ),
                following_count: Math.max(state.following_count - 1, 0),
              }),
              false,
              'followUser:rollback'
            )
          }
          
          try {
            // Apply optimistic update
            optimisticUpdate()
            
            // Perform API call
            const response = await followApi.followUser(userId, options)
            
            if (response.success) {
              // Update with real data
              set(
                (state) => ({
                  following_count: response.following_count,
                  loading: { ...state.loading, follow_action: false },
                }),
                false,
                'followUser:success'
              )
              
              // Invalidate cache to force refresh on next load
              set(
                (state) => ({
                  last_updated: { ...state.last_updated, following: null },
                }),
                false,
                'followUser:invalidateCache'
              )
              
              return true
            } else {
              rollback()
              throw new Error(response.message || 'Follow action failed')
            }
          } catch (error) {
            rollback()
            const errorMessage = createFollowStoreError(error, 'follow user')
            
            set(
              (state) => ({
                loading: { ...state.loading, follow_action: false },
                errors: { ...state.errors, follow_action: errorMessage },
              }),
              false,
              'followUser:error'
            )
            
            throw error
          }
        },
        
        unfollowUser: async (userId: string) => {
          const state = get()
          
          // Set loading state
          set(
            (state) => ({
              loading: { ...state.loading, follow_action: true },
              errors: { ...state.errors, follow_action: null },
            }),
            false,
            'unfollowUser:start'
          )
          
          // Optimistic update
          const optimisticUpdate = () => {
            set(
              (state) => ({
                following: state.following.map(user =>
                  user.id === userId ? createOptimisticUnfollowUser(user) : user
                ),
                following_count: Math.max(state.following_count - 1, 0),
                _lastActivity: Date.now(),
              }),
              false,
              'unfollowUser:optimistic'
            )
          }
          
          // Rollback function
          const rollback = () => {
            set(
              (state) => ({
                following: state.following.map(user =>
                  user.id === userId ? createOptimisticFollowUser(user) : user
                ),
                following_count: state.following_count + 1,
              }),
              false,
              'unfollowUser:rollback'
            )
          }
          
          try {
            // Apply optimistic update
            optimisticUpdate()
            
            // Perform API call
            const response = await followApi.unfollowUser(userId)
            
            if (response.success) {
              // Update with real data
              set(
                (state) => ({
                  following_count: response.following_count,
                  loading: { ...state.loading, follow_action: false },
                }),
                false,
                'unfollowUser:success'
              )
              
              // Invalidate cache
              set(
                (state) => ({
                  last_updated: { ...state.last_updated, following: null },
                }),
                false,
                'unfollowUser:invalidateCache'
              )
              
              return true
            } else {
              rollback()
              throw new Error(response.message || 'Unfollow action failed')
            }
          } catch (error) {
            rollback()
            const errorMessage = createFollowStoreError(error, 'unfollow user')
            
            set(
              (state) => ({
                loading: { ...state.loading, follow_action: false },
                errors: { ...state.errors, follow_action: errorMessage },
              }),
              false,
              'unfollowUser:error'
            )
            
            throw error
          }
        },
        
        // ================================
        // Data Loading Methods
        // ================================
        
        loadFollowers: async (userId?: string, query: FollowListQuery = {}) => {
          const state = get()
          
          // Check cache validity
          if (isCacheValid(state.last_updated.followers) && !query.page && !query.search) {
            return // Use cached data
          }
          
          set(
            (state) => ({
              loading: { ...state.loading, followers: true },
              errors: { ...state.errors, followers: null },
            }),
            false,
            'loadFollowers:start'
          )
          
          try {
            const response = await followApi.getFollowers(userId, query)
            
            set(
              (state) => ({
                followers: query.page && query.page > 1 
                  ? [...state.followers, ...response.users]
                  : response.users,
                follower_count: response.pagination.total,
                loading: { ...state.loading, followers: false },
                last_updated: { ...state.last_updated, followers: Date.now() },
              }),
              false,
              'loadFollowers:success'
            )
          } catch (error) {
            const errorMessage = createFollowStoreError(error, 'load followers')
            
            set(
              (state) => ({
                loading: { ...state.loading, followers: false },
                errors: { ...state.errors, followers: errorMessage },
              }),
              false,
              'loadFollowers:error'
            )
          }
        },
        
        loadFollowing: async (userId?: string, query: FollowListQuery = {}) => {
          const state = get()
          
          // Check cache validity
          if (isCacheValid(state.last_updated.following) && !query.page && !query.search) {
            return // Use cached data
          }
          
          set(
            (state) => ({
              loading: { ...state.loading, following: true },
              errors: { ...state.errors, following: null },
            }),
            false,
            'loadFollowing:start'
          )
          
          try {
            const response = await followApi.getFollowing(userId, query)
            
            set(
              (state) => ({
                following: query.page && query.page > 1
                  ? [...state.following, ...response.users]
                  : response.users,
                following_count: response.pagination.total,
                loading: { ...state.loading, following: false },
                last_updated: { ...state.last_updated, following: Date.now() },
              }),
              false,
              'loadFollowing:success'
            )
          } catch (error) {
            const errorMessage = createFollowStoreError(error, 'load following')
            
            set(
              (state) => ({
                loading: { ...state.loading, following: false },
                errors: { ...state.errors, following: errorMessage },
              }),
              false,
              'loadFollowing:error'
            )
          }
        },
        
        loadSuggestions: async (query: UserSuggestionsQuery = {}) => {
          const state = get()
          
          // Check cache validity
          if (isCacheValid(state.last_updated.suggestions) && state.suggestions.length > 0) {
            return // Use cached data
          }
          
          set(
            (state) => ({
              loading: { ...state.loading, suggestions: true },
              errors: { ...state.errors, suggestions: null },
            }),
            false,
            'loadSuggestions:start'
          )
          
          try {
            const response = await followApi.getUserSuggestions(query)
            
            set(
              (state) => ({
                suggestions: response.suggestions,
                loading: { ...state.loading, suggestions: false },
                last_updated: { ...state.last_updated, suggestions: Date.now() },
              }),
              false,
              'loadSuggestions:success'
            )
          } catch (error) {
            const errorMessage = createFollowStoreError(error, 'load suggestions')
            
            set(
              (state) => ({
                loading: { ...state.loading, suggestions: false },
                errors: { ...state.errors, suggestions: errorMessage },
              }),
              false,
              'loadSuggestions:error'
            )
          }
        },
        
        refreshSuggestions: async () => {
          set(
            (state) => ({
              loading: { ...state.loading, suggestions: true },
              errors: { ...state.errors, suggestions: null },
            }),
            false,
            'refreshSuggestions:start'
          )
          
          try {
            const response = await followApi.refreshSuggestions()
            
            set(
              (state) => ({
                suggestions: response.suggestions,
                loading: { ...state.loading, suggestions: false },
                last_updated: { ...state.last_updated, suggestions: Date.now() },
              }),
              false,
              'refreshSuggestions:success'
            )
          } catch (error) {
            const errorMessage = createFollowStoreError(error, 'refresh suggestions')
            
            set(
              (state) => ({
                loading: { ...state.loading, suggestions: false },
                errors: { ...state.errors, suggestions: errorMessage },
              }),
              false,
              'refreshSuggestions:error'
            )
          }
        },
        
        // ================================
        // Bulk Operations
        // ================================
        
        followMultipleUsers: async (userIds: string[]) => {
          try {
            const response = await followApi.followMultipleUsers(userIds)
            
            // Update following count optimistically
            if (response.success.length > 0) {
              set(
                (state) => ({
                  following_count: state.following_count + response.success.length,
                  last_updated: { ...state.last_updated, following: null }, // Invalidate cache
                }),
                false,
                'followMultipleUsers:success'
              )
            }
            
            return response
          } catch (error) {
            throw createFollowStoreError(error, 'follow multiple users')
          }
        },
        
        removeFollower: async (userId: string) => {
          try {
            const success = await followApi.removeFollower(userId)
            
            if (success) {
              set(
                (state) => ({
                  followers: state.followers.filter(user => user.id !== userId),
                  follower_count: Math.max(state.follower_count - 1, 0),
                }),
                false,
                'removeFollower:success'
              )
            }
            
            return success
          } catch (error) {
            throw createFollowStoreError(error, 'remove follower')
          }
        },
        
        // ================================
        // User Data Methods
        // ================================
        
        getUserFollowStatus: async (userId: string) => {
          try {
            return await followApi.getFollowStatus(userId)
          } catch (error) {
            throw createFollowStoreError(error, 'get follow status')
          }
        },
        
        searchUsers: async (query: string, filters: Partial<FollowListQuery> = {}) => {
          try {
            const searchQuery: UserSearchQuery = {
              query,
              ...filters,
            }
            
            const response = await followApi.searchUsers(searchQuery)
            return response.users
          } catch (error) {
            throw createFollowStoreError(error, 'search users')
          }
        },
        
        // ================================
        // Settings Management
        // ================================
        
        updateNotificationSettings: async (settings: Partial<FollowNotificationSettings>) => {
          const currentSettings = get().notification_settings
          
          // Optimistic update
          set(
            (state) => ({
              notification_settings: { ...state.notification_settings, ...settings },
            }),
            false,
            'updateNotificationSettings:optimistic'
          )
          
          try {
            const updatedSettings = await followApi.updateNotificationSettings(settings)
            
            set(
              (state) => ({
                notification_settings: updatedSettings,
              }),
              false,
              'updateNotificationSettings:success'
            )
          } catch (error) {
            // Rollback
            set(
              (state) => ({
                notification_settings: currentSettings,
              }),
              false,
              'updateNotificationSettings:rollback'
            )
            
            throw createFollowStoreError(error, 'update notification settings')
          }
        },
        
        // ================================
        // Cache Management
        // ================================
        
        clearCache: () => {
          set(
            {
              ...createInitialState(),
              _isInitialized: get()._isInitialized,
              _lastActivity: Date.now(),
            },
            false,
            'clearCache'
          )
        },
        
        invalidateUser: (userId: string) => {
          set(
            (state) => ({
              followers: state.followers.filter(user => user.id !== userId),
              following: state.following.filter(user => user.id !== userId),
              suggestions: state.suggestions.filter(suggestion => suggestion.user.id !== userId),
              last_updated: {
                ...state.last_updated,
                followers: null,
                following: null,
                suggestions: null,
              },
            }),
            false,
            'invalidateUser'
          )
        },
      })),
      {
        name: 'openpenpal-follow-store',
        partialize: (state) => ({
          notification_settings: state.notification_settings,
          _isInitialized: state._isInitialized,
        }),
      }
    ),
    {
      name: 'follow-store',
    }
  )
)

// ================================
// Convenience Hooks
// ================================

/**
 * Hook for accessing follow state and basic actions
 */
export const useFollow = () => {
  const store = useFollowStore()
  
  return {
    // State
    followers: store.followers,
    following: store.following,
    suggestions: store.suggestions,
    followerCount: store.follower_count,
    followingCount: store.following_count,
    
    // Loading states
    isLoading: store.loading,
    errors: store.errors,
    
    // Actions
    followUser: store.followUser,
    unfollowUser: store.unfollowUser,
    followMultipleUsers: store.followMultipleUsers,
    loadFollowers: store.loadFollowers,
    loadFollowing: store.loadFollowing,
    loadSuggestions: store.loadSuggestions,
    refreshSuggestions: store.refreshSuggestions,
  }
}

/**
 * Hook for user follow status checks
 */
export const useFollowStatus = (userId: string) => {
  const following = useFollowStore(state => state.following)
  const followers = useFollowStore(state => state.followers)
  
  const isFollowing = following.some(user => user.id === userId)
  const isFollower = followers.some(user => user.id === userId)
  const isMutual = isFollowing && isFollower
  
  return {
    isFollowing,
    isFollower,
    isMutual,
  }
}

/**
 * Hook for optimized follow actions
 */
export const useFollowActions = () => {
  const store = useFollowStore()
  
  return {
    followUser: store.followUser,
    unfollowUser: store.unfollowUser,
    followMultipleUsers: store.followMultipleUsers,
    removeFollower: store.removeFollower,
    searchUsers: store.searchUsers,
    getUserFollowStatus: store.getUserFollowStatus,
  }
}

export default useFollowStore