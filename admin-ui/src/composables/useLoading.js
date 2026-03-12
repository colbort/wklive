/**
 * 加载状态 Hook
 */
import { ref } from 'vue';
export function useLoading(initialState = false) {
    const loading = ref(initialState);
    const setLoading = (value) => {
        loading.value = value;
    };
    const startLoading = () => {
        loading.value = true;
    };
    const stopLoading = () => {
        loading.value = false;
    };
    const withLoading = async (callback) => {
        startLoading();
        try {
            return await callback();
        }
        finally {
            stopLoading();
        }
    };
    return {
        loading,
        setLoading,
        startLoading,
        stopLoading,
        withLoading,
    };
}
