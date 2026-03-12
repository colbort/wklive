/**
 * 分页 Hook
 */

import { ref, reactive, computed } from 'vue'

export interface PaginationState {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export function usePagination(initialPageSize = 10) {
  const pagination = reactive<PaginationState>({
    page: 1,
    pageSize: initialPageSize,
    total: 0,
    totalPages: 0,
  })

  const isLastPage = computed(() => pagination.page >= pagination.totalPages)

  const updateTotal = (total: number) => {
    pagination.total = total
    pagination.totalPages = Math.ceil(total / pagination.pageSize)
  }

  const reset = () => {
    pagination.page = 1
    pagination.total = 0
    pagination.totalPages = 0
  }

  const nextPage = () => {
    if (!isLastPage.value) {
      pagination.page++
    }
  }

  const prevPage = () => {
    if (pagination.page > 1) {
      pagination.page--
    }
  }

  const goPage = (page: number) => {
    if (page > 0 && page <= pagination.totalPages) {
      pagination.page = page
    }
  }

  return {
    pagination,
    isLastPage,
    updateTotal,
    reset,
    nextPage,
    prevPage,
    goPage,
  }
}
