import { computed } from 'vue'
import { useAuthStore } from '@/stores'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
const props = defineProps()
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { t, te } = useI18n()
const iconMap = ElementPlusIconsVue
const menuTree = computed(() =>
  (auth.menus || [])
    .filter((a) => a.visible !== 0 && a.status !== 0)
    .slice()
    .sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0)),
)
function labelById(id, fallback) {
  const key = `menu.${id}`
  return te(key) ? t(key) : fallback
}
function go(path) {
  if (path) router.push(path)
}
function childrenMenus(n) {
  return (n.children || [])
    .filter((x) => x.menuType === 2 && x.visible !== 0 && x.status !== 0)
    .slice()
    .sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
}
function iconComp(icon) {
  if (!icon) return iconMap.Menu
  return iconMap[icon] || iconMap.Menu
}
debugger /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {}
let __VLS_components
let __VLS_directives
// CSS variable injection
// CSS variable injection end
const __VLS_0 = {}.ElMenu
/** @type {[typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, ]} */ // @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(
  __VLS_0,
  new __VLS_0({
    ...{ class: 'aside-menu' },
    router: true,
    defaultActive: __VLS_ctx.route.path,
    collapse: props.collapsed,
    collapseTransition: false,
  }),
)
const __VLS_2 = __VLS_1(
  {
    ...{ class: 'aside-menu' },
    router: true,
    defaultActive: __VLS_ctx.route.path,
    collapse: props.collapsed,
    collapseTransition: false,
  },
  ...__VLS_functionalComponentArgsRest(__VLS_1),
)
var __VLS_4 = {}
__VLS_3.slots.default
const __VLS_5 = {}.ElMenuItem
/** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ // @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent(
  __VLS_5,
  new __VLS_5({
    ...{ onClick: {} },
    index: '/home',
  }),
)
const __VLS_7 = __VLS_6(
  {
    ...{ onClick: {} },
    index: '/home',
  },
  ...__VLS_functionalComponentArgsRest(__VLS_6),
)
let __VLS_9
let __VLS_10
let __VLS_11
const __VLS_12 = {
  onClick: (...[$event]) => {
    __VLS_ctx.go('/home')
  },
}
__VLS_8.slots.default
const __VLS_13 = {}.ElIcon
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ // @ts-ignore
const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({}))
const __VLS_15 = __VLS_14({}, ...__VLS_functionalComponentArgsRest(__VLS_14))
__VLS_16.slots.default
const __VLS_17 = __VLS_ctx.iconComp('House')
// @ts-ignore
const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({}))
const __VLS_19 = __VLS_18({}, ...__VLS_functionalComponentArgsRest(__VLS_18))
var __VLS_16
{
  const { title: __VLS_thisSlot } = __VLS_8.slots
  __VLS_ctx.t('route.home')
}
var __VLS_8
for (const [m] of __VLS_getVForSourceType(__VLS_ctx.menuTree)) {
  m.id
  if (m.menuType === 1) {
    const __VLS_21 = {}.ElSubMenu
    /** @type {[typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, ]} */ // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(
      __VLS_21,
      new __VLS_21({
        index: String(m.id),
      }),
    )
    const __VLS_23 = __VLS_22(
      {
        index: String(m.id),
      },
      ...__VLS_functionalComponentArgsRest(__VLS_22),
    )
    __VLS_24.slots.default
    {
      const { title: __VLS_thisSlot } = __VLS_24.slots
      const __VLS_25 = {}.ElIcon
      /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ // @ts-ignore
      const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({}))
      const __VLS_27 = __VLS_26({}, ...__VLS_functionalComponentArgsRest(__VLS_26))
      __VLS_28.slots.default
      const __VLS_29 = __VLS_ctx.iconComp(m.icon)
      // @ts-ignore
      const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({}))
      const __VLS_31 = __VLS_30({}, ...__VLS_functionalComponentArgsRest(__VLS_30))
      var __VLS_28
      __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({})
      __VLS_ctx.labelById(m.id, m.name)
    }
    for (const [c] of __VLS_getVForSourceType(__VLS_ctx.childrenMenus(m))) {
      const __VLS_33 = {}.ElMenuItem
      /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ // @ts-ignore
      const __VLS_34 = __VLS_asFunctionalComponent(
        __VLS_33,
        new __VLS_33({
          ...{ onClick: {} },
          key: c.id,
          index: c.path,
        }),
      )
      const __VLS_35 = __VLS_34(
        {
          ...{ onClick: {} },
          key: c.id,
          index: c.path,
        },
        ...__VLS_functionalComponentArgsRest(__VLS_34),
      )
      let __VLS_37
      let __VLS_38
      let __VLS_39
      const __VLS_40 = {
        onClick: (...[$event]) => {
          if (!(m.menuType === 1)) return
          __VLS_ctx.go(c.path)
        },
      }
      __VLS_36.slots.default
      const __VLS_41 = {}.ElIcon
      /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ // @ts-ignore
      const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({}))
      const __VLS_43 = __VLS_42({}, ...__VLS_functionalComponentArgsRest(__VLS_42))
      __VLS_44.slots.default
      const __VLS_45 = __VLS_ctx.iconComp(c.icon)
      // @ts-ignore
      const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({}))
      const __VLS_47 = __VLS_46({}, ...__VLS_functionalComponentArgsRest(__VLS_46))
      var __VLS_44
      {
        const { title: __VLS_thisSlot } = __VLS_36.slots
        __VLS_ctx.labelById(c.id, c.name)
      }
      var __VLS_36
    }
    var __VLS_24
  } else if (m.menuType === 2) {
    const __VLS_49 = {}.ElMenuItem
    /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ // @ts-ignore
    const __VLS_50 = __VLS_asFunctionalComponent(
      __VLS_49,
      new __VLS_49({
        ...{ onClick: {} },
        index: m.path,
      }),
    )
    const __VLS_51 = __VLS_50(
      {
        ...{ onClick: {} },
        index: m.path,
      },
      ...__VLS_functionalComponentArgsRest(__VLS_50),
    )
    let __VLS_53
    let __VLS_54
    let __VLS_55
    const __VLS_56 = {
      onClick: (...[$event]) => {
        if (!!(m.menuType === 1)) return
        if (!(m.menuType === 2)) return
        __VLS_ctx.go(m.path)
      },
    }
    __VLS_52.slots.default
    const __VLS_57 = {}.ElIcon
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ // @ts-ignore
    const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({}))
    const __VLS_59 = __VLS_58({}, ...__VLS_functionalComponentArgsRest(__VLS_58))
    __VLS_60.slots.default
    const __VLS_61 = __VLS_ctx.iconComp(m.icon)
    // @ts-ignore
    const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({}))
    const __VLS_63 = __VLS_62({}, ...__VLS_functionalComponentArgsRest(__VLS_62))
    var __VLS_60
    {
      const { title: __VLS_thisSlot } = __VLS_52.slots
      __VLS_ctx.labelById(m.id, m.name)
    }
    var __VLS_52
  }
}
var __VLS_3
/** @type {__VLS_StyleScopedClasses['aside-menu']} */ var __VLS_dollars
const __VLS_self = (await import('vue')).defineComponent({
  setup() {
    return {
      route: route,
      t: t,
      menuTree: menuTree,
      labelById: labelById,
      go: go,
      childrenMenus: childrenMenus,
      iconComp: iconComp,
    }
  },
  __typeProps: {},
})
export default (await import('vue')).defineComponent({
  setup() {
    return {}
  },
  __typeProps: {},
}) /* PartiallyEnd: #4569/main.vue */
