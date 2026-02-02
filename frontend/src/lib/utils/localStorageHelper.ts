import { GetConfig, SaveConfig } from "$lib/wailsjs/go/main/App";
import type { utils } from "$lib/wailsjs/go/models" 

async function loadConfig() {
    try {
        const config = await GetConfig();
        return config;
    } catch (error) {
        console.error("Failed to load config:", error);
        return null;
    }
}

async function saveConfig(config: utils.Config) {
    try {
        await SaveConfig(config);
    } catch (error) {
        console.error("Failed to save config:", error);
    }
}

function saveToLocalStorage(key: string, value: string): boolean {
    try {
        localStorage.setItem(key, value);
        return true;
    } catch (error) {
        console.error("Failed to save to localStorage:", error);    
        return false;
    }
}

function getFromLocalStorage(key: string): string | null {
    try {
        return localStorage.getItem(key);
    } catch (error) {
        console.error("Failed to get from localStorage:", error);
        return null;
    }
}

export {
    loadConfig,
    saveConfig,
    saveToLocalStorage,
    getFromLocalStorage
};

