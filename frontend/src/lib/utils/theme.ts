import { browser } from "$app/environment";

const THEME_KEY = "emly_theme";

export type Theme = "light" | "dark";

/**
 * Applies the theme to the document element and saves it to localStorage
 */
export function applyTheme(theme: Theme) {
	if (!browser) return;
	
	const isDark = theme === "dark";
	document.documentElement.classList.toggle("dark", isDark);
	
	try {
		localStorage.setItem(THEME_KEY, theme);
	} catch (e) {
		console.error("Failed to save theme to localStorage:", e);
	}
}

/**
 * Gets the current theme from localStorage or returns the default
 */
export function getStoredTheme(): Theme {
	if (!browser) return "light";
	
	try {
		const stored = localStorage.getItem(THEME_KEY);
		return stored === "light" || stored === "dark" ? stored : "light";
	} catch {
		return "light";
	}
}

/**
 * Toggles between light and dark theme
 */
export function toggleTheme(): Theme {
	const current = getStoredTheme();
	const newTheme: Theme = current === "dark" ? "light" : "dark";
	applyTheme(newTheme);
	return newTheme;
}
