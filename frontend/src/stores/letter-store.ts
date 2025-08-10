import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import type { 
  LetterDraft, 
  Letter, 
  LetterCode, 
  SentLetter, 
  ReceivedLetter,
  LetterStats,
  LetterStatus,
  LetterStyle 
} from '@/types/letter'
import { generateId, generateLetterCode } from '@/lib/utils'

interface LetterStore {
  // 状态
  currentDraft: LetterDraft | null
  savedDrafts: LetterDraft[]
  sentLetters: SentLetter[]
  receivedLetters: ReceivedLetter[]
  letterStats: LetterStats | null
  isLoading: boolean
  error: string | null

  // 草稿操作
  createDraft: (content: string, style?: LetterStyle) => LetterDraft
  saveDraft: (draft: LetterDraft) => void
  loadDraft: (id: string) => void
  deleteDraft: (id: string) => void
  clearCurrentDraft: () => void

  // 信件操作
  generateCode: (draftId: string) => Promise<LetterCode>
  sendLetter: (letter: Letter) => Promise<void>
  updateLetterStatus: (codeId: string, status: LetterStatus) => Promise<void>
  
  // 数据获取
  fetchSentLetters: () => Promise<void>
  fetchReceivedLetters: () => Promise<void>
  fetchLetterStats: () => Promise<void>
  
  // 搜索和筛选
  searchLetters: (query: string) => (SentLetter | ReceivedLetter)[]
  filterLettersByStatus: (status: LetterStatus) => SentLetter[]
  
  // 工具方法
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
}

export const useLetterStore = create<LetterStore>()(
  devtools(
    persist(
      immer((set, get) => ({
        // 初始状态
        currentDraft: null,
        savedDrafts: [],
        sentLetters: [],
        receivedLetters: [],
        letterStats: null,
        isLoading: false,
        error: null,

        // 草稿操作
        createDraft: (content: string, style: LetterStyle = 'classic') => {
          const draft: LetterDraft = {
            id: generateId(),
            content,
            style,
            createdAt: new Date(),
            updatedAt: new Date(),
          }
          
          set((state) => {
            state.currentDraft = draft
            // 自动保存到草稿列表
            const existingIndex = state.savedDrafts.findIndex(d => d.id === draft.id)
            if (existingIndex >= 0) {
              state.savedDrafts[existingIndex] = draft
            } else {
              state.savedDrafts.unshift(draft)
            }
          })
          
          return draft
        },

        saveDraft: (draft: LetterDraft) => {
          set((state) => {
            draft.updatedAt = new Date()
            
            const existingIndex = state.savedDrafts.findIndex(d => d.id === draft.id)
            if (existingIndex >= 0) {
              state.savedDrafts[existingIndex] = draft
            } else {
              state.savedDrafts.unshift(draft)
            }
            
            // 如果是当前草稿，也更新
            if (state.currentDraft?.id === draft.id) {
              state.currentDraft = draft
            }
          })
        },

        loadDraft: (id: string) => {
          set((state) => {
            const draft = state.savedDrafts.find(d => d.id === id)
            if (draft) {
              state.currentDraft = draft
            }
          })
        },

        deleteDraft: (id: string) => {
          set((state) => {
            state.savedDrafts = state.savedDrafts.filter(d => d.id !== id)
            if (state.currentDraft?.id === id) {
              state.currentDraft = null
            }
          })
        },

        clearCurrentDraft: () => {
          set((state) => {
            state.currentDraft = null
          })
        },

        // 信件操作
        generateCode: async (draftId: string) => {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000))
            
            const code: LetterCode = {
              id: generateId(),
              letter_id: draftId,
              code: generateLetterCode(),
              generated_at: new Date(),
            }

            // 这里应该调用实际的API
            // const response = await api.generateCode(draftId)
            
            set((state) => {
              state.isLoading = false
            })

            return code
          } catch (error) {
            set((state) => {
              state.isLoading = false
              state.error = error instanceof Error ? error.message : '生成编号失败'
            })
            throw error
          }
        },

        sendLetter: async (letter: Letter) => {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1500))
            
            // 这里应该调用实际的API
            // await api.sendLetter(letter)
            
            const sentLetter: SentLetter = {
              ...letter,
              statusLogs: [],
              photos: [],
            }

            set((state) => {
              state.sentLetters.unshift(sentLetter)
              // 从草稿中移除
              state.savedDrafts = state.savedDrafts.filter(d => d.id !== letter.id)
              if (state.currentDraft?.id === letter.id) {
                state.currentDraft = null
              }
              state.isLoading = false
            })
          } catch (error) {
            set((state) => {
              state.isLoading = false
              state.error = error instanceof Error ? error.message : '发送信件失败'
            })
            throw error
          }
        },

        updateLetterStatus: async (codeId: string, status: LetterStatus) => {
          try {
            // 这里应该调用实际的API
            // await api.updateLetterStatus(codeId, status)
            
            set((state) => {
              const letter = state.sentLetters.find(l => l.code?.id === codeId)
              if (letter) {
                letter.status = status
                letter.statusLogs.push({
                  id: generateId(),
                  codeId,
                  status,
                  updatedBy: 'system',
                  createdAt: new Date(),
                })
              }
            })
          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : '更新状态失败'
            })
            throw error
          }
        },

        // 数据获取
        fetchSentLetters: async () => {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000))
            
            // 这里应该调用实际的API
            // const letters = await api.getSentLetters()
            
            // 模拟数据
            const mockLetters: SentLetter[] = []
            
            set((state) => {
              state.sentLetters = mockLetters
              state.isLoading = false
            })
          } catch (error) {
            set((state) => {
              state.isLoading = false
              state.error = error instanceof Error ? error.message : '获取发送信件失败'
            })
          }
        },

        fetchReceivedLetters: async () => {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000))
            
            // 这里应该调用实际的API
            // const letters = await api.getReceivedLetters()
            
            // 模拟数据
            const mockLetters: ReceivedLetter[] = []
            
            set((state) => {
              state.receivedLetters = mockLetters
              state.isLoading = false
            })
          } catch (error) {
            set((state) => {
              state.isLoading = false
              state.error = error instanceof Error ? error.message : '获取收到信件失败'
            })
          }
        },

        fetchLetterStats: async () => {
          try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 500))
            
            const { sentLetters, receivedLetters, savedDrafts } = get()
            
            const stats: LetterStats = {
              totalSent: sentLetters.length,
              totalReceived: receivedLetters.length,
              inTransit: sentLetters.filter(l => l.status === 'in_transit').length,
              delivered: sentLetters.filter(l => l.status === 'delivered').length,
              drafts: savedDrafts.length,
            }
            
            set((state) => {
              state.letterStats = stats
            })
          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : '获取统计数据失败'
            })
          }
        },

        // 搜索和筛选
        searchLetters: (query: string) => {
          const { sentLetters, receivedLetters } = get()
          const allLetters = [...sentLetters, ...receivedLetters]
          
          if (!query.trim()) return allLetters
          
          const lowercaseQuery = query.toLowerCase()
          return allLetters.filter(letter => 
            letter.content.toLowerCase().includes(lowercaseQuery) ||
            letter.title?.toLowerCase().includes(lowercaseQuery)
          )
        },

        filterLettersByStatus: (status: LetterStatus) => {
          const { sentLetters } = get()
          return sentLetters.filter(letter => letter.status === status)
        },

        // 工具方法
        setLoading: (loading: boolean) => {
          set((state) => {
            state.isLoading = loading
          })
        },

        setError: (error: string | null) => {
          set((state) => {
            state.error = error
          })
        },

        clearError: () => {
          set((state) => {
            state.error = null
          })
        },
      })),
      {
        name: 'letter-store',
        partialize: (state) => ({
          savedDrafts: state.savedDrafts,
          currentDraft: state.currentDraft,
        }),
      }
    ),
    {
      name: 'letter-store',
    }
  )
)