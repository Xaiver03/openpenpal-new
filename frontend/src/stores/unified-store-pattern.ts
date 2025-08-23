import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

/**
 * 统一的Store创建模式
 * 遵循CLAUDE.md原则：SOTA principles，持续优化
 */

export interface StoreActions {
  reset: () => void
  clearError: () => void
}

export interface BaseState {
  loading: boolean
  error: Error | null
}

/**
 * 创建标准化的Zustand store
 * @param name Store名称
 * @param initialState 初始状态
 * @param actions Store actions
 * @param options 配置选项
 */
export function createStore<T extends BaseState>(
  name: string,
  initialState: T,
  actions: (set: any, get: any) => any,
  options?: {
    persist?: boolean
    immer?: boolean
  }
) {
  const { persist: usePersist = true, immer: useImmer = true } = options || {}

  const baseStore = (set: any, get: any) => ({
    ...initialState,
    ...actions(set, get),
    // 标准化的actions
    reset: () => set(initialState),
    clearError: () => set({ error: null }),
  })

  // 定义store，应用中间件
  let store: any = baseStore

  if (useImmer) {
    store = immer(store)
  }

  if (usePersist) {
    store = persist(store, {
      name: `openpenpal-${name}`,
      partialize: (state: any) => {
        // 不持久化loading和error状态
        const { loading, error, ...rest } = state
        return rest
      },
    })
  }

  if (process.env.NODE_ENV === 'development') {
    store = devtools(store, { name })
  }

  return create(store)
}

/**
 * 创建异步action的标准模式
 */
export function createAsyncAction<TArgs extends any[], TResult>(
  action: (...args: TArgs) => Promise<TResult>
) {
  return (set: any, get: any) => async (...args: TArgs) => {
    set({ loading: true, error: null })
    try {
      const result = await action(...args)
      set({ loading: false })
      return result
    } catch (error) {
      set({ loading: false, error })
      throw error
    }
  }
}

/**
 * 创建带乐观更新的action
 */
export function createOptimisticAction<TArgs extends any[], TResult>(
  optimisticUpdate: (state: any, ...args: TArgs) => void,
  action: (...args: TArgs) => Promise<TResult>,
  rollback?: (state: any, error: Error, ...args: TArgs) => void
) {
  return (set: any, get: any) => async (...args: TArgs) => {
    const previousState = get()
    
    // 乐观更新
    set((state: any) => {
      optimisticUpdate(state, ...args)
    })
    
    try {
      const result = await action(...args)
      return result
    } catch (error) {
      // 回滚
      if (rollback) {
        set((state: any) => {
          rollback(state, error as Error, ...args)
        })
      } else {
        // 默认回滚到之前的状态
        set(previousState)
      }
      throw error
    }
  }
}

/**
 * 使用示例：
 * 
 * interface UserState extends BaseState {
 *   users: User[]
 *   currentUser: User | null
 * }
 * 
 * const useUserStore = createStore<UserState>(
 *   'user',
 *   {
 *     loading: false,
 *     error: null,
 *     users: [],
 *     currentUser: null,
 *   },
 *   (set, get) => ({
 *     // 普通action
 *     setCurrentUser: (user: User) => set({ currentUser: user }),
 *     
 *     // 异步action
 *     fetchUsers: createAsyncAction(async () => {
 *       const response = await api.getUsers()
 *       set({ users: response.data })
 *       return response.data
 *     })(set, get),
 *     
 *     // 乐观更新action
 *     updateUser: createOptimisticAction(
 *       (state, userId: string, updates: Partial<User>) => {
 *         const index = state.users.findIndex(u => u.id === userId)
 *         if (index !== -1) {
 *           state.users[index] = { ...state.users[index], ...updates }
 *         }
 *       },
 *       async (userId: string, updates: Partial<User>) => {
 *         const response = await api.updateUser(userId, updates)
 *         return response.data
 *       }
 *     )(set, get),
 *   })
 * )
 */