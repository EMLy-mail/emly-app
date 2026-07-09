// Logs EMLyUpdater named-pipe IPC round trips straight to the DevTools
// console, distinct from LogDebug/LogInfo (which go to window.runtime and
// end up in the Go-side app log, not the browser console the user actually
// has open while diagnosing IPC issues).
const PREFIX = "[EMLy IPC]";

export function logIPCRequest(method: string) {
    console.log(`${PREFIX} -> ${method}`);
}

export function logIPCResponse(method: string, response: unknown) {
    console.log(`${PREFIX} <- ${method}`, response);
}

export function logIPCError(method: string, error: unknown) {
    console.error(`${PREFIX} x ${method}`, error);
}
