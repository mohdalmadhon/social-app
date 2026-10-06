import { ref } from 'vue';

import { getUnreadNotificationCount } from '@/api/common/notifications';
import { addNotification } from './notifications';
import { messageSent } from './chatState';

export const unreadNotificationCount = ref(0);

let countVersion = 0;

export async function refreshUnreadNotificationCount() {
    const version = ++countVersion;

    try {
        const result = await getUnreadNotificationCount();

        if (version !== countVersion) {
            return;
        }

        unreadNotificationCount.value = Math.max(0, Number(result.count) || 0);
    } catch (err) {
        console.error(err);
    }
}

export function setUnreadNotificationCount(count) {
    countVersion++;
    unreadNotificationCount.value = Math.max(0, Number(count) || 0);
}

export function incrementUnreadNotificationCount() {
    countVersion++;
    unreadNotificationCount.value += 1;
}

export function clearUnreadNotificationCount() {
    countVersion++;
    unreadNotificationCount.value = 0;
}

export function handleIncomingNotification(data) {
    if (data && data.error) {
        messageSent.value = false;
        addNotification(data.message, 'error');
    }
}
