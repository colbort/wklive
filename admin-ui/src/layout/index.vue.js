import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import SideMenu from './side-menu.vue';
import TopBar from './top-bar.vue';
import { getSystemCore } from '@/stores/core';
import { http } from '@/utils/request';
const collapsed = ref(false);
const systemCore = ref({
    siteName: 'Admin UI',
    siteLogo: '',
});
onMounted(async () => {
    try {
        const res = await getSystemCore();
        if (res && res.code === 200 && res.data) {
            if (res.data.siteName)
                systemCore.value.siteName = res.data.siteName;
            if (res.data.siteLogo)
                systemCore.value.siteLogo = res.data.siteLogo;
        }
    }
    catch (err) {
        console.error('getSystemCore failed', err);
    }
});
// 展开态宽度
const asideWidth = ref(240);
const MIN_W = 200;
const MAX_W = 360;
const COLLAPSED_W = 64;
const realAsideWidth = computed(() => (collapsed.value ? COLLAPSED_W : asideWidth.value));
let dragging = false;
let startX = 0;
let startW = 0;
function toggleSider() {
    collapsed.value = !collapsed.value;
}
function onDragStart(e) {
    if (collapsed.value)
        return;
    dragging = true;
    startX = e.clientX;
    startW = asideWidth.value;
    document.body.style.userSelect = 'none';
    document.addEventListener('mousemove', onDragMove);
    document.addEventListener('mouseup', onDragEnd);
}
function onDragMove(e) {
    if (!dragging)
        return;
    const dx = e.clientX - startX;
    const next = Math.min(MAX_W, Math.max(MIN_W, startW + dx));
    asideWidth.value = next;
}
function onDragEnd() {
    dragging = false;
    document.body.style.userSelect = '';
    document.removeEventListener('mousemove', onDragMove);
    document.removeEventListener('mouseup', onDragEnd);
}
onBeforeUnmount(() => {
    document.removeEventListener('mousemove', onDragMove);
    document.removeEventListener('mouseup', onDragEnd);
});
function formatUrl(url) {
    console.log('formatUrl', url);
    if (!url)
        return '';
    const fullUrl = url.startsWith('http')
        ? url
        : `${http.defaults.baseURL}${url}`;
    return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}t=${Date.now()}`;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['resizer']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "layout" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.aside, __VLS_intrinsicElements.aside)({
    ...{ class: "sider" },
    ...{ style: ({ width: __VLS_ctx.realAsideWidth + 'px' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "brand" },
});
if (__VLS_ctx.systemCore.siteLogo) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.formatUrl(__VLS_ctx.systemCore.siteLogo)),
        alt: "logo",
        ...{ class: "brand-logo" },
    });
}
if (!__VLS_ctx.collapsed) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "brand-text" },
    });
    (__VLS_ctx.systemCore.siteName);
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "brand-text" },
    });
    (__VLS_ctx.systemCore.siteName ? __VLS_ctx.systemCore.siteName.slice(0, 2).toUpperCase() : 'AI');
}
/** @type {[typeof SideMenu, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(SideMenu, new SideMenu({
    collapsed: (__VLS_ctx.collapsed),
}));
const __VLS_1 = __VLS_0({
    collapsed: (__VLS_ctx.collapsed),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
if (!__VLS_ctx.collapsed) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
        ...{ onMousedown: (__VLS_ctx.onDragStart) },
        ...{ class: "resizer" },
    });
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.main, __VLS_intrinsicElements.main)({
    ...{ class: "main" },
});
/** @type {[typeof TopBar, ]} */ ;
// @ts-ignore
const __VLS_3 = __VLS_asFunctionalComponent(TopBar, new TopBar({
    ...{ 'onToggleSider': {} },
    collapsed: (__VLS_ctx.collapsed),
}));
const __VLS_4 = __VLS_3({
    ...{ 'onToggleSider': {} },
    collapsed: (__VLS_ctx.collapsed),
}, ...__VLS_functionalComponentArgsRest(__VLS_3));
let __VLS_6;
let __VLS_7;
let __VLS_8;
const __VLS_9 = {
    onToggleSider: (__VLS_ctx.toggleSider)
};
var __VLS_5;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "content" },
});
const __VLS_10 = {}.RouterView;
/** @type {[typeof __VLS_components.RouterView, typeof __VLS_components.routerView, ]} */ ;
// @ts-ignore
const __VLS_11 = __VLS_asFunctionalComponent(__VLS_10, new __VLS_10({}));
const __VLS_12 = __VLS_11({}, ...__VLS_functionalComponentArgsRest(__VLS_11));
/** @type {__VLS_StyleScopedClasses['layout']} */ ;
/** @type {__VLS_StyleScopedClasses['sider']} */ ;
/** @type {__VLS_StyleScopedClasses['brand']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-logo']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-text']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-text']} */ ;
/** @type {__VLS_StyleScopedClasses['resizer']} */ ;
/** @type {__VLS_StyleScopedClasses['main']} */ ;
/** @type {__VLS_StyleScopedClasses['content']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            SideMenu: SideMenu,
            TopBar: TopBar,
            collapsed: collapsed,
            systemCore: systemCore,
            realAsideWidth: realAsideWidth,
            toggleSider: toggleSider,
            onDragStart: onDragStart,
            formatUrl: formatUrl,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
