/**
 * 分页 Hook
 */

import { reactive } from 'vue'

export interface PaginationState<TCursor = number> {
  cursor: TCursor | undefined
  nextCursor: TCursor | undefined
  prevCursor: TCursor | undefined
  limit: number
  total: number
  hasNext: boolean
  hasPrev: boolean
}

export function usePagination<TCursor = number>(initialLimit = 10) {
  const pagination = reactive({
    cursor: undefined,
    nextCursor: undefined,
    prevCursor: undefined,
    limit: initialLimit,
    total: 0,
    hasNext: false,
    hasPrev: false,
  }) as PaginationState<TCursor>

  const updatePagination = (
    total: number,
    hasNext: boolean,
    hasPrev: boolean,
    nextCursor?: TCursor,
    prevCursor?: TCursor,
  ) => {
    pagination.total = total
    pagination.hasNext = hasNext
    pagination.hasPrev = hasPrev
    pagination.nextCursor = nextCursor ?? undefined
    pagination.prevCursor = prevCursor ?? undefined
  }

  const reset = () => {
    pagination.cursor = undefined
    pagination.nextCursor = undefined
    pagination.prevCursor = undefined
    pagination.total = 0
    pagination.hasNext = false
    pagination.hasPrev = false
  }

  const nextPage = () => {
    if (pagination.hasNext && pagination.nextCursor !== undefined) {
      pagination.cursor = pagination.nextCursor
    }
  }

  const prevPage = () => {
    if (pagination.hasPrev && pagination.prevCursor !== undefined) {
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
