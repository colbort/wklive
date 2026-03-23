import { useI18n } from 'vue-i18n';
export default ((__VLS_props, __VLS_ctx, __VLS_expose, __VLS_setup = (async () => {
    const { t } = useI18n();
    const __VLS_props = withDefaults(defineProps(), {
        showActions: true,
        pageSizes: () => [10, 20, 50, 100],
    });
    debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
    const __VLS_withDefaultsArg = (function (t) { return t; })({
        showActions: true,
        pageSizes: () => [10, 20, 50, 100],
    });
    const __VLS_fnComponent = (await import('vue')).defineComponent({});
    const __VLS_ctx = {};
    let __VLS_components;
    let __VLS_directives;
    // CSS variable injection 
    // CSS variable injection end 
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "table-wrapper" },
    });
    const __VLS_0 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        data: (__VLS_ctx.data),
        stripe: true,
        border: true,
        ...{ class: "table-content" },
    }));
    const __VLS_2 = __VLS_1({
        data: (__VLS_ctx.data),
        stripe: true,
        border: true,
        ...{ class: "table-content" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    var __VLS_4 = {};
    if (__VLS_ctx.showActions) {
        const __VLS_6 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_7 = __VLS_asFunctionalComponent(__VLS_6, new __VLS_6({
            prop: "actions",
            label: "Actions",
            width: "150",
            align: "center",
            fixed: "right",
        }));
        const __VLS_8 = __VLS_7({
            prop: "actions",
            label: "Actions",
            width: "150",
            align: "center",
            fixed: "right",
        }, ...__VLS_functionalComponentArgsRest(__VLS_7));
        __VLS_9.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_9.slots;
            const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
            var __VLS_10 = {
                row: (row),
            };
        }
        var __VLS_9;
    }
    var __VLS_3;
    if (__VLS_ctx.pagination) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.t('common.totalItems', { count: __VLS_ctx.pagination.total }));
        const __VLS_12 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
            ...{ 'onClick': {} },
            disabled: (!__VLS_ctx.pagination.hasPrev),
        }));
        const __VLS_14 = __VLS_13({
            ...{ 'onClick': {} },
            disabled: (!__VLS_ctx.pagination.hasPrev),
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
        let __VLS_16;
        let __VLS_17;
        let __VLS_18;
        const __VLS_19 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.pagination))
                    return;
                __VLS_ctx.$emit('prev');
            }
        };
        __VLS_15.slots.default;
        (__VLS_ctx.t('common.prevPage'));
        var __VLS_15;
        const __VLS_20 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
            ...{ 'onClick': {} },
            disabled: (!__VLS_ctx.pagination.hasNext),
        }));
        const __VLS_22 = __VLS_21({
            ...{ 'onClick': {} },
            disabled: (!__VLS_ctx.pagination.hasNext),
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
        let __VLS_24;
        let __VLS_25;
        let __VLS_26;
        const __VLS_27 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.pagination))
                    return;
                __VLS_ctx.$emit('next');
            }
        };
        __VLS_23.slots.default;
        (__VLS_ctx.t('common.nextPage'));
        var __VLS_23;
        const __VLS_28 = {}.ElSelect;
        /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
        // @ts-ignore
        const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
            ...{ 'onChange': {} },
            modelValue: (__VLS_ctx.pagination.limit),
            ...{ style: {} },
        }));
        const __VLS_30 = __VLS_29({
            ...{ 'onChange': {} },
            modelValue: (__VLS_ctx.pagination.limit),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_29));
        let __VLS_32;
        let __VLS_33;
        let __VLS_34;
        const __VLS_35 = {
            onChange: (() => __VLS_ctx.$emit('change-limit'))
        };
        __VLS_31.slots.default;
        const __VLS_36 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
            label: "10",
            value: (10),
        }));
        const __VLS_38 = __VLS_37({
            label: "10",
            value: (10),
        }, ...__VLS_functionalComponentArgsRest(__VLS_37));
        const __VLS_40 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
            label: "20",
            value: (20),
        }));
        const __VLS_42 = __VLS_41({
            label: "20",
            value: (20),
        }, ...__VLS_functionalComponentArgsRest(__VLS_41));
        const __VLS_44 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
            label: "50",
            value: (50),
        }));
        const __VLS_46 = __VLS_45({
            label: "50",
            value: (50),
        }, ...__VLS_functionalComponentArgsRest(__VLS_45));
        var __VLS_31;
    }
    /** @type {__VLS_StyleScopedClasses['table-wrapper']} */ ;
    /** @type {__VLS_StyleScopedClasses['table-content']} */ ;
    // @ts-ignore
    var __VLS_5 = __VLS_4, __VLS_11 = __VLS_10;
    [__VLS_dollars.$attrs,];
    var __VLS_dollars;
    const __VLS_self = (await import('vue')).defineComponent({
        setup() {
            return {
                t: t,
            };
        },
        __typeProps: {},
        props: {},
    });
    return {};
})()) => ({})); /* PartiallyEnd: #4569/main.vue */
