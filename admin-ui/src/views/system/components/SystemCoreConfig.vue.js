import { computed, ref, nextTick, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import Cropper from 'cropperjs';
import { apiUploadAvatar } from '@/api/system/upload';
import { http } from '@/utils/request';
const { t } = useI18n();
const props = defineProps();
const emit = defineEmits();
const uploading = ref(false);
const showCropDialog = ref(false);
const previewImageUrl = ref('');
const imageRef = ref(null);
const cropperStageRef = ref(null);
let cropper = null;
let currentUploadFile = null;
let objectUrl = '';
const form = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value),
});
function clearObjectUrl() {
    if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
        objectUrl = '';
    }
}
function destroyCropper() {
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
}
function resetCrop() {
    destroyCropper();
    showCropDialog.value = false;
    previewImageUrl.value = '';
    currentUploadFile = null;
    clearObjectUrl();
}
async function initCropper() {
    await nextTick();
    if (!showCropDialog.value || !imageRef.value)
        return;
    destroyCropper();
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
    });
}
function handleLogoSelect(uploadFile) {
    if (!uploadFile.raw)
        return;
    if (!uploadFile.raw.type.startsWith('image/')) {
        ElMessage.error(t('app.pleaseSelectImageFile') || 'Please select an image');
        return;
    }
    if (uploadFile.raw.size > 5 * 1024 * 1024) {
        ElMessage.error(t('app.avatarSizeLimit') || 'Image size cannot exceed 5MB');
        return;
    }
    destroyCropper();
    clearObjectUrl();
    objectUrl = URL.createObjectURL(uploadFile.raw);
    previewImageUrl.value = objectUrl;
    currentUploadFile = uploadFile;
    showCropDialog.value = true;
}
async function confirmCrop() {
    if (!cropper || !currentUploadFile?.raw)
        return;
    try {
        uploading.value = true;
        const selection = cropper.getCropperSelection();
        if (!selection) {
            throw new Error(t('common.cropFailed') || 'Failed to crop image');
        }
        const canvas = await selection.$toCanvas({
            width: 100,
            height: 100,
        });
        if (!canvas) {
            throw new Error(t('common.cropFailed') || 'Failed to crop image');
        }
        const ctx = canvas.getContext('2d');
        if (ctx) {
            ctx.globalCompositeOperation = 'destination-over';
            ctx.fillStyle = '#ffffff';
            ctx.fillRect(0, 0, canvas.width, canvas.height);
        }
        const blob = await new Promise((resolve) => {
            canvas.toBlob(resolve, 'image/jpeg', 0.9);
        });
        if (!blob) {
            throw new Error(t('common.cropFailed') || 'Failed to crop image');
        }
        const croppedFile = new File([blob], currentUploadFile.raw.name, {
            type: 'image/jpeg',
        });
        const res = await apiUploadAvatar(croppedFile);
        if (res.code === 0 || res.code === 200) {
            form.value.site_logo = res.data?.url || '';
            ElMessage.success(t('common.uploadSuccess') || 'Upload successful');
            resetCrop();
        }
        else {
            throw new Error(res.msg || 'Upload failed');
        }
    }
    catch (e) {
        ElMessage.error(e?.message || t('common.uploadFailed') || 'Upload failed');
    }
    finally {
        uploading.value = false;
    }
}
function formatUrl(url) {
    if (!url)
        return '';
    const fullUrl = url.startsWith('http') ? url : `${http.defaults.baseURL}${url}`;
    return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`;
}
watch(showCropDialog, (newVal) => {
    if (!newVal) {
        destroyCropper();
    }
});
onBeforeUnmount(() => {
    destroyCropper();
    clearObjectUrl();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "system-core-config" },
});
const __VLS_0 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    label: (__VLS_ctx.t('system.configValue')),
    prop: "site_name",
}));
const __VLS_2 = __VLS_1({
    label: (__VLS_ctx.t('system.configValue')),
    prop: "site_name",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    modelValue: (__VLS_ctx.form.site_name),
    placeholder: (__VLS_ctx.t('system.siteName') || __VLS_ctx.t('common.pleaseEnter')),
}));
const __VLS_6 = __VLS_5({
    modelValue: (__VLS_ctx.form.site_name),
    placeholder: (__VLS_ctx.t('system.siteName') || __VLS_ctx.t('common.pleaseEnter')),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
var __VLS_3;
const __VLS_8 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    label: (__VLS_ctx.t('system.siteLogo')),
    prop: "site_logo",
}));
const __VLS_10 = __VLS_9({
    label: (__VLS_ctx.t('system.siteLogo')),
    prop: "site_logo",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "logo-upload-container" },
});
if (__VLS_ctx.form.site_logo) {
    const __VLS_12 = {}.ElImage;
    /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        src: (__VLS_ctx.formatUrl(__VLS_ctx.form.site_logo)),
        ...{ style: {} },
        previewTeleported: (true),
    }));
    const __VLS_14 = __VLS_13({
        src: (__VLS_ctx.formatUrl(__VLS_ctx.form.site_logo)),
        ...{ style: {} },
        previewTeleported: (true),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
const __VLS_16 = {}.ElUpload;
/** @type {[typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    action: "#",
    autoUpload: (false),
    onChange: (__VLS_ctx.handleLogoSelect),
    accept: "image/*",
}));
const __VLS_18 = __VLS_17({
    action: "#",
    autoUpload: (false),
    onChange: (__VLS_ctx.handleLogoSelect),
    accept: "image/*",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
const __VLS_20 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    type: "primary",
}));
const __VLS_22 = __VLS_21({
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
(__VLS_ctx.t('app.pleaseSelectImageFile'));
var __VLS_23;
var __VLS_19;
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ style: {} },
});
(__VLS_ctx.t('common.uploadImageTip'));
var __VLS_11;
const __VLS_24 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    ...{ 'onClose': {} },
    modelValue: (__VLS_ctx.showCropDialog),
    title: (__VLS_ctx.t('common.cropImage') || 'Crop Image'),
    width: "700px",
    center: true,
    destroyOnClose: true,
}));
const __VLS_26 = __VLS_25({
    ...{ 'onClose': {} },
    modelValue: (__VLS_ctx.showCropDialog),
    title: (__VLS_ctx.t('common.cropImage') || 'Crop Image'),
    width: "700px",
    center: true,
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
let __VLS_28;
let __VLS_29;
let __VLS_30;
const __VLS_31 = {
    onClose: (__VLS_ctx.resetCrop)
};
__VLS_27.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ref: "cropperStageRef",
    ...{ class: "cropper-stage" },
});
/** @type {typeof __VLS_ctx.cropperStageRef} */ ;
if (__VLS_ctx.previewImageUrl) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        ...{ onLoad: (__VLS_ctx.initCropper) },
        ref: "imageRef",
        src: (__VLS_ctx.previewImageUrl),
        alt: "Crop preview",
    });
    /** @type {typeof __VLS_ctx.imageRef} */ ;
}
{
    const { footer: __VLS_thisSlot } = __VLS_27.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dialog-footer" },
    });
    const __VLS_32 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        ...{ 'onClick': {} },
    }));
    const __VLS_34 = __VLS_33({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    let __VLS_36;
    let __VLS_37;
    let __VLS_38;
    const __VLS_39 = {
        onClick: (__VLS_ctx.resetCrop)
    };
    __VLS_35.slots.default;
    (__VLS_ctx.t('common.cancel') || 'Cancel');
    var __VLS_35;
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.uploading),
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.uploading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (__VLS_ctx.confirmCrop)
    };
    __VLS_43.slots.default;
    (__VLS_ctx.t('common.cropAndUpload') || 'Crop & Upload');
    var __VLS_43;
}
var __VLS_27;
/** @type {__VLS_StyleScopedClasses['system-core-config']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-upload-container']} */ ;
/** @type {__VLS_StyleScopedClasses['cropper-stage']} */ ;
/** @type {__VLS_StyleScopedClasses['dialog-footer']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            uploading: uploading,
            showCropDialog: showCropDialog,
            previewImageUrl: previewImageUrl,
            imageRef: imageRef,
            cropperStageRef: cropperStageRef,
            form: form,
            resetCrop: resetCrop,
            initCropper: initCropper,
            handleLogoSelect: handleLogoSelect,
            confirmCrop: confirmCrop,
            formatUrl: formatUrl,
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
