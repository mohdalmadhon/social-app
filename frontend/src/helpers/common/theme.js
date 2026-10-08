const THEME_COOKIE = 'theme';
const DEFAULT_THEME = 'light';

export const THEMES = [
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
    { value: 'blue', label: 'Blue' },
    { value: 'pink', label: 'Pink' }
];

function isValidTheme(theme) {
    return THEMES.some(t => t.value === theme);
}

export function getThemeCookie() {
    const match = document.cookie.match(/(?:^|;\s*)theme=([^;]*)/);
    const value = match ? decodeURIComponent(match[1]) : null;

    return isValidTheme(value) ? value : DEFAULT_THEME;
}

export function setThemeCookie(theme) {
    if (!isValidTheme(theme)) {
        return;
    }

    const oneYear = 60 * 60 * 24 * 365;
    document.cookie = `${THEME_COOKIE}=${encodeURIComponent(theme)}; path=/; max-age=${oneYear}; SameSite=Lax`;
}

export function applyTheme(theme) {
    const value = isValidTheme(theme) ? theme : DEFAULT_THEME;
    document.documentElement.setAttribute('data-theme', value);
}

export function initTheme() {
    const theme = getThemeCookie();
    applyTheme(theme);
    return theme;
}

export function setTheme(theme) {
    setThemeCookie(theme);
    applyTheme(theme);
}
