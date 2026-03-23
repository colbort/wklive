/**
 * 分页 Hook
 */

import { ref, reactive, computed } from 'vue'

export interface PaginationState {
  cursor: string | null
  nextCursor: string | null
  prevCursor: string | null
  limit: number
  total: number
  hasNext: boolean
  hasPrev: boolean
}

export function usePagination(initialLimit = 10) {
  const pagination = reactive<PaginationState>({
    cursor: null,
    nextCursor: null,
    prevCursor: null,
    limit: initialLimit,
    total: 0,
    hasNext: false,
    hasPrev: false,
  })

  const updatePagination = (
    total: number,
    hasNext: boolean,
    hasPrev: boolean,
    nextCursor: string | null = null,
    prevCursor: string | null = null,
  ) => {
    pagination.total = total
    pagination.hasNext = hasNext
    pagination.hasPrev = hasPrev
    pagination.nextCursor = nextCursor
    pagination.prevCursor = prevCursor
  }

  const reset = () => {
    pagination.cursor = null
    pagination.nextCursor = null
    pagination.prevCursor = null
    pagination.total = 0
    pagination.hasNext = false
    pagination.hasPrev = false
  }

  const nextPage = () => {
    if (pagination.hasNext && pagination.nextCursor) {
      pagination.cursor = pagination.nextCursor
    }
  }

  const prevPage = () => {
    if (pagination.hasPrev && pagination.prevCursor) {
      pagination.cursor = pagination.prevCursor
    }
  }

  return {
    pagination,
    updatePagination,
    reset,
    nextPage,
    prevPage,
  }
}

