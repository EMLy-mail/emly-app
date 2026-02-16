import type { ActiveUser } from '$lib/stores/presence.svelte';

// In-memory presence tracking - shared between SSE and heartbeat endpoints
export const presenceMap = new Map<string, ActiveUser>();

// SSE client connections
export const sseClients = new Map<string, ReadableStreamDefaultController>();

export function broadcastPresence() {
	const users = Array.from(presenceMap.values()).filter((u) => Date.now() - u.lastSeen < 60000);
	const data = `data: ${JSON.stringify(users)}\n\n`;

	for (const [clientId, controller] of sseClients.entries()) {
		try {
			controller.enqueue(data);
		} catch {
			sseClients.delete(clientId);
		}
	}
}
