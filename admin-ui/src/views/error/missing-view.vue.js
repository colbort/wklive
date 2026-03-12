import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
const route = useRoute();
const { t } = useI18n();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElResult;
/** @type {[typeof __VLS_components.ElResult, typeof __VLS_components.elResult, typeof __VLS_components.ElResult, typeof __VLS_components.elResult, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    icon: "error",
    title: (__VLS_ctx.t('error.missingComponent')),
    subTitle: (__VLS_ctx.t('error.missingComponentSubtitle')),
}));
const __VLS_2 = __VLS_1({
    icon: "error",
    title: (__VLS_ctx.t('error.missingComponent')),
    subTitle: (__VLS_ctx.t('error.missingComponentSubtitle')),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
{
    const { extra: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    (__VLS_ctx.route.meta.missing);
    const __VLS_5 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
        ...{ 'onClick': {} },
        ...{ style: {} },
    }));
    const __VLS_7 = __VLS_6({
        ...{ 'onClick': {} },
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    let __VLS_9;
    let __VLS_10;
    let __VLS_11;
    const __VLS_12 = {
        onClick: (...[$event]) => {
            __VLS_ctx.$router.push('/');
        }
    };
    __VLS_8.slots.default;
    (__VLS_ctx.t('common.goHome'));
    var __VLS_8;
}
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            route: route,
            t: t,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
