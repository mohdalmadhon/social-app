export const LIMITS = Object.freeze({
    identifier: 75,
    identifierMin: 3,
    password: 75,
    passwordMin: 8,
    name: 15,
    nameMin: 2,
    username: 12,
    usernameMin: 3,
    email: 75,
    emailMin: 5,
    about: 1000,
    aboutField: 200,
    code: 6,
    postContent: 1000,
    postLocation: 200,
    postTags: 50,
    comment: 200,
    commentLines: 10,
    chatMessage: 1000,
    chatLines: 15,
    groupTitle: 15,
    groupDescription: 200,
    eventTitle: 100,
    eventDescription: 1000,
    search: 100,
    deleteConfirm: 6,
    avatarSize: 5 * 1024 * 1024,
    postMediaSize: 50 * 1024 * 1024,
    idList: 100
});

export function charCount(value) {
    return Array.from(String(value ?? '')).length;
}

export function normalizeNewlines(value) {
    return String(value ?? '').replace(/\r\n?/g, '\n');
}

export function cleanText(value) {
    return normalizeNewlines(value).trim();
}

export function lineCount(value) {
    const text = normalizeNewlines(value);

    return text === '' ? 0 : text.split('\n').length;
}

export function limitLines(value, maxLines) {
    const lines = normalizeNewlines(value).split('\n');

    if (lines.length <= maxLines) {
        return normalizeNewlines(value);
    }

    return lines.slice(0, maxLines).join('\n');
}

export function clampText(value, max) {
    const chars = Array.from(String(value ?? ''));

    return chars.length > max ? chars.slice(0, max).join('') : chars.join('');
}

export function clampMultiline(value, max, maxLines) {
    return clampText(limitLines(value, maxLines), max);
}

export function hasControlChars(value) {
    return /[\u0000-\u0008\u000b-\u001f\u007f]/.test(String(value ?? ''));
}

export function validateText(label, value, { min = 0, max, maxLines = 0, singleLine = false } = {}) {
    const text = String(value ?? '');

    if (hasControlChars(text)) {
        return `${label} contains invalid characters`;
    }

    if (singleLine && /[\r\n]/.test(text)) {
        return `${label} cannot contain new lines`;
    }

    const length = charCount(text);

    if (min > 0 && length < min) {
        return min === 1
            ? `${label} cannot be empty`
            : `${label} must be at least ${min} characters`;
    }

    if (length > max) {
        return `${label} cannot be more than ${max} characters`;
    }

    if (maxLines > 0 && lineCount(text) > maxLines) {
        return `${label} cannot be more than ${maxLines} lines`;
    }

    return '';
}

export function validateSearch(value) {
    return validateText('Search', String(value ?? '').trim(), {
        max: LIMITS.search,
        singleLine: true
    });
}

export function validateChatMessage(value) {
    return validateText('Message', cleanText(value), {
        min: 1,
        max: LIMITS.chatMessage,
        maxLines: LIMITS.chatLines
    });
}

export function validateCommentText(value, hasImage = false) {
    const content = cleanText(value);

    if (!content && hasImage) {
        return '';
    }

    return validateText('Comment', content, {
        min: 1,
        max: LIMITS.comment,
        maxLines: LIMITS.commentLines
    });
}

export function validateGroupTitle(value) {
    return validateText('Group name', String(value ?? '').trim(), {
        min: 1,
        max: LIMITS.groupTitle,
        singleLine: true
    });
}

export function validateGroupDescription(value) {
    return validateText('Description', cleanText(value), {
        min: 1,
        max: LIMITS.groupDescription
    });
}

export function validateEventTitle(value) {
    return validateText('Title', String(value ?? '').trim(), {
        min: 1,
        max: LIMITS.eventTitle,
        singleLine: true
    });
}

export function validateEventDescription(value) {
    return validateText('Description', cleanText(value), {
        max: LIMITS.eventDescription
    });
}

export function validateImageFile(file, maxSize = LIMITS.avatarSize, types = ['image/jpeg', 'image/png', 'image/gif']) {
    if (!file) {
        return '';
    }

    if (file.size <= 0) {
        return 'File is empty';
    }

    if (!types.includes(file.type)) {
        return 'Unsupported file type';
    }

    if (file.size > maxSize) {
        return `File must be smaller than ${Math.round(maxSize / (1024 * 1024))}MB`;
    }

    return '';
}

export function limitIdList(ids) {
    return Array.isArray(ids) && ids.length <= LIMITS.idList;
}

export function safeSearch(value) {
    return clampText(String(value ?? '').replace(/[\r\n]+/g, ' ').trim(), LIMITS.search);
}
