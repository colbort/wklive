<template>
  <div class="system-core-config">
    <el-form-item :label="t('system.configValue')" prop="site_name">
      <el-input
        v-model="form.site_name"
        :placeholder="t('system.siteName') || t('common.pleaseEnter')"
      />
    </el-form-item>
    <el-form-item :label="t('system.siteLogo')" prop="site_logo">
      <div class="logo-upload-container">
        <el-image
          v-if="form.site_logo"
          :src="formatUrl(form.site_logo)"
          style="width: 100px; height: 100px; object-fit: contain; border: 1px solid #ddd; border-radius: 4px; margin-right: 12px;"
          :preview-teleported="true"
        />
        <div>
          <el-upload
            action="#"
            :auto-upload="false"
            :on-change="handleLogoChange"
            accept="image/*"
          >
            <el-button type="primary" :loading="uploading">
              {{ uploading ? t('common.uploading') : t('common.upload') }}
            </el-button>
          </el-upload>
          <p style="margin-top: 8px; font-size: 12px; color: #909399;">{{ t('common.uploadImageTip') }}</p>
        </div>
      </div>
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import type { SystemCore } from '@/services/system/ConfigService'
import { apiUploadAvatar } from '@/api/system/upload'
import { http } from '@/utils/request'

const { t } = useI18n()

interface Props {
  modelValue: SystemCore
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: SystemCore]
}>()

const uploading = ref(false)

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

async function handleLogoChange(uploadFile: UploadFile) {
  if (!uploadFile.raw) return

  try {
    uploading.value = true
    const res = await apiUploadAvatar(uploadFile.raw)
    if (res.code === 0 || res.code === 200) {
      form.value.site_logo = res.data?.url || ''
      ElMessage.success(t('common.uploadSuccess') || 'Upload successful')
    } else {
      throw new Error(res.msg || 'Upload failed')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.uploadFailed') || 'Upload failed')
  } finally {
    uploading.value = false
  }
}

function formatUrl(url: string | undefined) {
   console.log('formatUrl', url)
  if (!url) return ''
  const fullUrl = url.startsWith('http')
              ? url
              : `${http.defaults.baseURL}${url}`
  return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`
}
</script>

<style scoped>
.system-core-config {
  width: 100%;
}

.logo-upload-container {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
</style>
