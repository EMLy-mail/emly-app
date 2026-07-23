import type { internal } from "$lib/wailsjs/go/models";
import { SetCurrentMailFilePath } from "$lib/wailsjs/go/main/App";

export type AppTab =
    | { id: string; type: "email"; email: internal.EmailData; filePath?: string }
    | { id: string; type: "pdf";   filename: string; base64Data: string }
    | { id: string; type: "image"; filename: string; base64Data: string };

// Keep EmailTab exported for any existing consumers
export type EmailTab = Extract<AppTab, { type: "email" }>;

class MailState {
    tabs = $state<AppTab[]>([]);
    activeTabId = $state<string | null>(null);

    get currentEmail(): internal.EmailData | null {
        if (this.tabs.length === 0 || !this.activeTabId) return null;
        const tab = this.tabs.find(t => t.id === this.activeTabId);
        return tab?.type === "email" ? tab.email : null;
    }

    // Keeps the Go-side "current mail file" (used by bug reports) pointed at
    // whichever tab is actually active, instead of drifting to a stale tab.
    private syncCurrentMailFilePath() {
        const tab = this.tabs.find(t => t.id === this.activeTabId);
        const filePath = tab?.type === "email" ? (tab.filePath ?? "") : "";
        SetCurrentMailFilePath(filePath).catch(() => {});
    }

    setParams(email: internal.EmailData | null, filePath?: string) {
        if (!email) {
            this.clear();
            return;
        }
        const id = crypto.randomUUID();
        this.tabs = [{ id, type: "email", email, filePath }];
        this.activeTabId = id;
        this.syncCurrentMailFilePath();
    }

    addTab(email: internal.EmailData, filePath?: string): string {
        const id = crypto.randomUUID();
        this.tabs = [...this.tabs, { id, type: "email", email, filePath }];
        this.activeTabId = id;
        this.syncCurrentMailFilePath();
        return id;
    }

    addPDFTab(filename: string, base64Data: string): string {
        const id = crypto.randomUUID();
        this.tabs = [...this.tabs, { id, type: "pdf", filename, base64Data }];
        this.activeTabId = id;
        this.syncCurrentMailFilePath();
        return id;
    }

    addImageTab(filename: string, base64Data: string): string {
        const id = crypto.randomUUID();
        this.tabs = [...this.tabs, { id, type: "image", filename, base64Data }];
        this.activeTabId = id;
        this.syncCurrentMailFilePath();
        return id;
    }

    updateTabEmail(tabId: string, email: internal.EmailData) {
        this.tabs = this.tabs.map(t =>
            t.id === tabId && t.type === "email" ? { ...t, email } : t
        );
    }

    removeTab(id: string) {
        const idx = this.tabs.findIndex(t => t.id === id);
        if (idx === -1) return;
        const newTabs = this.tabs.filter(t => t.id !== id);
        this.tabs = newTabs;
        if (this.activeTabId === id) {
            this.activeTabId = newTabs[Math.max(0, idx - 1)]?.id ?? newTabs[0]?.id ?? null;
            this.syncCurrentMailFilePath();
        }
    }

    setActiveTab(id: string) {
        this.activeTabId = id;
        this.syncCurrentMailFilePath();
    }

    clear() {
        this.tabs = [];
        this.activeTabId = null;
        this.syncCurrentMailFilePath();
    }

    getAllTabs() {
        return this.tabs;
    }

    getActiveTabId() {
        return this.activeTabId;
    }
}

export const mailState = new MailState();
