/**
 * Museum React Hooks
 * 信件博物馆 React Hooks
 */

import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { museumService, MuseumEntry, MuseumExhibition, GetMuseumEntriesParams } from '@/lib/services/museum-service'
import { normalizeResponse } from '@/lib/api/response'
import { toast } from '@/components/ui/use-toast'

// Query Keys
const MUSEUM_QUERY_KEYS = {
  all: ['museum'] as const,
  entries: (params?: GetMuseumEntriesParams) => ['museum', 'entries', params] as const,
  entry: (id: string) => ['museum', 'entry', id] as const,
  exhibitions: (isActive?: boolean) => ['museum', 'exhibitions', isActive] as const,
  exhibition: (id: string) => ['museum', 'exhibition', id] as const,
  stats: ['museum', 'stats'] as const,
  featured: (limit: number) => ['museum', 'featured', limit] as const,
}

/**
 * Hook to fetch museum entries
 */
export function useMuseumEntries(params?: GetMuseumEntriesParams) {
  return useQuery({
    queryKey: MUSEUM_QUERY_KEYS.entries(params),
    queryFn: async () => {
      const rawResponse = await museumService.getEntries(params)
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook to fetch a single museum entry
 */
export function useMuseumEntry(id: string) {
  const queryClient = useQueryClient()

  // Increment view count on mount
  useEffect(() => {
    if (id) {
      museumService.incrementViewCount(id).catch(console.error)
    }
  }, [id])

  return useQuery({
    queryKey: MUSEUM_QUERY_KEYS.entry(id),
    queryFn: async () => {
      const rawResponse = await museumService.getEntry(id)
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    enabled: !!id,
  })
}

/**
 * Hook to fetch museum exhibitions
 */
export function useMuseumExhibitions(isActive?: boolean) {
  return useQuery({
    queryKey: MUSEUM_QUERY_KEYS.exhibitions(isActive),
    queryFn: async () => {
      const rawResponse = await museumService.getExhibitions(isActive)
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    staleTime: 10 * 60 * 1000, // 10 minutes
  })
}

/**
 * Hook to fetch featured entries
 */
export function useFeaturedEntries(limit: number = 6) {
  return useQuery({
    queryKey: MUSEUM_QUERY_KEYS.featured(limit),
    queryFn: async () => {
      const rawResponse = await museumService.getFeaturedEntries(limit)
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook to submit a letter to museum
 */
export function useSubmitToMuseum() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: museumService.submitToMuseum,
    onSuccess: (rawResponse) => {
      const response = normalizeResponse(rawResponse)
      if (response.code === 0) {
        toast({
          title: "提交成功",
          description: "您的信件已提交至博物馆，等待审核。",
        })
        // Invalidate related queries
        queryClient.invalidateQueries({ queryKey: MUSEUM_QUERY_KEYS.all })
      } else {
        throw new Error(response.message)
      }
    },
    onError: (error: Error) => {
      toast({
        title: "提交失败",
        description: error.message,
        variant: "destructive",
      })
    },
  })
}

/**
 * Hook to like/unlike museum entries
 */
export function useMuseumLike(entryId: string) {
  const queryClient = useQueryClient()
  const [isLiked, setIsLiked] = useState(false)
  const [likeCount, setLikeCount] = useState(0)

  const likeMutation = useMutation({
    mutationFn: () => museumService.likeEntry(entryId),
    onSuccess: (response) => {
      if (response.code === 0 && response.data) {
        setIsLiked(true)
        setLikeCount(response.data.likeCount)
        // Update cache
        queryClient.setQueryData(
          MUSEUM_QUERY_KEYS.entry(entryId),
          (old: any) => ({
            ...old,
            data: {
              ...old?.data,
              likeCount: response.data?.likeCount ?? (old?.data?.likeCount ?? 0),
              is_liked: true,
            },
          })
        )
      }
    },
  })

  const unlikeMutation = useMutation({
    mutationFn: () => museumService.unlikeEntry(entryId),
    onSuccess: (response) => {
      if (response.code === 0 && response.data) {
        setIsLiked(false)
        setLikeCount(response.data.likeCount)
        // Update cache
        queryClient.setQueryData(
          MUSEUM_QUERY_KEYS.entry(entryId),
          (old: any) => ({
            ...old,
            data: {
              ...old?.data,
              likeCount: response.data?.likeCount ?? (old?.data?.likeCount ?? 0),
              is_liked: false,
            },
          })
        )
      }
    },
  })

  const toggleLike = useCallback(() => {
    if (isLiked) {
      unlikeMutation.mutate()
    } else {
      likeMutation.mutate()
    }
  }, [isLiked, likeMutation, unlikeMutation])

  return {
    isLiked,
    likeCount,
    toggleLike,
    isLoading: likeMutation.isPending || unlikeMutation.isPending,
  }
}

/**
 * Hook for museum statistics
 */
export function useMuseumStats() {
  return useQuery({
    queryKey: MUSEUM_QUERY_KEYS.stats,
    queryFn: async () => {
      const rawResponse = await museumService.getStats()
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    staleTime: 30 * 60 * 1000, // 30 minutes
  })
}

/**
 * Hook for searching museum entries
 */
export function useMuseumSearch() {
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState<{
    theme?: string
    tags?: string[]
    date_from?: string
    date_to?: string
  }>({})

  const searchQuery = useQuery({
    queryKey: ['museum', 'search', query, filters],
    queryFn: async () => {
      if (!query.trim()) return []
      const rawResponse = await museumService.searchEntries(query, filters)
      const response = normalizeResponse(rawResponse)
      if (response.code !== 0) {
        throw new Error(response.message)
      }
      return response.data
    },
    enabled: !!query.trim(),
    staleTime: 2 * 60 * 1000, // 2 minutes
  })

  return {
    query,
    setQuery,
    filters,
    setFilters,
    results: searchQuery.data || [],
    isLoading: searchQuery.isLoading,
    error: searchQuery.error,
  }
}

/**
 * Admin hook for museum moderation
 */
export function useMuseumModeration() {
  const queryClient = useQueryClient()

  const approveMutation = useMutation({
    mutationFn: museumService.approveEntry,
    onSuccess: () => {
      toast({
        title: "审核通过",
        description: "该信件已通过审核并发布到博物馆。",
      })
      queryClient.invalidateQueries({ queryKey: MUSEUM_QUERY_KEYS.all })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => 
      museumService.rejectEntry(id, reason),
    onSuccess: () => {
      toast({
        title: "已拒绝",
        description: "该信件已被拒绝。",
      })
      queryClient.invalidateQueries({ queryKey: MUSEUM_QUERY_KEYS.all })
    },
  })

  const toggleFeatureMutation = useMutation({
    mutationFn: ({ id, featured }: { id: string; featured: boolean }) =>
      museumService.toggleFeature(id, featured),
    onSuccess: (_, variables) => {
      toast({
        title: variables.featured ? "已设为精选" : "已取消精选",
        description: variables.featured 
          ? "该信件已设置为精选展示。" 
          : "该信件已从精选中移除。",
      })
      queryClient.invalidateQueries({ queryKey: MUSEUM_QUERY_KEYS.all })
    },
  })

  return {
    approve: approveMutation.mutate,
    reject: rejectMutation.mutate,
    toggleFeature: toggleFeatureMutation.mutate,
    isApproving: approveMutation.isPending,
    isRejecting: rejectMutation.isPending,
    isTogglingFeature: toggleFeatureMutation.isPending,
  }
}