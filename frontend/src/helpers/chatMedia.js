export const CHAT_MEDIA_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
export const CHAT_MEDIA_ACCEPT = CHAT_MEDIA_TYPES.join(',');
export const CHAT_MEDIA_MAX_SIZE = 8 * 1024 * 1024;

export function parseChatMedia(content) {
    if (typeof content !== 'string' || !content.startsWith('{')) {
        return null;
    }

    try {
        const data = JSON.parse(content);

        if (
            data &&
            (data.type === 'image' || data.type === 'gif') &&
            typeof data.path === 'string' &&
            data.path.startsWith('chats/')
        ) {
            return {
                type: data.type,
                path: data.path
            };
        }
    } catch {
        return null;
    }

    return null;
}

export function validateChatMedia(file) {
    if (!file) {
        return 'No image selected';
    }

    if (!CHAT_MEDIA_TYPES.includes(file.type)) {
        return 'Only jpg, png, gif and webp images are allowed';
    }

    if (file.size > CHAT_MEDIA_MAX_SIZE) {
        return 'Image must be 8MB or smaller';
    }

    return '';
}
