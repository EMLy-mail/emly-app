/**
 * Fire-and-forget checkpoints into the startup trace file (see
 * backend/logger/trace.go, startup-trace.log next to app.log). Kept
 * separate from the regular app log so a slow "Inizializzazione..." /
 * "Initialization..." screen can be diagnosed step-by-step - which exact
 * step ate the time, e.g. parsing a heavy .msg vs. sanitizing a huge HTML
 * body - instead of guessed at.
 *
 * Never awaited by callers: tracing must not itself slow down the path it's
 * measuring. Failures are swallowed - a missing binding shouldn't break
 * app startup.
 */
import { TraceStartupStep } from '$lib/wailsjs/go/main/App';

export function trace(step: string, detail?: string | number): void {
  try {
    void TraceStartupStep(step, detail === undefined ? '' : String(detail));
  } catch {
    // best-effort only - never let tracing itself break startup
  }
}
