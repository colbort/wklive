export default ((__VLS_props, __VLS_ctx, __VLS_expose, __VLS_setup = (async () => {
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
        const __VLS_12 = {}.ElPagination;
        /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
            currentPage: (__VLS_ctx.pagination.page),
            pageSize: (__VLS_ctx.pagination.pageSize),
            pageSizes: (__VLS_ctx.pageSizes),
            total: (__VLS_ctx.pagination.total),
            layout: "total, sizes, prev, pager, next, jumper",
            ...{ class: "table-pagination" },
        }));
        const __VLS_14 = __VLS_13({
            currentPage: (__VLS_ctx.pagination.page),
            pageSize: (__VLS_ctx.pagination.pageSize),
            pageSizes: (__VLS_ctx.pageSizes),
            total: (__VLS_ctx.pagination.total),
            layout: "total, sizes, prev, pager, next, jumper",
            ...{ class: "table-pagination" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    }
    /** @type {__VLS_StyleScopedClasses['table-wrapper']} */ ;
    /** @type {__VLS_StyleScopedClasses['table-content']} */ ;
    /** @type {__VLS_StyleScopedClasses['table-pagination']} */ ;
    // @ts-ignore
    var __VLS_5 = __VLS_4, __VLS_11 = __VLS_10;
    [__VLS_dollars.$attrs,];
    var __VLS_dollars;
    const __VLS_self = (await import('vue')).defineComponent({
        setup() {
            return {};
        },
        __typeProps: {},
        props: {},
    });
    return {};
})()) => ({})); /* PartiallyEnd: #4569/main.vue */
