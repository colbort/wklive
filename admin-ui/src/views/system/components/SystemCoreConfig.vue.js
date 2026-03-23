import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const props = defineProps();
const emit = defineEmits();
const form = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
});
import { computed } from 'vue';
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
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
const __VLS_12 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    modelValue: (__VLS_ctx.form.site_logo),
    placeholder: (__VLS_ctx.t('system.siteLogo') || __VLS_ctx.t('common.pleaseEnter')),
}));
const __VLS_14 = __VLS_13({
    modelValue: (__VLS_ctx.form.site_logo),
    placeholder: (__VLS_ctx.t('system.siteLogo') || __VLS_ctx.t('common.pleaseEnter')),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
/** @type {__VLS_StyleScopedClasses['system-core-config']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            form: form,
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
