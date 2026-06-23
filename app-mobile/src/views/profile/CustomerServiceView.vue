<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiCreateChatToken } from '@/api/chat'
import CommonPage from '@/components/common/CommonPage.vue'
import { useI18n } from '@/i18n'

const router = useRouter()
const { t } = useI18n()
const chatToken = ref('')
const chatUiUrl = ref('')
const chatWsUrl = ref('')
const loading = ref(false)
const loadError = ref('')

const chatFrameUrl = computed(() => {
  const baseUrl = chatUiUrl.value.trim()
  const wsUrl = chatWsUrl.value.trim()

  if (!baseUrl || !chatToken.value) return ''

  const url = new URL(baseUrl, window.location.origin)
  if (url.pathname.replace(/\/$/, '').endsWith('/chat')) {
    url.pathname = url.pathname.replace(/\/chat\/?$/, '') || '/'
  }
  url.searchParams.set('page', 'chat')
  url.searchParams.set('mode', 'mobile')
  url.searchParams.set('chatToken', chatToken.value)
  if (wsUrl) {
    url.searchParams.set('wsUrl', wsUrl)
  }

  return url.toString()
})

onMounted(() => {
  void loadChatToken()
})

async function loadChatToken() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await apiCreateChatToken()
    if (res.code === 200) {
      chatToken.value = res.data.chatToken
      chatUiUrl.value = res.data.chatUiUrl
      chatWsUrl.value = res.data.chatWsUrl || ''
    } else {
      loadError.value = res.msg || '客服配置未完成'
    }
  } catch (error) {
    console.warn('load chat token failed', error)
    loadError.value = '客服配置未完成'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <CommonPage
    :title="t('userMenu.customerService')"
    :nav-height="58"
    @back="router.back()"
  >
    <section class="customer-service-page">
      <iframe
        v-if="chatFrameUrl"
        class="customer-service-page__frame"
        :src="chatFrameUrl"
        title="Customer support"
      />
      <div
        v-else
        class="customer-service-page__empty"
      >
        {{ loading ? '客服加载中' : loadError || '客服配置未完成' }}
      </div>
    </section>
  </CommonPage>
</template>

<style scoped>
.customer-service-page {
  min-height: calc(100vh - 58px);
  background: #eef3f8;
}

.customer-service-page__frame {
  display: block;
  width: 100%;
  height: calc(100vh - 58px);
  border: 0;
  background: #fff;
}

.customer-service-page__empty {
  min-height: calc(100vh - 58px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  color: var(--text-secondary);
  font-size: 0.92rem;
  font-weight: 700;
  text-align: center;
}
</style>
