<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import Cropper from "cropperjs";
import { updateProfile, uploadProfileAvatar } from "@/api/chat";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const merchantDrawerVisible = ref(false);
const passwordDialogVisible = ref(false);
const avatarDialogVisible = ref(false);
const avatarDrawerVisible = ref(false);
const isMobile = ref(false);
const profileSaving = ref(false);
const avatarInputRef = ref<HTMLInputElement>();
const cropperImage = ref<HTMLImageElement | null>(null);
const hasAvatarImage = ref(false);
let cropper: Cropper | null = null;
let avatarObjectUrl = "";
let clampingSelection = false;
const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});

type RectBox = {
  left: number;
  top: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
};

type CropperSelectionChange = {
  x: number;
  y: number;
  width: number;
  height: number;
};

const merchantMenu = [
  { path: "/merchant/agents", label: "坐席管理" },
  { path: "/merchant/groups", label: "客服分组" },
  { path: "/merchant/categories", label: "问题分类" },
];

const currentMenuLabel = computed(
  () =>
    merchantMenu.find((item) => item.path === route.path)?.label || "客服管理",
);
const merchantName = computed(
  () => auth.user?.nickname || auth.user?.username || "商户",
);
const merchantAvatarText = computed(() => merchantName.value.slice(0, 1));

async function logout() {
  await auth.logout();
  router.replace("/login");
}

function resetPasswordForm() {
  passwordForm.oldPassword = "";
  passwordForm.newPassword = "";
  passwordForm.confirmPassword = "";
}

function openPasswordDialog() {
  resetPasswordForm();
  passwordDialogVisible.value = true;
}

function openAvatarDialog() {
  resetAvatarCropper();
  if (isMobile.value) {
    avatarDrawerVisible.value = true;
  } else {
    avatarDialogVisible.value = true;
  }
}

function chooseAvatarFile() {
  avatarInputRef.value?.click();
}

function updateMobileState() {
  isMobile.value = window.matchMedia("(max-width: 768px)").matches;
}

async function handleSettingsCommand(command: string) {
  if (command === "password") {
    openPasswordDialog();
  } else if (command === "avatar") {
    openAvatarDialog();
  } else if (command === "logout") {
    await logout();
  }
}

async function submitPassword() {
  if (!passwordForm.oldPassword || !passwordForm.newPassword) {
    ElMessage.warning("请输入原密码和新密码");
    return;
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning("两次输入的新密码不一致");
    return;
  }
  profileSaving.value = true;
  try {
    await updateProfile({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    });
    passwordDialogVisible.value = false;
    ElMessage.success("密码已修改");
  } finally {
    profileSaving.value = false;
  }
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy();
    cropper = null;
  }
}

function clearAvatarObjectUrl() {
  if (avatarObjectUrl) {
    URL.revokeObjectURL(avatarObjectUrl);
    avatarObjectUrl = "";
  }
}

function resetAvatarCropper() {
  destroyCropper();
  clearAvatarObjectUrl();
  hasAvatarImage.value = false;
  clampingSelection = false;
}

function getRelativeRect(el: Element, relativeTo: Element): RectBox {
  const rect = el.getBoundingClientRect();
  const baseRect = relativeTo.getBoundingClientRect();
  const left = rect.left - baseRect.left;
  const top = rect.top - baseRect.top;

  return {
    left,
    top,
    right: left + rect.width,
    bottom: top + rect.height,
    width: rect.width,
    height: rect.height,
  };
}

function clamp(value: number, min: number, max: number) {
  if (max < min) return min;
  return Math.min(Math.max(value, min), max);
}

function clampSelectionToImage(
  selection: CropperSelectionChange,
  imageEl: Element,
  canvasEl: Element,
) {
  const imageRect = getRelativeRect(imageEl, canvasEl);

  if (!imageRect.width || !imageRect.height) return selection;

  const maxSize = Math.min(imageRect.width, imageRect.height);
  const size = Math.min(selection.width, maxSize);
  const x = clamp(selection.x, imageRect.left, imageRect.right - size);
  const y = clamp(selection.y, imageRect.top, imageRect.bottom - size);

  return {
    x: Math.round(x),
    y: Math.round(y),
    width: Math.round(size),
    height: Math.round(size),
  };
}

function selectionChanged(
  a: CropperSelectionChange,
  b: CropperSelectionChange,
) {
  return (
    a.x !== b.x || a.y !== b.y || a.width !== b.width || a.height !== b.height
  );
}

function keepSelectionInsideImage(
  selectionEl: CropperSelectionChange & {
    $change: (
      x: number,
      y: number,
      width: number,
      height: number,
      aspectRatio?: number,
      silent?: boolean,
    ) => void;
  },
  imageEl: Element,
  canvasEl: Element,
) {
  const next = clampSelectionToImage(
    {
      x: selectionEl.x,
      y: selectionEl.y,
      width: selectionEl.width,
      height: selectionEl.height,
    },
    imageEl,
    canvasEl,
  );

  if (selectionChanged(next, selectionEl)) {
    selectionEl.$change(next.x, next.y, next.width, next.height, 1, true);
  }
}

function createAvatarCanvasFromVisibleSelection(
  imageSource: HTMLImageElement,
  imageEl: Element,
  selectionEl: Element,
) {
  const imageRect = imageEl.getBoundingClientRect();
  const selectionRect = selectionEl.getBoundingClientRect();
  const naturalWidth = imageSource.naturalWidth;
  const naturalHeight = imageSource.naturalHeight;

  if (
    !imageRect.width ||
    !imageRect.height ||
    !naturalWidth ||
    !naturalHeight
  ) {
    return null;
  }

  const sourceX =
    ((selectionRect.left - imageRect.left) / imageRect.width) * naturalWidth;
  const sourceY =
    ((selectionRect.top - imageRect.top) / imageRect.height) * naturalHeight;
  const sourceWidth = (selectionRect.width / imageRect.width) * naturalWidth;
  const sourceHeight =
    (selectionRect.height / imageRect.height) * naturalHeight;

  const avatarCanvas = document.createElement("canvas");
  avatarCanvas.width = 200;
  avatarCanvas.height = 200;

  const ctx = avatarCanvas.getContext("2d");
  if (!ctx) return null;

  ctx.save();
  ctx.beginPath();
  ctx.arc(100, 100, 100, 0, Math.PI * 2);
  ctx.clip();
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
  );
  ctx.restore();

  return avatarCanvas;
}

async function createAvatarCropper() {
  await nextTick();
  if (!cropperImage.value) return;

  cropperImage.value.onload = async () => {
    if (!cropperImage.value) return;

    destroyCropper();
    cropper = new Cropper(cropperImage.value, {
      container: ".avatar-cropper-stage",
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
    });

    const imageEl = cropper.getCropperImage();
    const selectionEl = cropper.getCropperSelection();
    const canvasEl = cropper.getCropperCanvas();

    if (!imageEl || !selectionEl || !canvasEl) return;

    if (typeof imageEl.$ready === "function") {
      await imageEl.$ready();
    }

    selectionEl.addEventListener("change", (event: Event) => {
      if (clampingSelection) return;

      const e = event as CustomEvent<CropperSelectionChange>;
      if (!e.detail) return;

      const next = clampSelectionToImage(e.detail, imageEl, canvasEl);
      if (!selectionChanged(next, e.detail)) return;

      e.preventDefault();
      requestAnimationFrame(() => {
        clampingSelection = true;
        selectionEl.$change(next.x, next.y, next.width, next.height, 1, true);
        clampingSelection = false;
      });
    });

    requestAnimationFrame(() => {
      keepSelectionInsideImage(selectionEl, imageEl, canvasEl);
    });
  };

  cropperImage.value.src = avatarObjectUrl;
}

async function submitAvatar() {
  if (!cropper) {
    ElMessage.warning("请选择头像图片");
    return;
  }

  const imageEl = cropper.getCropperImage();
  const selectionEl = cropper.getCropperSelection();
  if (!cropperImage.value || !imageEl || !selectionEl) {
    ElMessage.warning("头像裁剪失败");
    return;
  }

  const avatarCanvas = createAvatarCanvasFromVisibleSelection(
    cropperImage.value,
    imageEl,
    selectionEl,
  );
  if (!avatarCanvas) {
    ElMessage.warning("头像裁剪失败");
    return;
  }

  profileSaving.value = true;
  try {
    await new Promise<void>((resolve, reject) => {
      avatarCanvas.toBlob(async (blob) => {
        if (!blob) {
          reject(new Error("crop failed"));
          return;
        }

        try {
          await uploadProfileAvatar(blob);
          await auth.fetchProfile();
          avatarDialogVisible.value = false;
          avatarDrawerVisible.value = false;
          resetAvatarCropper();
          ElMessage.success("头像已修改");
          resolve();
        } catch (error) {
          reject(error);
        }
      }, "image/png");
    });
  } finally {
    profileSaving.value = false;
  }
}

function cancelAvatarCrop() {
  avatarDialogVisible.value = false;
  avatarDrawerVisible.value = false;
  resetAvatarCropper();
}

async function onAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) return;

  if (!file.type.startsWith("image/")) {
    ElMessage.warning("请选择图片文件");
    return;
  }

  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning("头像图片不能超过 5MB");
    return;
  }

  resetAvatarCropper();
  avatarObjectUrl = URL.createObjectURL(file);
  hasAvatarImage.value = true;
  await createAvatarCropper();
}

function openMerchantMenu() {
  merchantDrawerVisible.value = true;
}

function goMerchant(path: string) {
  merchantDrawerVisible.value = false;
  router.push(path);
}

onBeforeUnmount(() => {
  window.removeEventListener("resize", updateMobileState);
  resetAvatarCropper();
});

onMounted(() => {
  updateMobileState();
  window.addEventListener("resize", updateMobileState);
});
</script>

<template>
  <div
    v-if="auth.isMerchant"
    class="merchant-shell"
  >
    <aside class="merchant-sidebar">
      <div class="brand">
        <button
          class="brand-mark brand-button"
          type="button"
          @click="openMerchantMenu"
        >
          <span class="brand-initials">CS</span>
          <el-avatar
            class="brand-avatar"
            :size="38"
            :src="auth.user?.avatarUrl"
          >
            {{ merchantAvatarText }}
          </el-avatar>
        </button>
        <div>
          <div class="brand-title">
            <span class="desktop-brand-text">客服工作台</span>
            <span class="mobile-brand-text">{{ currentMenuLabel }}</span>
          </div>
          <div class="brand-subtitle">
            <span class="desktop-brand-text">商户后台</span>
            <span class="mobile-brand-text">{{ merchantName }}</span>
          </div>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="item in merchantMenu"
          :key="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
          type="button"
          @click="goMerchant(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>

      <div class="sidebar-settings">
        <el-dropdown
          trigger="click"
          @command="handleSettingsCommand"
        >
          <button
            class="nav-item settings-trigger"
            type="button"
          >
            设置
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="password">
                修改密码
              </el-dropdown-item>
              <el-dropdown-item command="avatar">
                修改头像
              </el-dropdown-item>
              <el-dropdown-item command="logout">
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </aside>

    <section class="merchant-main">
      <header class="merchant-topbar">
        <h1>{{ currentMenuLabel }}</h1>
        <div class="merchant-profile">
          <div class="profile-text">
            <strong>{{ auth.user?.nickname }}</strong>
          </div>
          <el-avatar
            :size="36"
            :src="auth.user?.avatarUrl"
          >
            {{ auth.user?.nickname?.slice(0, 1) || "商" }}
          </el-avatar>
        </div>
      </header>

      <main class="merchant-content">
        <RouterView />
      </main>
    </section>

    <el-drawer
      v-model="merchantDrawerVisible"
      class="merchant-menu-drawer"
      direction="ltr"
      size="260px"
      :with-header="false"
    >
      <div class="drawer-brand">
        <div class="brand-mark">
          CS
        </div>
        <div>
          <div class="brand-title">
            客服工作台
          </div>
          <div class="brand-subtitle">
            商户后台
          </div>
        </div>
      </div>
      <nav class="nav drawer-nav">
        <button
          v-for="item in merchantMenu"
          :key="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
          type="button"
          @click="goMerchant(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>
    </el-drawer>
  </div>

  <div
    v-else
    class="agent-shell"
  >
    <header class="agent-topbar">
      <div class="brand agent-brand">
        <el-avatar
          class="agent-brand-avatar"
          :size="42"
          :src="auth.user?.avatarUrl"
        >
          {{ auth.user?.nickname?.slice(0, 1) || "坐" }}
        </el-avatar>
        <div>
          <div class="brand-title">
            {{ auth.user?.nickname || auth.user?.username || "坐席" }}
          </div>
          <div class="brand-subtitle">
            {{ auth.agent?.agentNo || "-" }}
          </div>
        </div>
      </div>
      <div class="profile">
        <el-button
          class="agent-profile-action"
          @click="openPasswordDialog"
        >
          修改密码
        </el-button>
        <el-button
          class="agent-profile-action"
          @click="openAvatarDialog"
        >
          修改头像
        </el-button>
        <el-button
          class="agent-profile-action"
          @click="logout"
        >
          退出
        </el-button>
        <el-dropdown
          class="agent-mobile-settings"
          trigger="click"
          @command="handleSettingsCommand"
        >
          <button
            class="agent-settings-trigger"
            type="button"
          >
            设置
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="password">
                修改密码
              </el-dropdown-item>
              <el-dropdown-item command="avatar">
                修改头像
              </el-dropdown-item>
              <el-dropdown-item command="logout">
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <main class="agent-content">
      <RouterView />
    </main>
  </div>

  <el-dialog
    v-model="passwordDialogVisible"
    title="修改密码"
    width="420px"
  >
    <el-form
      label-width="86px"
      :model="passwordForm"
    >
      <el-form-item label="原密码">
        <el-input
          v-model="passwordForm.oldPassword"
          type="password"
          show-password
        />
      </el-form-item>
      <el-form-item label="新密码">
        <el-input
          v-model="passwordForm.newPassword"
          type="password"
          show-password
        />
      </el-form-item>
      <el-form-item label="确认密码">
        <el-input
          v-model="passwordForm.confirmPassword"
          type="password"
          show-password
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="passwordDialogVisible = false">
        取消
      </el-button>
      <el-button
        type="primary"
        :loading="profileSaving"
        @click="submitPassword"
      >
        保存
      </el-button>
    </template>
  </el-dialog>

  <input
    ref="avatarInputRef"
    class="avatar-file-input"
    type="file"
    accept="image/*"
    @change="onAvatarFileChange"
  >

  <el-dialog
    v-if="!isMobile"
    v-model="avatarDialogVisible"
    title="修改头像"
    width="640px"
    :before-close="cancelAvatarCrop"
    :destroy-on-close="true"
    :close-on-click-modal="false"
    append-to-body
  >
    <div class="avatar-editor">
      <div class="avatar-cropper-stage">
        <button
          v-if="!hasAvatarImage"
          class="avatar-picker-empty"
          type="button"
          @click="chooseAvatarFile"
        >
          <span>选择图片</span>
        </button>
      </div>
      <img
        ref="cropperImage"
        alt="avatar preview"
        style="display: none"
      >
    </div>

    <template #footer>
      <el-button @click="cancelAvatarCrop">
        取消
      </el-button>
      <el-button
        type="primary"
        :loading="profileSaving"
        :disabled="!hasAvatarImage"
        @click="submitAvatar"
      >
        保存
      </el-button>
    </template>
  </el-dialog>

  <el-drawer
    v-if="isMobile"
    v-model="avatarDrawerVisible"
    title="修改头像"
    direction="btt"
    size="82vh"
    class="avatar-cropper-drawer"
    :before-close="cancelAvatarCrop"
    :destroy-on-close="true"
    :close-on-click-modal="false"
    append-to-body
  >
    <div class="avatar-editor">
      <div class="avatar-cropper-stage">
        <button
          v-if="!hasAvatarImage"
          class="avatar-picker-empty"
          type="button"
          @click="chooseAvatarFile"
        >
          <span>选择图片</span>
        </button>
      </div>
      <img
        ref="cropperImage"
        alt="avatar preview"
        style="display: none"
      >
    </div>

    <template #footer>
      <div class="avatar-drawer-footer">
        <el-button @click="cancelAvatarCrop">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="profileSaving"
          :disabled="!hasAvatarImage"
          @click="submitAvatar"
        >
          保存
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style scoped>
.avatar-editor {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0;
}

.avatar-file-input {
  display: none;
}

.avatar-cropper-stage {
  position: relative;
  width: min(520px, 100%);
  height: 440px;
  margin: 0 auto;
  overflow: hidden;
  border-radius: 8px;
  background: #f5f7fa;
}

.avatar-picker-empty {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  width: 100%;
  height: 100%;
  border: 1px dashed #cfd6e4;
  border-radius: inherit;
  background: #f8fafc;
  color: #667085;
  cursor: pointer;
  font-size: 15px;
}

.avatar-picker-empty:hover {
  border-color: #409eff;
  color: #409eff;
  background: #f4f9ff;
}

.avatar-drawer-footer {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
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

:deep(cropper-handle[action="move"]) {
  background: transparent;
}

:deep(cropper-handle[action$="resize"]) {
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #ff3b30;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.28);
}

:deep(.avatar-cropper-drawer) {
  border-radius: 16px 16px 0 0;
}

:deep(.avatar-cropper-drawer .el-drawer__body) {
  padding: 8px 16px 12px;
}

:deep(.avatar-cropper-drawer .el-drawer__footer) {
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  border-top: 1px solid #eef0f4;
}

@media (max-width: 768px) {
  .avatar-editor {
    padding: 0;
  }

  .avatar-cropper-stage {
    width: 100%;
    height: min(58vh, 460px);
    border-radius: 10px;
  }
}
</style>
