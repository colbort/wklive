/**
 * 分页 Hook
 */
import { reactive } from 'vue';
export function usePagination(initialLimit = 10) {
    const pagination = reactive({
        cursor: null,
        nextCursor: null,
        prevCursor: null,
        limit: initialLimit,
        total: 0,
        hasNext: false,
        hasPrev: false,
    });
    const updatePagination = (total, hasNext, hasPrev, nextCursor = null, prevCursor = null) => {
        pagination.total = total;
        pagination.hasNext = hasNext;
        pagination.hasPrev = hasPrev;
        pagination.nextCursor = nextCursor;
        pagination.prevCursor = prevCursor;
    };
    const reset = () => {
        pagination.cursor = null;
        pagination.nextCursor = null;
        pagination.prevCursor = null;
        pagination.total = 0;
        pagination.hasNext = false;
        pagination.hasPrev = false;
    };
    const nextPage = () => {
        if (pagination.hasNext && pagination.nextCursor) {
            pagination.cursor = pagination.nextCursor;
        }
    };
    const prevPage = () => {
        if (pagination.hasPrev && pagination.prevCursor) {
            pagination.cursor = pagination.prevCursor;
        }
    };
    return {
        pagination,
        updatePagination,
        reset,
        nextPage,
        prevPage,
    };
}
