<script setup lang="ts">
import { computed, ref, nextTick, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, type Locale } from '@/i18n'
import { useAuthStore, apiUpdateProfile } from '@/stores'
import { useRouter } from 'vue-router'
import { Expand, Fold, User, Setting, Lock } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { uploadService } from '@/services'
import { http } from '@/utils/request'
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
          console.error('更新密码失败:', err)
          ElMessage.error(t('app.updatePasswordFailed'))
        })
      console.log(t('app.newPasswordPrompt'), data.value)
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
          console.error('更新昵称失败:', err)
          ElMessage.error(t('app.updateNicknameFailed'))
        })
      console.log(t('app.newNicknamePrompt'), data.value)
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

type Box = {
  left: number
  top: number
  right: number
  bottom: number
  width: number
  height: number
}

function getRelativeBox(el: Element, relativeTo: Element): Box {
  const rect = el.getBoundingClientRect()
  const baseRect = relativeTo.getBoundingClientRect()

  const left = rect.left - baseRect.left
  const top = rect.top - baseRect.top
  const width = rect.width
  const height = rect.height

  return {
    left,
    top,
    right: left + width,
    bottom: top + height,
    width,
    height,
  }
}

function boxContains(outer: Box, inner: Box) {
  return (
    inner.left >= outer.left &&
    inner.top >= outer.top &&
    inner.right <= outer.right &&
    inner.bottom <= outer.bottom
  )
}

function fitImageToSelection(imageEl: any, selectionEl: any, canvasEl: any) {
  const imageBox = getRelativeBox(imageEl, canvasEl)
  const selectionBox = getRelativeBox(selectionEl, canvasEl)

  if (!imageBox.width || !imageBox.height) return

  const scaleX = selectionBox.width / imageBox.width
  const scaleY = selectionBox.height / imageBox.height
  const scale = Math.max(scaleX, scaleY)

  if (scale > 1 && typeof imageEl.$scale === 'function') {
    imageEl.$scale(scale)
  }
}

function ensureImageCoversSelection(imageEl: any, selectionEl: any, canvasEl: any) {
  // 先保证尺寸足够覆盖裁剪框
  fitImageToSelection(imageEl, selectionEl, canvasEl)

  const imageBox = getRelativeBox(imageEl, canvasEl)
  const selectionBox = getRelativeBox(selectionEl, canvasEl)

  let moveX = 0
  let moveY = 0

  if (imageBox.left > selectionBox.left) {
    moveX = selectionBox.left - imageBox.left
  } else if (imageBox.right < selectionBox.right) {
    moveX = selectionBox.right - imageBox.right
  }

  if (imageBox.top > selectionBox.top) {
    moveY = selectionBox.top - imageBox.top
  } else if (imageBox.bottom < selectionBox.bottom) {
    moveY = selectionBox.bottom - imageBox.bottom
  }

  if ((moveX !== 0 || moveY !== 0) && typeof imageEl.$move === 'function') {
    imageEl.$move(moveX, moveY)
  }
}

function canTransformKeepCover(imageEl: any, selectionEl: any, canvasEl: any, matrix: number[]) {
  const clone = imageEl.cloneNode() as HTMLElement
  clone.style.position = 'absolute'
  clone.style.visibility = 'hidden'
  clone.style.pointerEvents = 'none'
  clone.style.transform = `matrix(${matrix.join(',')})`

  canvasEl.appendChild(clone)

  const imageBox = getRelativeBox(clone, canvasEl)
  const selectionBox = getRelativeBox(selectionEl, canvasEl)

  canvasEl.removeChild(clone)

  return boxContains(imageBox, selectionBox)
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
              scalable="false"
              skewable="false"
              translatable
            ></cropper-image>

            <cropper-shade hidden></cropper-shade>

            <cropper-selection
              initial-coverage="1"
              aspect-ratio="1"
              movable="false"
              resizable="false"
            >
              <cropper-grid role="grid" covered></cropper-grid>
              <cropper-crosshair centered></cropper-crosshair>
            </cropper-selection>

            <cropper-handle
              action="move"
              plain
              theme-color="rgba(64, 158, 255, 0.18)"
            ></cropper-handle>
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

      // 初始化时保证图片完整覆盖裁剪框
      requestAnimationFrame(() => {
        ensureImageCoversSelection(imageEl, selectionEl, canvasEl)
      })

      // 拖动时边界检查：如果拖动后裁剪框会露空白，则阻止这次变换
      imageEl.addEventListener('transform', (event: Event) => {
        const e = event as CustomEvent<{ matrix: number[] }>
        if (!e.detail?.matrix) return

        if (!canTransformKeepCover(imageEl, selectionEl, canvasEl, e.detail.matrix)) {
          e.preventDefault()
        }
      })

      // 交互结束后再次兜底修正
      canvasEl.addEventListener('actionend', () => {
        requestAnimationFrame(() => {
          ensureImageCoversSelection(imageEl, selectionEl, canvasEl)
        })
      })
    }

    cropperImage.value.src = objectUrl
  }

  input.click()
}

async function confirmCrop() {
  if (!cropper) return

  try {
    const selection = cropper.getCropperSelection()
    if (!selection) {
      ElMessage.error(t('app.cropFailed'))
    }

    const canvas = await selection!.$toCanvas({
      width: 200,
      height: 200,
    })

    if (!canvas) {
      ElMessage.error(t('app.cropFailed'))
      return
    }

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.globalCompositeOperation = 'destination-over'
      ctx.fillStyle = '#ffffff'
      ctx.fillRect(0, 0, canvas.width, canvas.height)
    }

    canvas.toBlob(
      async (blob: Blob | null) => {
        if (!blob) {
          ElMessage.error(t('app.cropFailed'))
          return
        }

        try {
          const file = new File([blob], 'avatar.jpg', {
            type: 'image/jpeg',
          })

          const result = await uploadService.uploadAvatar(file)

          if (result.code !== 200) {
            throw new Error(result.msg || 'upload failed')
          }

          if (auth.user && result.data?.url) {
            apiUpdateProfile({ avatar: result.data.url }).catch((err) => {
              console.error('更新头像失败:', err)
              ElMessage.error(t('app.updateAvatarFailed'))
            })
            auth.user.avatar = result.data.url
          }

          ElMessage.success(t('app.avatarUpdated'))
          cropperDialogVisible.value = false
          resetCropperState()
        } catch (error) {
          console.error('头像上传失败:', error)
          ElMessage.error(t('app.avatarUploadFailed'))
        }
      },
      'image/jpeg',
      0.9,
    )
  } catch (error) {
    console.error(t('app.cropFailed'), error)
    ElMessage.error(t('app.cropFailed'))
  }
}

function formatAvatar(avatar: string | undefined) {
  if (!avatar) return ''
  const fullUrl = avatar.startsWith('http') ? avatar : `${http.defaults.baseURL}${avatar}`
  return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`
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
    width="520px"
    :before-close="cancelCrop"
    :destroy-on-close="true"
    :close-on-click-modal="false"
    append-to-body
  >
    <div class="cropper-dialog-body">
      <div class="cropper-stage"></div>
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
  width: 360px;
  height: 360px;
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
}

:deep(cropper-selection) {
  border-radius: 0;
}

:deep(cropper-grid),
:deep(cropper-crosshair),
:deep(cropper-handle) {
  color: #409eff;
}
</style>
