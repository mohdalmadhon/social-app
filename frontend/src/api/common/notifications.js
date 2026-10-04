export async function getNotifications(offset = 0) {
    const response = await fetch(`/api/notifications?offset=${offset}`, {
        credentials: 'include'
    });

    return await response.json();
}

export async function getUnreadNotificationCount() {
    const response = await fetch('/api/notifications/unread', {
        credentials: 'include'
    });

    return await response.json();
}

export async function markNotificationsRead() {
    const response = await fetch('/api/notifications/read', {
        method: 'POST',
        credentials: 'include'
    });

    return await response.json();
}

export async function respondToGroupInvite(groupID, status) {
    const response = await fetch('/api/groups/status', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status, groupID })
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'could not update invite');
    }

    return result;
}

export async function acceptFollowRequest(userID) {
    const response = await fetch(
        `/api/follow/accept?targetid=${userID}`,
        {
            method: 'POST',
            credentials: 'include'
        }
    );

    return await response.json();
}

export async function rejectFollowRequest(userID) {
    const response = await fetch(
        `/api/follow/reject?targetid=${userID}`,
        {
            method: 'POST',
            credentials: 'include'
        }
    );

    return await response.json();
}

export async function respondToJoinRequest(groupID, userID, code) {
    const response = await fetch('/api/groups/requests', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ groupID, userID, code })
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'could not update request');
    }

    return result;
}
