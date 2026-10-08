import { LIMITS, charCount } from '@/helpers/limits';

const RULES = {
    website: { label: 'Website', hosts: null, needsPath: false },
    linkedin: { label: 'LinkedIn', hosts: ['linkedin.com'], needsPath: true },
    twitter: { label: 'Twitter / X', hosts: ['twitter.com', 'x.com'], needsPath: true },
    instagram: { label: 'Instagram', hosts: ['instagram.com', 'instagr.am'], needsPath: true }
};

const SCHEME_REGEX = /^[a-z][a-z0-9+.\-]*:\/\//i;
const HOST_REGEX = /^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+([a-z]{2,24}|xn--[a-z0-9-]{2,59})$/;

export function isLinkField(field) {
    return Object.prototype.hasOwnProperty.call(RULES, field);
}

export function parseLink(field, value) {
    const rule = RULES[field];

    if (!rule) {
        return { value: String(value ?? ''), error: '' };
    }

    let text = String(value ?? '').trim();

    if (text === '') {
        return { value: '', error: '' };
    }

    if (/\s/.test(text)) {
        return { value: text, error: `${rule.label} link cannot contain spaces` };
    }

    if (!SCHEME_REGEX.test(text)) {
        text = `https://${text}`;
    }

    let url;

    try {
        url = new URL(text);
    } catch {
        return { value: text, error: `${rule.label} link is not a valid URL` };
    }

    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
        return { value: text, error: `${rule.label} link must start with http:// or https://` };
    }

    if (url.username || url.password) {
        return { value: text, error: `${rule.label} link cannot contain credentials` };
    }

    const host = url.hostname.toLowerCase();

    if (!HOST_REGEX.test(host)) {
        return { value: text, error: `${rule.label} link is not a valid URL` };
    }

    if (rule.hosts) {
        if (url.port) {
            return { value: text, error: `${rule.label} link is not a valid URL` };
        }

        const matched = rule.hosts.some(allowed => host === allowed || host.endsWith(`.${allowed}`));

        if (!matched) {
            return {
                value: text,
                error: `${rule.label} link must be a ${rule.label} address (${rule.hosts.join(' or ')})`
            };
        }

        if (rule.needsPath && url.pathname.replace(/\//g, '') === '') {
            return { value: text, error: `${rule.label} link must point to your profile` };
        }
    }

    url.hash = '';

    const normalized = url.toString();

    if (charCount(normalized) > LIMITS.aboutField) {
        return {
            value: normalized,
            error: `${rule.label} link cannot be more than ${LIMITS.aboutField} characters`
        };
    }

    return { value: normalized, error: '' };
}

export function validateLink(field, value) {
    return parseLink(field, value).error;
}

export function normalizeLink(field, value) {
    const result = parseLink(field, value);

    return result.error ? String(value ?? '').trim() : result.value;
}

export function safeHref(field, value) {
    if (!value) {
        return undefined;
    }

    const result = parseLink(field, value);

    return result.error ? undefined : result.value;
}
