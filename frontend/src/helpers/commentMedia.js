import { ref, onBeforeUnmount } from 'vue';
import { CHAT_MEDIA_ACCEPT, validateChatMedia } from '@/helpers/chatMedia';

export const COMMENT_MEDIA_ACCEPT = CHAT_MEDIA_ACCEPT;

export function commentImageSrc(path) {
    if (!path) {
        return '';
    }

    return path.startsWith('blob:') ? path : `/uploads/${path}`;
}

export function buildCommentRequest(postId, content, replyTo = null, image = null) {
    if (!image) {
        return {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                postId,
                content,
                replyTo
            })
        };
    }

    const form = new FormData();

    form.append('postId', String(postId));
    form.append('content', content);

    if (replyTo !== null) {
        form.append('replyTo', String(replyTo));
    }

    form.append('image', image);

    return {
        method: 'POST',
        credentials: 'include',
        body: form
    };
}

export function useCommentMedia() {
    const file = ref(null);
    const previewUrl = ref('');
    const error = ref('');
    const fileInput = ref(null);

    function release() {
        if (previewUrl.value) {
            URL.revokeObjectURL(previewUrl.value);
        }

        file.value = null;
        previewUrl.value = '';
    }

    function clearMedia() {
        release();
        error.value = '';
    }

    function openPicker() {
        fileInput.value?.click();
    }

    function onSelect(event) {
        const selected = event.target.files?.[0];

        event.target.value = '';

        if (!selected) {
            return;
        }

        const message = validateChatMedia(selected);

        if (message) {
            error.value = message;
            return;
        }

        release();
        error.value = '';
        file.value = selected;
        previewUrl.value = URL.createObjectURL(selected);
    }

    function takeMedia() {
        if (!file.value) {
            return null;
        }

        const media = {
            file: file.value,
            previewUrl: previewUrl.value
        };

        file.value = null;
        previewUrl.value = '';
        error.value = '';

        return media;
    }

    function restoreMedia(media) {
        if (!media) {
            return;
        }

        release();
        file.value = media.file;
        previewUrl.value = media.previewUrl;
    }

    function revokeMedia(media) {
        if (media?.previewUrl) {
            URL.revokeObjectURL(media.previewUrl);
        }
    }

    onBeforeUnmount(release);

    return {
        file,
        previewUrl,
        error,
        fileInput,
        clearMedia,
        openPicker,
        onSelect,
        takeMedia,
        restoreMedia,
        revokeMedia
    };
}
