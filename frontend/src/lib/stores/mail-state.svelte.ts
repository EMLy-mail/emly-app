import type { internal } from "$lib/wailsjs/go/models";

export interface EmailTab {
    id: string;
    email: internal.EmailData;
}

class MailState {
    tabs = $state<EmailTab[]>([]);
    activeTabId = $state<string | null>(null);

    get currentEmail(): internal.EmailData | null {
        if (this.tabs.length === 0 || !this.activeTabId) return null;
        return this.tabs.find(t => t.id === this.activeTabId)?.email ?? null;
    }

    setParams(email: internal.EmailData | null) {
        if (!email) {
            this.clear();
            return;
        }
        const id = crypto.randomUUID();
        this.tabs = [{ id, email }];
        this.activeTabId = id;
    }

    addTab(email: internal.EmailData): string {
        const id = crypto.randomUUID();
        this.tabs = [...this.tabs, { id, email }];
        this.activeTabId = id;
        return id;
    }

    updateTabEmail(tabId: string, email: internal.EmailData) {
        this.tabs = this.tabs.map(t => t.id === tabId ? { ...t, email } : t);
    }

    removeTab(id: string) {
        const idx = this.tabs.findIndex(t => t.id === id);
        if (idx === -1) return;
        const newTabs = this.tabs.filter(t => t.id !== id);
        this.tabs = newTabs;
        if (this.activeTabId === id) {
            this.activeTabId = newTabs[Math.max(0, idx - 1)]?.id ?? newTabs[0]?.id ?? null;
        }
    }

    setActiveTab(id: string) {
        this.activeTabId = id;
    }

    clear() {
        this.tabs = [];
        this.activeTabId = null;
    }

    getAllTabs() {
        return this.tabs;
    }

    getActiveTabId() {
        return this.activeTabId;
    }
}

export const mailState = new MailState();
