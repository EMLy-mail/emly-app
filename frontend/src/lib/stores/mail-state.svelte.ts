import type { internal } from "$lib/wailsjs/go/models";

class MailState {
    currentEmail = $state<internal.EmailData | null>(null);

    setParams(email: internal.EmailData | null) {
        this.currentEmail = email;
    }
    
    clear() {
        this.currentEmail = null;
    }
}

export const mailState = new MailState();
