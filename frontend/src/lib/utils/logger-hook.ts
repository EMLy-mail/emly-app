import { FrontendLog } from '$lib/wailsjs/go/main/App';

function safeStringify(obj: any): string {
    try {
        if (typeof obj === 'object' && obj !== null) {
             return JSON.stringify(obj);
        }
        return String(obj);
    } catch (e) {
        return '[Circular/Error]';
    }
}

export function setupConsoleLogger() {
    if ((window as any).__logger_initialized__) return;
    (window as any).__logger_initialized__ = true;

    const originalLog = console.log;
    const originalWarn = console.warn;
    const originalError = console.error;
    const originalInfo = console.info;

    function logToBackend(level: string, args: any[]) {
        try {
            // Avoid logging if wails runtime is not ready or function is missing
            if (typeof FrontendLog !== 'function') return;

            const message = args.map(arg => safeStringify(arg)).join(' ');
            FrontendLog(level, message).catch(() => {});
        } catch (e) {
            // ignore
        }
    }

    console.log = (...args) => {
        originalLog(...args);
        logToBackend("INFO", args);
    };

    console.warn = (...args) => {
        originalWarn(...args);
        logToBackend("WARN", args);
    };

    console.error = (...args) => {
        originalError(...args);
        logToBackend("ERROR", args);
    };

    console.info = (...args) => {
        originalInfo(...args);
        logToBackend("INFO", args);
    };

    originalLog("Console logger hooked to Wails backend");
}
