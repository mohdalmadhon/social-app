import { reactive } from 'vue';

export const notifications = reactive([]);

// post dialog opened from a toast (mounted once, in NotificationContainer)
export const postDialog = reactive({
    show: false,
    postId: null
});

export function openPostDialog(postId) {
    if (!postId) {
        return;
    }

    postDialog.postId = Number(postId);
    postDialog.show = true;
}

export function closePostDialog() {
    postDialog.show = false;
    postDialog.postId = null;
}

// returns '' for posts without an image and for videos
export function postImageUrl(imagePath) {
    if (!imagePath) {
        return '';
    }

    const path = imagePath.toLowerCase();

    if (
        path.endsWith('.mp4') ||
        path.endsWith('.webm') ||
        path.endsWith('.mov') ||
        path.endsWith('.avi')
    ) {
        return '';
    }

    return `/uploads/${imagePath}`;
}

export function avatarUrl(avatarPath) {
    return avatarPath ? `/uploads/${avatarPath}` : '';
}

let nextNotificationId = 0;

// options:
//   avatar   - url of the sender's avatar
//   initial  - letter shown when the sender has no avatar
//   image    - post image thumbnail
//   postId   - clicking opens the post dialog
//   route    - clicking navigates there (vue-router location)
export function addNotification(message, type = 'success', options = {}) {
    const id = ++nextNotificationId;

    const clickable = !!(options.postId || options.route);

    notifications.push({
        id,
        message,
        type,
        avatar: options.avatar || '',
        initial: options.initial || '',
        image: options.image || '',
        postId: options.postId || null,
        route: options.route || null
    });

    setTimeout(() => {
        removeNotification(id);
    }, clickable ? 6000 : 3000);
}

export function removeNotification(id) {
    const index = notifications.findIndex(
        notification => notification.id === id
    );

    if (index !== -1) {
        notifications.splice(index, 1);
    }
}