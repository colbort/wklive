/**
 * 格式化日期
 * @param timestamp 时间戳（秒）
 * @returns 格式化后的日期字符串
 */
export function formatDate(timestamp) {
    if (!timestamp)
        return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
}
/**
 * 格式化日期（日期部分）
 * @param timestamp 时间戳（秒）
 * @returns 格式化后的日期字符串
 */
export function formatDateOnly(timestamp) {
    if (!timestamp)
        return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString();
}
/**
 * 格式化时间（时间部分）
 * @param timestamp 时间戳（秒）
 * @returns 格式化后的时间字符串
 */
export function formatTimeOnly(timestamp) {
    if (!timestamp)
        return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString();
}
