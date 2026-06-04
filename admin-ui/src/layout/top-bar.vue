<script setup lang="ts">
import { computed, ref, nextTick, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/i18n'
import { useAuthStore, apiUpdateProfile } from '@/stores'
import { useRouter } from 'vue-router'
import { Expand, Fold, User, Setting, Lock } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { uploadService } from '@/services'
import { logger } from '@/utils/logger'
import { buildAssetUrl } from '@/utils/file-url'
import { resetDynamicRoutes } from '@/router'
import Cropper from 'cropperjs'

const props = defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-sider'): void
}>()

const { t, locale } = useI18n()
const auth = useAuthStore()
const router = useRouter()

const current = computed(() => locale.value as Locale)

const cropperDialogVisible = ref(false)
const cropperImage = ref<HTMLImageElement | null>(null)

let cropper: Cropper | null = null
let objectUrl = ''
let clampingSelection = false

function change(val: Locale) {
  setLocale(val)
}

function changePassword() {
  ElMessageBox.prompt(t('app.newPasswordPrompt'), t('app.changePassword'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    inputPattern: /^.{6,}$/,
    inputErrorMessage: t('app.passwordMinLength'),
  })
    .then((data: any) => {
      apiUpdateProfile({ password: data.value })
        .then(() => {
          ElMessage.success(t('app.passwordUpdated'))
        })
        .catch((err) => {
          logger.error('更新密码失败', err)
          ElMessage.error(t('app.updatePasswordFailed'))
        })
    })
    .catch(() => {})
}

function openSettings() {
  ElMessageBox.prompt(t('app.newNicknamePrompt'), t('app.settings'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    inputValue: auth.user?.nickname || '',
  })
    .then((data: any) => {
      apiUpdateProfile({ nickname: data.value })
        .then(() => {
          if (auth.user) {
            auth.user.nickname = data.value
          }
          ElMessage.success(t('app.nicknameUpdated'))
        })
        .catch((err) => {
          logger.error('更新昵称失败', err)
          ElMessage.error(t('app.updateNicknameFailed'))
        })
    })
    .catch(() => {})
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

function clearObjectUrl() {
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    objectUrl = ''
  }
}

function resetCropperState() {
  destroyCropper()
  clearObjectUrl()
}

type RectBox = {
  left: number
  top: number
  right: number
  bottom: number
  width: number
  height: number
}

type CropperSelectionChange = {
  x: number
  y: number
  width: number
  height: number
}

function getRelativeRect(el: Element, relativeTo: Element): RectBox {
  const rect = el.getBoundingClientRect()
  const baseRect = relativeTo.getBoundingClientRect()
  const left = rect.left - baseRect.left
  const top = rect.top - baseRect.top

  return {
    left,
    top,
    right: left + rect.width,
    bottom: top + rect.height,
    width: rect.width,
    height: rect.height,
  }
}

function clamp(value: number, min: number, max: number) {
  if (max < min) return min
  return Math.min(Math.max(value, min), max)
}

function clampSelectionToImage(
  selection: CropperSelectionChange,
  imageEl: Element,
  canvasEl: Element,
) {
  const imageRect = getRelativeRect(imageEl, canvasEl)

  if (!imageRect.width || !imageRect.height) return selection

  const maxSize = Math.min(imageRect.width, imageRect.height)
  const size = Math.min(selection.width, maxSize)
  const x = clamp(selection.x, imageRect.left, imageRect.right - size)
  const y = clamp(selection.y, imageRect.top, imageRect.bottom - size)

  return {
    x: Math.round(x),
    y: Math.round(y),
    width: Math.round(size),
    height: Math.round(size),
  }
}

function selectionChanged(a: CropperSelectionChange, b: CropperSelectionChange) {
  return a.x !== b.x || a.y !== b.y || a.width !== b.width || a.height !== b.height
}

function keepSelectionInsideImage(selectionEl: any, imageEl: Element, canvasEl: Element) {
  const next = clampSelectionToImage(
    {
      x: selectionEl.x,
      y: selectionEl.y,
      width: selectionEl.width,
      height: selectionEl.height,
    },
    imageEl,
    canvasEl,
  )

  if (selectionChanged(next, selectionEl)) {
    selectionEl.$change(next.x, next.y, next.width, next.height, 1, true)
  }
}

function createAvatarCanvasFromVisibleSelection(
  imageSource: HTMLImageElement,
  imageEl: Element,
  selectionEl: Element,
) {
  const imageRect = imageEl.getBoundingClientRect()
  const selectionRect = selectionEl.getBoundingClientRect()
  const naturalWidth = imageSource.naturalWidth
  const naturalHeight = imageSource.naturalHeight

  if (!imageRect.width || !imageRect.height || !naturalWidth || !naturalHeight) {
    return null
  }

  const sourceX = ((selectionRect.left - imageRect.left) / imageRect.width) * naturalWidth
  const sourceY = ((selectionRect.top - imageRect.top) / imageRect.height) * naturalHeight
  const sourceWidth = (selectionRect.width / imageRect.width) * naturalWidth
  const sourceHeight = (selectionRect.height / imageRect.height) * naturalHeight

  const avatarCanvas = document.createElement('canvas')
  avatarCanvas.width = 200
  avatarCanvas.height = 200

  const ctx = avatarCanvas.getContext('2d')
  if (!ctx) return null

  ctx.save()
  ctx.beginPath()
  ctx.arc(100, 100, 100, 0, Math.PI * 2)
  ctx.clip()
  ctx.drawImage(
    imageSource,
    sourceX,
    sourceY,
    sourceWidth,
    sourceHeight,
    0,
    0,
    avatarCanvas.width,
    avatarCanvas.height,
  )
  ctx.restore()

  return avatarCanvas
}

async function onAvatarClick() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'

  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      ElMessage.error(t('app.pleaseSelectImageFile'))
      return
    }

    if (file.size > 5 * 1024 * 1024) {
      ElMessage.error(t('app.avatarSizeLimit'))
      return
    }

    resetCropperState()
    objectUrl = URL.createObjectURL(file)
    cropperDialogVisible.value = true

    await nextTick()

    if (!cropperImage.value) return

    cropperImage.value.onload = async () => {
      if (!cropperImage.value) return

      destroyCropper()

      cropper = new Cropper(cropperImage.value, {
        container: '.cropper-stage',
        template: `
          <cropper-canvas background>
            <cropper-image
              rotatable="false"
              scalable
              skewable="false"
            ></cropper-image>

            <cropper-shade></cropper-shade>

            <cropper-selection
              initial-coverage="0.56"
              aspect-ratio="1"
              movable
              resizable
            >
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

      const imageEl = cropper.getCropperImage()
      const selectionEl = cropper.getCropperSelection()
      const canvasEl = cropper.getCropperCanvas()

      if (!imageEl || !selectionEl || !canvasEl) return

      if (typeof imageEl.$ready === 'function') {
        await imageEl.$ready()
      }

      selectionEl.addEventListener('change', (event: Event) => {
        if (clampingSelection) return

        const e = event as CustomEvent<CropperSelectionChange>
        if (!e.detail) return

        const next = clampSelectionToImage(e.detail, imageEl, canvasEl)
        if (!selectionChanged(next, e.detail)) return

        e.preventDefault()
        requestAnimationFrame(() => {
          clampingSelection = true
          selectionEl.$change(next.x, next.y, next.width, next.height, 1, true)
          clampingSelection = false
        })
      })

      requestAnimationFrame(() => {
        keepSelectionInsideImage(selectionEl, imageEl, canvasEl)
      })
    }

    cropperImage.value.src = objectUrl
  }

  input.click()
}

async function confirmCrop() {
  if (!cropper) return

  try {
    const imageEl = cropper.getCropperImage()
    const selectionEl = cropper.getCropperSelection()

    if (!cropperImage.value || !imageEl || !selectionEl) {
      ElMessage.error(t('app.cropFailed'))
      return
    }

    const avatarCanvas = createAvatarCanvasFromVisibleSelection(
      cropperImage.value,
      imageEl,
      selectionEl,
    )
    if (!avatarCanvas) {
      ElMessage.error(t('app.cropFailed'))
      return
    }

    avatarCanvas.toBlob(async (blob: Blob | null) => {
      if (!blob) {
        ElMessage.error(t('app.cropFailed'))
        return
      }

      try {
        const file = new File([blob], 'avatar.png', {
          type: 'image/png',
        })

        const result = await uploadService.uploadAvatar(file)

        if (result.code !== 200) {
          throw new Error(result.msg || 'upload failed')
        }

        if (auth.user && result.data?.url) {
          apiUpdateProfile({ avatar: result.data.url }).catch((err) => {
            logger.error('更新头像失败', err)
            ElMessage.error(t('app.updateAvatarFailed'))
          })
          auth.user.avatar = result.data.url
        }

        ElMessage.success(t('app.avatarUpdated'))
        cropperDialogVisible.value = false
        resetCropperState()
      } catch (error) {
        logger.error('头像上传失败', error)
        ElMessage.error(t('app.avatarUploadFailed'))
      }
    }, 'image/png')
  } catch (error) {
    logger.error(t('app.cropFailed'), error)
    ElMessage.error(t('app.cropFailed'))
  }
}

function formatAvatar(avatar: string | undefined) {
  return buildAssetUrl(avatar)
}

function cancelCrop() {
  cropperDialogVisible.value = false
  resetCropperState()
}

function handleCommand(command: string) {
  switch (command) {
    case 'changePassword':
      changePassword()
      break
    case 'settings':
      openSettings()
      break
    case 'logout':
      logout()
      break
  }
}

function logout() {
  auth.logout()
  resetDynamicRoutes()
  router.push('/login')
}

onBeforeUnmount(() => {
  resetCropperState()
})
</script>

<template>
  <div class="topbar">
    <div class="left">
      <el-button text class="collapse-btn" @click="emit('toggle-sider')">
        <el-icon>
          <component :is="props.collapsed ? Expand : Fold" />
        </el-icon>
      </el-button>

      <span class="title">{{ t('app.title') }}</span>
    </div>

    <div class="right">
      <el-select style="width: 140px" :model-value="current" @update:model-value="change">
        <el-option label="中文" value="zh-CN" />
        <el-option label="English" value="en-US" />
      </el-select>

      <el-dropdown trigger="contextmenu" @command="handleCommand">
        <div
          class="avatar-container"
          :title="t('app.uploadAvatar')"
          @click.stop.prevent="onAvatarClick"
        >
          <el-avatar
            :size="32"
            :src="formatAvatar(auth.user?.avatar)"
            :alt="auth.user?.nickname || auth.user?.username"
          >
            <el-icon><User /></el-icon>
          </el-avatar>
          <span class="user-nickname">{{ auth.user?.nickname || auth.user?.username }}</span>
        </div>

        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="changePassword">
              <el-icon><Lock /></el-icon>
              {{ t('app.changePassword') }}
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              {{ t('app.settings') }}
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              {{ t('app.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>

  <el-dialog
    v-model="cropperDialogVisible"
    :title="t('app.cropAvatar')"
    width="640px"
    :before-close="cancelCrop"
    :destroy-on-close="true"
    :close-on-click-modal="false"
    append-to-body
  >
    <div class="cropper-dialog-body">
      <div class="cropper-stage" />
      <img ref="cropperImage" alt="avatar preview" style="display: none" />
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="cancelCrop">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmCrop">{{ t('common.confirm') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<style scoped>
.topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.title {
  font-weight: 700;
}

.right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.avatar-container {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-nickname {
  font-size: 14px;
  color: #333;
  font-weight: 500;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collapse-btn {
  padding: 6px;
}

.cropper-dialog-body {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0;
}

.cropper-stage {
  width: min(520px, 100%);
  height: 440px;
  margin: 0 auto;
  overflow: hidden;
  border-radius: 8px;
  background: #f5f7fa;
}

:deep(cropper-canvas) {
  width: 100%;
  height: 100%;
  display: block;
}

:deep(cropper-image) {
  display: block;
  max-width: 100%;
  max-height: 100%;
}

:deep(cropper-selection) {
  border: 2px solid #ff3b30;
  border-radius: 50%;
  background: transparent;
  box-sizing: border-box;
}

:deep(cropper-shade) {
  background: rgba(0, 0, 0, 0.24);
}

:deep(cropper-crosshair) {
  color: rgba(255, 59, 48, 0.28);
}

:deep(cropper-handle[action='move']) {
  background: transparent;
}

:deep(cropper-handle[action$='resize']) {
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #ff3b30;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.28);
}
</style>
