// Logs EMLyUpdater named-pipe IPC round trips straight to the DevTools
// console, distinct from LogDebug/LogInfo (which go to window.runtime and
// end up in the Go-side app log, not the browser console the user actually
// has open while diagnosing IPC issues).
import type { updateripc } from "$lib/wailsjs/go/models";

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

// Anything carrying the IPCMeta the Go client attaches to every updater IPC
// result — SystemInfo, ADStatus, UpdaterIPCStatus. Structural rather than a
// union of the generated classes so a new Meta-carrying binding needs no
// edit here.
type HasIPCMeta = { Meta?: updateripc.IPCMeta };

// Prints the per-frame trail of a single v2 exchange (dial, ClientHello /
// ServerAnswHello, semver check, auth challenge / HMAC response, payload) as
// a collapsed console group with one table row per frame. Each of these is a
// separate round trip the Go client used to swallow into one opaque
// success/failure — this is the DevTools-side view of the same steps
// backend/utils/updateripc/trace.go writes to the app log.
//
// Safe by construction: the Go side puts only sizes in HandshakeStep.Detail,
// never the shared secret, the auth nonce, or the computed HMAC.
export function logIPCHandshake(method: string, result: HasIPCMeta | undefined) {
    const meta = result?.Meta;
    if (!meta) return;

    const steps = meta.Steps ?? [];
    const failed = steps.some((step) => !!step.Error);
    const label =
        `${PREFIX} handshake ${method} — ` +
        `protocol v${meta.RequestProtocolVersion}->v${meta.ResponseProtocolVersion || "?"}, ` +
        `${steps.length} frame(s)${failed ? ", FAILED" : ""}`;

    // Failures stay expanded: a broken handshake is the one case where the
    // trail is the point, not background noise.
    if (failed) {
        console.group(label);
    } else {
        console.groupCollapsed(label);
    }
    try {
        console.table(
            steps.map((step) => ({
                "#": step.Seq,
                phase: step.Phase,
                dir: step.Direction,
                frame: step.Frame,
                detail: step.Detail,
                ms: step.ElapsedMs,
                error: step.Error,
            })),
        );
        if (meta.ErrorCode || meta.ErrorMessage) {
            console.warn(
                `${PREFIX} updater reported ${meta.ErrorCode}: ${meta.ErrorMessage}`,
            );
        }
        if (meta.ResponseSenderVersion) {
            console.log(
                `${PREFIX} EMLy ${meta.RequestSenderVersion} <-> EMLyUpdater ${meta.ResponseSenderVersion}`,
            );
        }
    } finally {
        console.groupEnd();
    }
}
