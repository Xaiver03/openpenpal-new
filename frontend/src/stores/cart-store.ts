/**
 * Global Cart State Store - 全局购物车状态管理
 * Unified state management for shopping cart functionality
 */

import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

export interface CartItem {
  id: number
  name: string
  description: string
  price: number
  originalPrice: number
  quantity: number
  image: string
  category: string
  tags: string[]
  maxQuantity?: number
}

export interface CartStore {
  items: CartItem[]
  total: number
  itemCount: number
  
  // Actions
  addItem: (product: Omit<CartItem, 'quantity'>, quantity?: number) => void
  removeItem: (id: number) => void
  updateQuantity: (id: number, quantity: number) => void
  clearCart: () => void
  getItemQuantity: (id: number) => number
  
  // Computed values
  calculateTotal: () => number
  calculateItemCount: () => number
}

export const useCartStore = create<CartStore>()(
  devtools(
    persist(
      (set, get) => ({
        items: [],
        total: 0,
        itemCount: 0,
        
        addItem: (product, quantity = 1) => {
          set((state) => {
            const existingItem = state.items.find(item => item.id === product.id)
            
            if (existingItem) {
              // Update quantity if item already exists
              const newItems = state.items.map(item =>
                item.id === product.id
                  ? { ...item, quantity: item.quantity + quantity }
                  : item
              )
              
              return {
                items: newItems,
                total: get().calculateTotal(),
                itemCount: get().calculateItemCount()
              }
            } else {
              // Add new item
              const newItem: CartItem = {
                ...product,
                quantity
              }
              
              const newItems = [...state.items, newItem]
              
              return {
                items: newItems,
                total: get().calculateTotal(),
                itemCount: get().calculateItemCount()
              }
            }
          })
        },
        
        removeItem: (id) => {
          set((state) => {
            const newItems = state.items.filter(item => item.id !== id)
            
            return {
              items: newItems,
              total: get().calculateTotal(),
              itemCount: get().calculateItemCount()
            }
          })
        },
        
        updateQuantity: (id, quantity) => {
          set((state) => {
            if (quantity <= 0) {
              // Remove item if quantity is 0 or less
              return {
                items: state.items.filter(item => item.id !== id),
                total: get().calculateTotal(),
                itemCount: get().calculateItemCount()
              }
            }
            
            const newItems = state.items.map(item =>
              item.id === id
                ? { ...item, quantity: Math.min(quantity, item.maxQuantity || 999) }
                : item
            )
            
            return {
              items: newItems,
              total: get().calculateTotal(),
              itemCount: get().calculateItemCount()
            }
          })
        },
        
        clearCart: () => {
          set({
            items: [],
            total: 0,
            itemCount: 0
          })
        },
        
        getItemQuantity: (id) => {
          const item = get().items.find(item => item.id === id)
          return item?.quantity || 0
        },
        
        calculateTotal: () => {
          const items = get().items
          return items.reduce((total, item) => total + (item.price * item.quantity), 0)
        },
        
        calculateItemCount: () => {
          const items = get().items
          return items.reduce((count, item) => count + item.quantity, 0)
        }
      }),
      {
        name: 'cart-storage',
        partialize: (state) => ({ 
          items: state.items 
        })
      }
    )
  )
)

// Helper hooks
export const useCartTotal = () => useCartStore((state) => state.total)
export const useCartItemCount = () => useCartStore((state) => state.itemCount)
export const useCartItems = () => useCartStore((state) => state.items)