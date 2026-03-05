/**
 * Structured frontend logger that sends log entries to the Go backend
 * via the Wails FrontendLog binding.
 *
 * Each log entry includes browser context (current URL, user agent)
 * and is forwarded to slog on the backend as structured JSON.
 *
 * Usage:
 *   import { logger } from '$lib/utils/logger';
 *   logger.info('email loaded', { filePath: '/tmp/test.eml' });
 *   logger.error('failed to parse', { error: err.message });
 */

import { FrontendLog } from '$lib/wailsjs/go/main/App';

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

interface LogContext {
  [key: string]: unknown;
}

function getBrowserContext(): LogContext {
  const ctx: LogContext = {};
  if (typeof window !== 'undefined') {
    ctx.url = window.location.pathname;
    ctx.userAgent = navigator.userAgent;
  }
  return ctx;
}

function safeStringify(obj: unknown): string {
  try {
    if (typeof obj === 'object' && obj !== null) {
      return JSON.stringify(obj);
    }
    return String(obj);
  } catch {
    return '[Circular/Error]';
  }
}

function send(level: LogLevel, message: string, extra?: LogContext): void {
  try {
    if (typeof FrontendLog !== 'function') return;

    const context: LogContext = {
      ...getBrowserContext(),
      ...extra,
    };

    FrontendLog(level, message, JSON.stringify(context)).catch(() => {});
  } catch {
    // Silently ignore — logging must never break the app
  }
}

export const logger = {
  debug(message: string, context?: LogContext): void {
    send('DEBUG', message, context);
  },

  info(message: string, context?: LogContext): void {
    send('INFO', message, context);
  },

  warn(message: string, context?: LogContext): void {
    send('WARN', message, context);
  },

  error(message: string, context?: LogContext): void {
    send('ERROR', message, context);
  },
};
