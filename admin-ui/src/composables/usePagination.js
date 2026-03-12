/**
 * 分页 Hook
 */
import { reactive, computed } from 'vue';
export function usePagination(initialPageSize = 10) {
    const pagination = reactive({
        page: 1,
        pageSize: initialPageSize,
        total: 0,
        totalPages: 0,
    });
    const isLastPage = computed(() => pagination.page >= pagination.totalPages);
    const updateTotal = (total) => {
        pagination.total = total;
        pagination.totalPages = Math.ceil(total / pagination.pageSize);
    };
    const reset = () => {
        pagination.page = 1;
        pagination.total = 0;
        pagination.totalPages = 0;
    };
    const nextPage = () => {
        if (!isLastPage.value) {
            pagination.page++;
        }
    };
    const prevPage = () => {
        if (pagination.page > 1) {
            pagination.page--;
        }
    };
    const goPage = (page) => {
        if (page > 0 && page <= pagination.totalPages) {
            pagination.page = page;
        }
    };
    return {
        pagination,
        isLastPage,
        updateTotal,
        reset,
        nextPage,
        prevPage,
        goPage,
    };
}
