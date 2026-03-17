import { computed, ref, nextTick, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { setLocale } from '@/i18n';
import { useAuthStore, apiUpdateProfile } from '@/stores';
import { useRouter } from 'vue-router';
import { Expand, Fold, User, Setting, Lock } from '@element-plus/icons-vue';
import { ElMessageBox, ElMessage } from 'element-plus';
import { uploadService } from '@/services';
import { http } from '@/utils/request';
import Cropper from 'cropperjs';
const props = defineProps();
const emit = defineEmits();
const { t, locale } = useI18n();
const auth = useAuthStore();
const router = useRouter();
const current = computed(() => locale.value);
const cropperDialogVisible = ref(false);
const cropperImage = ref(null);
let cropper = null;
let objectUrl = '';
function change(val) {
    setLocale(val);
}
function changePassword() {
    ElMessageBox.prompt(t('app.newPasswordPrompt'), t('app.changePassword'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputPattern: /^.{6,}$/,
        inputErrorMessage: t('app.passwordMinLength'),
    })
        .then((data) => {
        apiUpdateProfile({ password: data.value })
            .then(() => {
            ElMessage.success(t('app.passwordUpdated'));
        })
            .catch((err) => {
            console.error('更新密码失败:', err);
            ElMessage.error(t('app.updatePasswordFailed'));
        });
        console.log(t('app.newPasswordPrompt'), data.value);
    })
        .catch(() => { });
}
function openSettings() {
    ElMessageBox.prompt(t('app.newNicknamePrompt'), t('app.settings'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputValue: auth.user?.nickname || '',
    })
        .then((data) => {
        apiUpdateProfile({ nickname: data.value })
            .then(() => {
            if (auth.user) {
                auth.user.nickname = data.value;
            }
            ElMessage.success(t('app.nicknameUpdated'));
        })
            .catch((err) => {
            console.error('更新昵称失败:', err);
            ElMessage.error(t('app.updateNicknameFailed'));
        });
        console.log(t('app.newNicknamePrompt'), data.value);
    })
        .catch(() => { });
}
function destroyCropper() {
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
}
function clearObjectUrl() {
    if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
        objectUrl = '';
    }
}
function resetCropperState() {
    destroyCropper();
    clearObjectUrl();
}
function getRelativeBox(el, relativeTo) {
    const rect = el.getBoundingClientRect();
    const baseRect = relativeTo.getBoundingClientRect();
    const left = rect.left - baseRect.left;
    const top = rect.top - baseRect.top;
    const width = rect.width;
    const height = rect.height;
    return {
        left,
        top,
        right: left + width,
        bottom: top + height,
        width,
        height,
    };
}
function boxContains(outer, inner) {
    return (inner.left >= outer.left &&
        inner.top >= outer.top &&
        inner.right <= outer.right &&
        inner.bottom <= outer.bottom);
}
function fitImageToSelection(imageEl, selectionEl, canvasEl) {
    const imageBox = getRelativeBox(imageEl, canvasEl);
    const selectionBox = getRelativeBox(selectionEl, canvasEl);
    if (!imageBox.width || !imageBox.height)
        return;
    const scaleX = selectionBox.width / imageBox.width;
    const scaleY = selectionBox.height / imageBox.height;
    const scale = Math.max(scaleX, scaleY);
    if (scale > 1 && typeof imageEl.$scale === 'function') {
        imageEl.$scale(scale);
    }
}
function ensureImageCoversSelection(imageEl, selectionEl, canvasEl) {
    // 先保证尺寸足够覆盖裁剪框
    fitImageToSelection(imageEl, selectionEl, canvasEl);
    const imageBox = getRelativeBox(imageEl, canvasEl);
    const selectionBox = getRelativeBox(selectionEl, canvasEl);
    let moveX = 0;
    let moveY = 0;
    if (imageBox.left > selectionBox.left) {
        moveX = selectionBox.left - imageBox.left;
    }
    else if (imageBox.right < selectionBox.right) {
        moveX = selectionBox.right - imageBox.right;
    }
    if (imageBox.top > selectionBox.top) {
        moveY = selectionBox.top - imageBox.top;
    }
    else if (imageBox.bottom < selectionBox.bottom) {
        moveY = selectionBox.bottom - imageBox.bottom;
    }
    if ((moveX !== 0 || moveY !== 0) && typeof imageEl.$move === 'function') {
        imageEl.$move(moveX, moveY);
    }
}
function canTransformKeepCover(imageEl, selectionEl, canvasEl, matrix) {
    const clone = imageEl.cloneNode();
    clone.style.position = 'absolute';
    clone.style.visibility = 'hidden';
    clone.style.pointerEvents = 'none';
    clone.style.transform = `matrix(${matrix.join(',')})`;
    canvasEl.appendChild(clone);
    const imageBox = getRelativeBox(clone, canvasEl);
    const selectionBox = getRelativeBox(selectionEl, canvasEl);
    canvasEl.removeChild(clone);
    return boxContains(imageBox, selectionBox);
}
async function onAvatarClick() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';
    input.onchange = async (e) => {
        const file = e.target.files?.[0];
        if (!file)
            return;
        if (!file.type.startsWith('image/')) {
            ElMessage.error(t('app.pleaseSelectImageFile'));
            return;
        }
        if (file.size > 5 * 1024 * 1024) {
            ElMessage.error(t('app.avatarSizeLimit'));
            return;
        }
        resetCropperState();
        objectUrl = URL.createObjectURL(file);
        cropperDialogVisible.value = true;
        await nextTick();
        if (!cropperImage.value)
            return;
        cropperImage.value.onload = async () => {
            if (!cropperImage.value)
                return;
            destroyCropper();
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
            });
            const imageEl = cropper.getCropperImage();
            const selectionEl = cropper.getCropperSelection();
            const canvasEl = cropper.getCropperCanvas();
            if (!imageEl || !selectionEl || !canvasEl)
                return;
            if (typeof imageEl.$ready === 'function') {
                await imageEl.$ready();
            }
            // 初始化时保证图片完整覆盖裁剪框
            requestAnimationFrame(() => {
                ensureImageCoversSelection(imageEl, selectionEl, canvasEl);
            });
            // 拖动时边界检查：如果拖动后裁剪框会露空白，则阻止这次变换
            imageEl.addEventListener('transform', (event) => {
                const e = event;
                if (!e.detail?.matrix)
                    return;
                if (!canTransformKeepCover(imageEl, selectionEl, canvasEl, e.detail.matrix)) {
                    e.preventDefault();
                }
            });
            // 交互结束后再次兜底修正
            canvasEl.addEventListener('actionend', () => {
                requestAnimationFrame(() => {
                    ensureImageCoversSelection(imageEl, selectionEl, canvasEl);
                });
            });
        };
        cropperImage.value.src = objectUrl;
    };
    input.click();
}
async function confirmCrop() {
    if (!cropper)
        return;
    try {
        const selection = cropper.getCropperSelection();
        if (!selection) {
            ElMessage.error(t('app.cropFailed'));
        }
        const canvas = await selection.$toCanvas({
            width: 200,
            height: 200,
        });
        if (!canvas) {
            ElMessage.error(t('app.cropFailed'));
            return;
        }
        const ctx = canvas.getContext('2d');
        if (ctx) {
            ctx.globalCompositeOperation = 'destination-over';
            ctx.fillStyle = '#ffffff';
            ctx.fillRect(0, 0, canvas.width, canvas.height);
        }
        canvas.toBlob(async (blob) => {
            if (!blob) {
                ElMessage.error(t('app.cropFailed'));
                return;
            }
            try {
                const file = new File([blob], 'avatar.jpg', {
                    type: 'image/jpeg',
                });
                const result = await uploadService.uploadAvatar(file);
                if (result.code !== 200) {
                    throw new Error(result.msg || 'upload failed');
                }
                if (auth.user && result.data?.url) {
                    apiUpdateProfile({ avatar: result.data.url }).catch((err) => {
                        console.error('更新头像失败:', err);
                        ElMessage.error(t('app.updateAvatarFailed'));
                    });
                    auth.user.avatar = result.data.url;
                }
                ElMessage.success(t('app.avatarUpdated'));
                cropperDialogVisible.value = false;
                resetCropperState();
            }
            catch (error) {
                console.error('头像上传失败:', error);
                ElMessage.error(t('app.avatarUploadFailed'));
            }
        }, 'image/jpeg', 0.9);
    }
    catch (error) {
        console.error(t('app.cropFailed'), error);
        ElMessage.error(t('app.cropFailed'));
    }
}
function formatAvatar(avatar) {
    if (!avatar)
        return '';
    const fullUrl = avatar.startsWith('http')
        ? avatar
        : `${http.defaults.baseURL}${avatar}`;
    return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`;
}
function cancelCrop() {
    cropperDialogVisible.value = false;
    resetCropperState();
}
function handleCommand(command) {
    switch (command) {
        case 'changePassword':
            changePassword();
            break;
        case 'settings':
            openSettings();
            break;
        case 'logout':
            logout();
            break;
    }
}
function logout() {
    auth.logout();
    router.push('/login');
}
onBeforeUnmount(() => {
    resetCropperState();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "topbar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "left" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    text: true,
    ...{ class: "collapse-btn" },
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    text: true,
    ...{ class: "collapse-btn" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (...[$event]) => {
        __VLS_ctx.emit('toggle-sider');
    }
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = ((props.collapsed ? __VLS_ctx.Expand : __VLS_ctx.Fold));
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "title" },
});
(__VLS_ctx.t('app.title'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "right" },
});
const __VLS_16 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onUpdate:modelValue': {} },
    ...{ style: {} },
    modelValue: (__VLS_ctx.current),
}));
const __VLS_18 = __VLS_17({
    ...{ 'onUpdate:modelValue': {} },
    ...{ style: {} },
    modelValue: (__VLS_ctx.current),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    'onUpdate:modelValue': (__VLS_ctx.change)
};
__VLS_19.slots.default;
const __VLS_24 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "中文",
    value: "zh-CN",
}));
const __VLS_26 = __VLS_25({
    label: "中文",
    value: "zh-CN",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "English",
    value: "en-US",
}));
const __VLS_30 = __VLS_29({
    label: "English",
    value: "en-US",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
var __VLS_19;
const __VLS_32 = {}.ElDropdown;
/** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ...{ 'onCommand': {} },
    trigger: "contextmenu",
}));
const __VLS_34 = __VLS_33({
    ...{ 'onCommand': {} },
    trigger: "contextmenu",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
let __VLS_36;
let __VLS_37;
let __VLS_38;
const __VLS_39 = {
    onCommand: (__VLS_ctx.handleCommand)
};
__VLS_35.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ onClick: (__VLS_ctx.onAvatarClick) },
    ...{ class: "avatar-container" },
    title: (__VLS_ctx.t('app.uploadAvatar')),
});
const __VLS_40 = {}.ElAvatar;
/** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    size: (32),
    src: (__VLS_ctx.formatAvatar(__VLS_ctx.auth.user?.avatar)),
    alt: (__VLS_ctx.auth.user?.nickname || __VLS_ctx.auth.user?.username),
}));
const __VLS_42 = __VLS_41({
    size: (32),
    src: (__VLS_ctx.formatAvatar(__VLS_ctx.auth.user?.avatar)),
    alt: (__VLS_ctx.auth.user?.nickname || __VLS_ctx.auth.user?.username),
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
const __VLS_48 = {}.User;
/** @type {[typeof __VLS_components.User, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
var __VLS_47;
var __VLS_43;
{
    const { dropdown: __VLS_thisSlot } = __VLS_35.slots;
    const __VLS_52 = {}.ElDropdownMenu;
    /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
    const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
    __VLS_55.slots.default;
    const __VLS_56 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        command: "changePassword",
    }));
    const __VLS_58 = __VLS_57({
        command: "changePassword",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    const __VLS_60 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({}));
    const __VLS_62 = __VLS_61({}, ...__VLS_functionalComponentArgsRest(__VLS_61));
    __VLS_63.slots.default;
    const __VLS_64 = {}.Lock;
    /** @type {[typeof __VLS_components.Lock, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
    const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
    var __VLS_63;
    (__VLS_ctx.t('app.changePassword'));
    var __VLS_59;
    const __VLS_68 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
        command: "settings",
    }));
    const __VLS_70 = __VLS_69({
        command: "settings",
    }, ...__VLS_functionalComponentArgsRest(__VLS_69));
    __VLS_71.slots.default;
    const __VLS_72 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({}));
    const __VLS_74 = __VLS_73({}, ...__VLS_functionalComponentArgsRest(__VLS_73));
    __VLS_75.slots.default;
    const __VLS_76 = {}.Setting;
    /** @type {[typeof __VLS_components.Setting, ]} */ ;
    // @ts-ignore
    const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({}));
    const __VLS_78 = __VLS_77({}, ...__VLS_functionalComponentArgsRest(__VLS_77));
    var __VLS_75;
    (__VLS_ctx.t('app.settings'));
    var __VLS_71;
    const __VLS_80 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        command: "logout",
        divided: true,
    }));
    const __VLS_82 = __VLS_81({
        command: "logout",
        divided: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    (__VLS_ctx.t('app.logout'));
    var __VLS_83;
    var __VLS_55;
}
var __VLS_35;
const __VLS_84 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    modelValue: (__VLS_ctx.cropperDialogVisible),
    title: (__VLS_ctx.t('app.cropAvatar')),
    width: "520px",
    beforeClose: (__VLS_ctx.cancelCrop),
    destroyOnClose: (true),
    closeOnClickModal: (false),
    appendToBody: true,
}));
const __VLS_86 = __VLS_85({
    modelValue: (__VLS_ctx.cropperDialogVisible),
    title: (__VLS_ctx.t('app.cropAvatar')),
    width: "520px",
    beforeClose: (__VLS_ctx.cancelCrop),
    destroyOnClose: (true),
    closeOnClickModal: (false),
    appendToBody: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "cropper-dialog-body" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "cropper-stage" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
    ref: "cropperImage",
    alt: "avatar preview",
    ...{ style: {} },
});
/** @type {typeof __VLS_ctx.cropperImage} */ ;
{
    const { footer: __VLS_thisSlot } = __VLS_87.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "dialog-footer" },
    });
    const __VLS_88 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        ...{ 'onClick': {} },
    }));
    const __VLS_90 = __VLS_89({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    let __VLS_92;
    let __VLS_93;
    let __VLS_94;
    const __VLS_95 = {
        onClick: (__VLS_ctx.cancelCrop)
    };
    __VLS_91.slots.default;
    (__VLS_ctx.t('common.cancel'));
    var __VLS_91;
    const __VLS_96 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_98 = __VLS_97({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    let __VLS_100;
    let __VLS_101;
    let __VLS_102;
    const __VLS_103 = {
        onClick: (__VLS_ctx.confirmCrop)
    };
    __VLS_99.slots.default;
    (__VLS_ctx.t('common.confirm'));
    var __VLS_99;
}
var __VLS_87;
/** @type {__VLS_StyleScopedClasses['topbar']} */ ;
/** @type {__VLS_StyleScopedClasses['left']} */ ;
/** @type {__VLS_StyleScopedClasses['collapse-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['right']} */ ;
/** @type {__VLS_StyleScopedClasses['avatar-container']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-dialog-body']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['dialog-footer']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Expand: Expand,
            Fold: Fold,
            User: User,
            Setting: Setting,
            Lock: Lock,
            emit: emit,
            t: t,
            auth: auth,
            current: current,
            cropperDialogVisible: cropperDialogVisible,
            cropperImage: cropperImage,
            change: change,
            onAvatarClick: onAvatarClick,
            confirmCrop: confirmCrop,
            formatAvatar: formatAvatar,
            cancelCrop: cancelCrop,
            handleCommand: handleCommand,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
