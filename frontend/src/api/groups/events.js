import { refreshUnreadNotificationCount } from '@/data/notificationCount';

export async function createGroupEvent(event) {
    const response = await fetch('/api/group/events', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            groupId: Number(event.groupId),
            title: event.title,
            description: event.description || '',
            eventTime: event.eventTime,
            tzOffset: new Date().getTimezoneOffset()
        })
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'Could not create event');
    }

    return result.data;
}

export async function fetchGroupEvents(groupID, offset = 0, limit = 10) {
    const params = new URLSearchParams({
        groupID: String(groupID),
        offset: String(offset),
        limit: String(limit)
    });

    const response = await fetch(`/api/group/events?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'Could not load events');
    }

    return {
        events: result.data || [],
        userId: result.userId ?? null
    };
}

export async function respondGroupEvent(eventId, response) {
    const request = await fetch('/api/group/event/response', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            eventId: Number(eventId),
            response: Number(response)
        })
    });

    const result = await request.json();

    if (!request.ok || !result.status) {
        throw new Error(result.message || 'Could not save response');
    }

    refreshUnreadNotificationCount();

    return result.data;
}

export async function fetchGroupEventVotes(eventId, response, offset = 0, limit = 10) {
    const params = new URLSearchParams({
        eventID: String(eventId),
        response: String(response),
        offset: String(offset),
        limit: String(limit)
    });

    const request = await fetch(`/api/group/event/votes?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await request.json();

    if (!request.ok || !result.status) {
        throw new Error(result.message || 'Could not load votes');
    }

    return result.data || [];
}

export async function fetchGroupEvent(eventId) {
    const response = await fetch(`/api/group/event?eventID=${Number(eventId)}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'Could not load event');
    }

    return result.data;
}
