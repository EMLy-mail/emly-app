import type { RequestHandler } from './$types';
import { presenceMap, sseClients, broadcastPresence } from './state';

export const GET: RequestHandler = async ({ locals }) => {
	if (!locals.user) {
		return new Response('Unauthorized', { status: 401 });
	}

	const userId = locals.user.id;

	const stream = new ReadableStream({
		start(controller) {
			const clientId = `${userId}-${Date.now()}`;

			sseClients.set(clientId, controller);

			// Send current state immediately
			const users = Array.from(presenceMap.values()).filter(
				(u) => Date.now() - u.lastSeen < 60000
			);
			controller.enqueue(`data: ${JSON.stringify(users)}\n\n`);

			// Cleanup on close
			const cleanup = () => {
				sseClients.delete(clientId);
				presenceMap.delete(userId);
				broadcastPresence();
			};

			// Use a heartbeat to detect disconnection
			const keepAlive = setInterval(() => {
				try {
					controller.enqueue(': keepalive\n\n');
				} catch {
					cleanup();
					clearInterval(keepAlive);
				}
			}, 30000);
		}
	});

	return new Response(stream, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			Connection: 'keep-alive'
		}
	});
};
