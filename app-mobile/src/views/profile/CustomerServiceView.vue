<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getAccessToken } from '@/api/http'
import { apiGetProfile } from '@/api/userPrivate'
import CommonPage from '@/components/common/CommonPage.vue'
import { useI18n } from '@/i18n'
import type { UserProfile } from '@/types/user'

const router = useRouter()
const { t } = useI18n()
const profile = ref<UserProfile | null>(null)
const profileLoaded = ref(false)

const chatFrameUrl = computed(() => {
  const baseUrl = import.meta.env.VITE_CHAT_UI_URL?.trim()
  const apiKey = import.meta.env.VITE_CHAT_API_KEY?.trim()
  const apiSecret = import.meta.env.VITE_CHAT_API_SECRET?.trim()
  const wsUrl = import.meta.env.VITE_CHAT_WS_URL?.trim()

  if (!baseUrl || !apiKey || !apiSecret) return ''
  if (getAccessToken() && (!profileLoaded.value || !profile.value?.user?.id)) return ''

  const url = new URL(baseUrl, window.location.origin)
  if (url.pathname.replace(/\/$/, '').endsWith('/chat')) {
    url.pathname = url.pathname.replace(/\/chat\/?$/, '') || '/'
  }
  url.searchParams.set('page', 'chat')
  url.searchParams.set('mode', 'mobile')
  url.searchParams.set('apiKey', apiKey)
  url.searchParams.set('apiSecret', apiSecret)
  if (wsUrl) {
    url.searchParams.set('wsUrl', wsUrl)
  }

  const user = profile.value?.user
  if (user?.id) {
    url.searchParams.set('userId', String(user.id))
  }
  const nickname = user?.nickname || user?.username
  if (nickname) {
    url.searchParams.set('nickname', nickname)
  }
  if (user?.avatar) {
    url.searchParams.set('avatarUrl', user.avatar)
  }

  return url.toString()
})

onMounted(() => {
  if (getAccessToken()) {
    void loadProfile()
  }
})

async function loadProfile() {
  try {
    const res = await apiGetProfile()
    if (res.code === 200) {
      profile.value = res.data
    }
  } catch (error) {
    console.warn('load chat profile failed', error)
  } finally {
    profileLoaded.value = true
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
        客服配置未完成
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
