import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  apiGetKline,
  apiGetQuote,
  apiListVisibleCategories,
  apiListVisibleProducts,
} from '@/api/market'
import type {
  GetKlineReq,
  GetQuoteReq,
  MarketTenantCategory,
  MarketTenantProduct,
  Kline,
  ListVisibleCategoriesReq,
  ListVisibleProductsReq,
  Quote,
} from '@/types/market'

export const useMarketStore = defineStore('market', () => {
  const categories = ref<MarketTenantCategory[]>([])
  const products = ref<MarketTenantProduct[]>([])
  const klines = ref<Kline[]>([])
  const currentQuote = ref<Quote | null>(null)
  const loading = ref(false)

  async function listVisibleCategories(params: ListVisibleCategoriesReq) {
    loading.value = true
    try {
      const res = await apiListVisibleCategories(params)
      categories.value = res.data || []
      return categories.value
    } finally {
      loading.value = false
    }
  }

  async function listVisibleProducts(params: ListVisibleProductsReq) {
    loading.value = true
    try {
      const res = await apiListVisibleProducts(params)
      products.value = res.data || []
      return products.value
    } finally {
      loading.value = false
    }
  }

  async function getKline(params: GetKlineReq) {
    loading.value = true
    try {
      const res = await apiGetKline(params)
      klines.value = res.data || []
      return klines.value
    } finally {
      loading.value = false
    }
  }

  async function getQuote(params: GetQuoteReq) {
    loading.value = true
    try {
      const res = await apiGetQuote(params)
      currentQuote.value = res.data || null
      return currentQuote.value
    } finally {
      loading.value = false
    }
  }

  return {
    categories,
    products,
    klines,
    currentQuote,
    loading,
    listVisibleCategories,
    listVisibleProducts,
    getKline,
    getQuote,
  }
})
