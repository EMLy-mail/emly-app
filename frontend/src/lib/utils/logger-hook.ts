/**
 * Console logger hook — intercepts console.log/warn/error/info and forwards
 * them to the Go backend via FrontendLog for unified structured logging.
 *
 * Call setupConsoleLogger() once at app startup (e.g. in root layout).
 */

import { FrontendLog } from '$lib/wailsjs/go/main/App';

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

function getBrowserContext(): string {
  try {
    return JSON.stringify({
      url: window.location.pathname,
      userAgent: navigator.userAgent,
    });
  } catch {
    return '{}';
  }
}

export function setupConsoleLogger() {
  if ((window as any).__logger_initialized__) return;
  (window as any).__logger_initialized__ = true;

  const originalLog = console.log;
  const originalWarn = console.warn;
  const originalError = console.error;
  const originalInfo = console.info;

  function logToBackend(level: string, args: unknown[]) {
    try {
      if (typeof FrontendLog !== 'function') return;

      const message = args.map((arg) => safeStringify(arg)).join(' ');
      FrontendLog(level, message, getBrowserContext()).catch(() => {});
    } catch {
      // ignore
    }
  }

  console.log = (...args) => {
    originalLog(...args);
    logToBackend('INFO', args);
  };

  console.warn = (...args) => {
    originalWarn(...args);
    logToBackend('WARN', args);
  };

  console.error = (...args) => {
    originalError(...args);
    logToBackend('ERROR', args);
  };

  console.info = (...args) => {
    originalInfo(...args);
    logToBackend('INFO', args);
  };

  originalLog('Console logger hooked to Wails backend');
}
