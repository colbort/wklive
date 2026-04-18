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
          :src="buildAssetUrl(form.site_logo)"
          style="
            width: 100px;
            height: 100px;
            object-fit: contain;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-right: 12px;
          "
          :preview-teleported="true"
        />
        <div>
          <el-upload action="#" :auto-upload="false" :on-change="handleLogoSelect" accept="image/*">
            <el-button type="primary">
              {{ t('app.pleaseSelectImageFile') }}
            </el-button>
          </el-upload>
          <p style="margin-top: 8px; font-size: 12px; color: #909399">
            {{ t('common.uploadImageTip') }}
          </p>
        </div>
      </div>
    </el-form-item>

    <!-- Image Crop Dialog -->
    <el-dialog
      v-model="showCropDialog"
      :title="t('common.cropImage') || 'Crop Image'"
      width="700px"
      center
      destroy-on-close
      @close="resetCrop"
    >
      <div ref="cropperStageRef" class="cropper-stage">
        <img
          v-if="previewImageUrl"
          ref="imageRef"
          :src="previewImageUrl"
          alt="Crop preview"
          @load="initCropper"
        />
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="resetCrop">
            {{ t('common.cancel') || 'Cancel' }}
          </el-button>
          <el-button type="primary" :loading="uploading" @click="confirmCrop">
            {{ t('common.cropAndUpload') || 'Crop & Upload' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, nextTick, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import Cropper from 'cropperjs'
import type { UploadFile } from 'element-plus'
import type { SystemCore } from '@/services/system/ConfigService'
import { apiUploadAvatar } from '@/api/system/upload'
import { buildAssetUrl } from '@/utils/file-url'

const { t } = useI18n()

interface Props {
  modelValue: SystemCore
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: SystemCore]
}>()

const uploading = ref(false)
const showCropDialog = ref(false)
const previewImageUrl = ref('')
const imageRef = ref<HTMLImageElement | null>(null)
const cropperStageRef = ref<HTMLElement | null>(null)

let cropper: Cropper | null = null
let currentUploadFile: UploadFile | null = null
let objectUrl = ''

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

function clearObjectUrl() {
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    objectUrl = ''
  }
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

function resetCrop() {
  destroyCropper()
  showCropDialog.value = false
  previewImageUrl.value = ''
  currentUploadFile = null
  clearObjectUrl()
}

async function initCropper() {
  await nextTick()

  if (!showCropDialog.value || !imageRef.value) return

  destroyCropper()

  cropper = new Cropper(imageRef.value, {
    container: cropperStageRef.value || '.cropper-stage',
    template: `
      <cropper-canvas background>
        <cropper-image
          rotatable="false"
          scalable
          skewable="false"
          translatable
        ></cropper-image>

        <cropper-shade></cropper-shade>

        <cropper-selection
          initial-coverage="0.6"
          aspect-ratio="1"
          movable
          resizable
        >
          <cropper-grid role="grid" covered></cropper-grid>
          <cropper-crosshair centered></cropper-crosshair>

          <cropper-handle action="move" plain></cropper-handle>
          <cropper-handle action="n-resize"></cropper-handle>
          <cropper-handle action="e-resize"></cropper-handle>
          <cropper-handle action="s-resize"></cropper-handle>
          <cropper-handle action="w-resize"></cropper-handle>
          <cropper-handle action="ne-resize"></cropper-handle>
          <cropper-handle action="nw-resize"></cropper-handle>
          <cropper-handle action="se-resize"></cropper-handle>
          <cropper-handle action="sw-resize"></cropper-handle>
        </cropper-selection>
      </cropper-canvas>
    `,
  })
}

function handleLogoSelect(uploadFile: UploadFile) {
  if (!uploadFile.raw) return

  if (!uploadFile.raw.type.startsWith('image/')) {
    ElMessage.error(t('app.pleaseSelectImageFile') || 'Please select an image')
    return
  }

  if (uploadFile.raw.size > 5 * 1024 * 1024) {
    ElMessage.error(t('app.avatarSizeLimit') || 'Image size cannot exceed 5MB')
    return
  }

  destroyCropper()
  clearObjectUrl()

  objectUrl = URL.createObjectURL(uploadFile.raw)
  previewImageUrl.value = objectUrl
  currentUploadFile = uploadFile
  showCropDialog.value = true
}

async function confirmCrop() {
  if (!cropper || !currentUploadFile?.raw) return

  try {
    uploading.value = true

    const selection = cropper.getCropperSelection()
    if (!selection) {
      throw new Error(t('common.cropFailed') || 'Failed to crop image')
    }

    const canvas = await selection.$toCanvas({
      width: 100,
      height: 100,
    })

    if (!canvas) {
      throw new Error(t('common.cropFailed') || 'Failed to crop image')
    }

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.globalCompositeOperation = 'destination-over'
      ctx.fillStyle = '#ffffff'
      ctx.fillRect(0, 0, canvas.width, canvas.height)
    }

    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob(resolve, 'image/jpeg', 0.9)
    })

    if (!blob) {
      throw new Error(t('common.cropFailed') || 'Failed to crop image')
    }

    const croppedFile = new File([blob], currentUploadFile.raw.name, {
      type: 'image/jpeg',
    })

    const res = await apiUploadAvatar(croppedFile)

    if (res.code === 0 || res.code === 200) {
      form.value.site_logo = res.data?.url || ''
      ElMessage.success(t('common.uploadSuccess') || 'Upload successful')
      resetCrop()
    } else {
      throw new Error(res.msg || 'Upload failed')
    }
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : t('common.uploadFailed') || 'Upload failed')
  } finally {
    uploading.value = false
  }
}

watch(showCropDialog, (newVal) => {
  if (!newVal) {
    destroyCropper()
  }
})

onBeforeUnmount(() => {
  destroyCropper()
  clearObjectUrl()
})
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

.cropper-stage {
  width: 100%;
  height: 420px;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.cropper-stage img {
  display: block;
  max-width: 100%;
}

.cropper-stage :deep(cropper-canvas) {
  width: 100%;
  height: 100%;
}

.cropper-stage :deep(cropper-image) {
  max-width: 100%;
}

.cropper-stage :deep(cropper-shade) {
  background: rgba(0, 0, 0, 0.45);
}

.cropper-stage :deep(cropper-selection) {
  background: transparent;
  border: 1px solid #409eff;
}

.cropper-stage :deep(cropper-grid) {
  color: rgba(255, 255, 255, 0.35);
}

.cropper-stage :deep(cropper-handle[action='move']) {
  background: transparent;
}

.cropper-stage :deep(cropper-handle[action$='resize']) {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #409eff;
  border: 1px solid #fff;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
