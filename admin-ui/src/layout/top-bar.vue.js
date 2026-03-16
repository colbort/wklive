import { computed, ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { setLocale } from '@/i18n';
import { useAuthStore } from '@/stores';
import { useRouter } from 'vue-router';
import { Expand, Fold, User, Setting, Lock } from '@element-plus/icons-vue';
import { ElMessageBox, ElMessage } from 'element-plus';
import { apiUploadAvatar } from '@/api/system/upload';
import { http } from '@/utils/request';
import Cropper from 'cropperjs';
const props = defineProps();
const emit = defineEmits();
const { t, locale } = useI18n();
const auth = useAuthStore();
const router = useRouter();
const current = computed(() => locale.value);
// Avatar cropper variables
const cropperDialogVisible = ref(false);
const cropperImage = ref();
let cropper = null;
function change(val) {
    setLocale(val);
}
function changePassword() {
    // TODO: 实现修改密码逻辑，可以打开一个对话框
    ElMessageBox.prompt(t('app.newPasswordPrompt'), t('app.changePassword'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputPattern: /^.{6,}$/,
        inputErrorMessage: t('app.passwordMinLength'),
    }).then((data) => {
        // 调用API修改密码
        console.log('新密码:', data.value);
        // auth.changePassword(data.value)
    }).catch(() => {
        console.log('取消修改密码');
    });
}
function openSettings() {
    // TODO: 实现设置逻辑，可以打开一个对话框修改昵称等
    ElMessageBox.prompt(t('app.newNicknamePrompt'), t('app.settings'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputValue: auth.user?.nickname || '',
    }).then((data) => {
        // 调用API修改昵称
        console.log('新昵称:', data.value);
        // auth.updateProfile({ nickname: data.value })
    }).catch(() => {
        console.log('取消设置');
    });
}
// Avatar upload handling
async function onAvatarClick() {
    // trigger hidden file input
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';
    input.onchange = async (e) => {
        const file = e.target.files?.[0];
        if (!file)
            return;
        // Validate file size (e.g., max 5MB)
        if (file.size > 5 * 1024 * 1024) {
            ElMessage.error(t('app.avatarSizeLimit'));
            return;
        }
        // Create object URL for preview
        const url = URL.createObjectURL(file);
        // Show cropper dialog
        cropperDialogVisible.value = true;
        // Wait for next tick to ensure DOM is updated
        await nextTick();
        if (cropperImage.value) {
            cropperImage.value.src = url;
            cropper = new Cropper(cropperImage.value, {
                aspectRatio: 1, // Square for avatar
                viewMode: 1,
                autoCropArea: 0.8,
                responsive: true,
                restore: false,
                modal: true,
                guides: true,
                center: true,
                highlight: false,
                background: false,
                scalable: true,
                zoomable: true,
                zoomOnTouch: true,
                zoomOnWheel: true,
                cropBoxMovable: true,
                cropBoxResizable: true,
            });
        }
    };
    input.click();
}
function confirmCrop() {
    if (!cropper)
        return;
    const canvas = cropper.getCroppedCanvas({
        width: 200,
        height: 200,
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high',
    });
    canvas.toBlob(async (blob) => {
        if (!blob) {
            ElMessage.error('裁切失败');
            return;
        }
        try {
            // Convert blob to file
            const file = new File([blob], 'avatar.jpg', { type: 'image/jpeg' });
            // Upload cropped image
            const result = await apiUploadAvatar(file);
            if (result.code !== 200) {
                throw new Error(result.msg || 'Upload failed');
            }
            // Update user avatar
            if (auth.user && result.data?.url) {
                const fullUrl = result.data.url.startsWith('http')
                    ? result.data.url
                    : `${http.defaults.baseURL}${result.data.url}`;
                auth.user.avatar = fullUrl;
                ElMessage.success(t('app.avatarUpdated'));
            }
            // Close dialog
            cropperDialogVisible.value = false;
            destroyCropper();
        }
        catch (error) {
            console.error('Avatar upload error:', error);
            ElMessage.error(t('app.avatarUploadFailed'));
        }
    }, 'image/jpeg', 0.9);
}
function cancelCrop() {
    cropperDialogVisible.value = false;
    destroyCropper();
}
function destroyCropper() {
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
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
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['cropper-container']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-container']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-container']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-view-box']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-face']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-center']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-dashed']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-dashed']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-point']} */ ;
/** @type {__VLS_StyleScopedClasses['point-se']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-line']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-line']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-line']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-line']} */ ;
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
    src: (__VLS_ctx.auth.user?.avatar),
    alt: (__VLS_ctx.auth.user?.nickname || __VLS_ctx.auth.user?.username),
}));
const __VLS_42 = __VLS_41({
    size: (32),
    src: (__VLS_ctx.auth.user?.avatar),
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
    width: "600px",
    beforeClose: (__VLS_ctx.cancelCrop),
}));
const __VLS_86 = __VLS_85({
    modelValue: (__VLS_ctx.cropperDialogVisible),
    title: (__VLS_ctx.t('app.cropAvatar')),
    width: "600px",
    beforeClose: (__VLS_ctx.cancelCrop),
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "cropper-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
    ref: "cropperImage",
    alt: "Avatar Preview",
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
/** @type {__VLS_StyleScopedClasses['cropper-container']} */ ;
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
