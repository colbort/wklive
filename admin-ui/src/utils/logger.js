/**
 * 日志工具
 */
import { ENV } from '@/config/environment';
const LogLevelMap = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
};
class Logger {
    enableLog = ENV.ENABLE_LOG;
    log(level, message, data) {
        if (!this.enableLog)
            return;
        const timestamp = new Date().toLocaleTimeString();
        const prefix = `[${timestamp}] [${level.toUpperCase()}]`;
        const styles = {
            debug: 'color: #0ea5e9; font-weight: bold;',
            info: 'color: #10b981; font-weight: bold;',
            warn: 'color: #f59e0b; font-weight: bold;',
            error: 'color: #ef4444; font-weight: bold;',
        };
        if (data !== undefined) {
            console.log(`%c${prefix} ${message}`, styles[level], data);
        }
        else {
            console.log(`%c${prefix} ${message}`, styles[level]);
        }
    }
    debug(message, data) {
        this.log('debug', message, data);
    }
    info(message, data) {
        this.log('info', message, data);
    }
    warn(message, data) {
        this.log('warn', message, data);
    }
    error(message, data) {
        this.log('error', message, data);
    }
}
export const logger = new Logger();
