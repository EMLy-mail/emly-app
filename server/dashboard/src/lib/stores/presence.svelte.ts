import { browser } from '$app/environment';
import { page } from '$app/stores';

export interface ActiveUser {
	userId: string;
	username: string;
	displayname: string;
	currentPath: string;
	reportId: number | null;
	lastSeen: number;
}

class PresenceStore {
	activeUsers = $state<ActiveUser[]>([]);
	connected = $state(false);

	private eventSource: EventSource | null = null;
	private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
	private currentPath = '/';
	private unsubscribePage: (() => void) | null = null;

	connect() {
		if (!browser || this.eventSource) return;

		this.eventSource = new EventSource('/api/presence');

		this.eventSource.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				this.activeUsers = data;
			} catch {
				// ignore parse errors
			}
		};

		this.eventSource.onopen = () => {
			this.connected = true;
		};

		this.eventSource.onerror = () => {
			this.connected = false;
			// EventSource auto-reconnects
		};

		// Track current page and send heartbeats
		this.unsubscribePage = page.subscribe((p) => {
			this.currentPath = p.url.pathname;
		});

		// Send heartbeat every 15 seconds
		this.sendHeartbeat();
		this.heartbeatInterval = setInterval(() => this.sendHeartbeat(), 15000);
	}

	disconnect() {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
		if (this.heartbeatInterval) {
			clearInterval(this.heartbeatInterval);
			this.heartbeatInterval = null;
		}
		if (this.unsubscribePage) {
			this.unsubscribePage();
			this.unsubscribePage = null;
		}
		this.connected = false;
		this.activeUsers = [];
	}

	private async sendHeartbeat() {
		try {
			const reportMatch = this.currentPath.match(/^\/reports\/(\d+)/);
			const reportId = reportMatch ? Number(reportMatch[1]) : null;

			await fetch('/api/presence/heartbeat', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					currentPath: this.currentPath,
					reportId
				})
			});
		} catch {
			// ignore heartbeat failures
		}
	}

	getViewersForReport(reportId: number): ActiveUser[] {
		return this.activeUsers.filter((u) => u.reportId === reportId);
	}
}

export const presence = new PresenceStore();
