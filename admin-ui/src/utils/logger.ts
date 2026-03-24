/**
 * 日志工具
 */

import { ENV } from '@/config/environment'

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const LogLevelMap: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
}

class Logger {
  private enableLog = ENV.ENABLE_LOG

  private log(level: LogLevel, message: string, data?: any) {
    if (!this.enableLog) return

    const timestamp = new Date().toLocaleTimeString()
    const prefix = `[${timestamp}] [${level.toUpperCase()}]`

    const styles = {
      debug: 'color: #0ea5e9; font-weight: bold;',
      info: 'color: #10b981; font-weight: bold;',
      warn: 'color: #f59e0b; font-weight: bold;',
      error: 'color: #ef4444; font-weight: bold;',
    }

    if (data !== undefined) {
      console.log(`%c${prefix} ${message}`, styles[level], data)
    } else {
      console.log(`%c${prefix} ${message}`, styles[level])
    }
  }

  debug(message: string, data?: any) {
    this.log('debug', message, data)
  }

  info(message: string, data?: any) {
    this.log('info', message, data)
  }

  warn(message: string, data?: any) {
    this.log('warn', message, data)
  }

  error(message: string, data?: any) {
    this.log('error', message, data)
  }
}

export const logger = new Logger()
