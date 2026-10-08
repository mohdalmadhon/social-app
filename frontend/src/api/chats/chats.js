import { safeSearch } from '@/helpers/limits';
import { sendWS } from "../socket/socket";

export async function getPrivateChatsLists(offset) {
    const resp = await fetch(`/api/groups?offset=${offset}&private=1`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "coud not get data")
    }

    return result;
}

export async function searchChats(offset, searchValue) {
    const resp = await fetch(`/api/groups/search?offset=${offset}&private=1&search=${encodeURIComponent(safeSearch(searchValue))}`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "coud not get data")
    }

    return result;
}

export async function getChatSuggestions() {
    const resp = await fetch(`/api/chats/suggestions`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "could not get suggestions")
    }

    return result;
}

export async function sendMessage(data) {
    const resp = await fetch(`/api/chats`, {
        method: "POST",
        credentials: 'include',
        body: JSON.stringify(data)
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "coud not get data")
    }

    return result;
}

export async function sendChatMedia(file, { userID = -1, groupID = -1 } = {}) {
    const form = new FormData();
    form.append('file', file);
    form.append('userID', String(userID ?? -1));
    form.append('groupID', String(groupID ?? -1));

    const resp = await fetch(`/api/chats/media`, {
        method: "POST",
        credentials: 'include',
        body: form
    });

    let result = null;
    try {
        result = await resp.json();
    } catch {
        throw new Error('could not send image');
    }

    if (!resp.ok || !result.status) {
        throw new Error(result.message || 'could not send image');
    }

    return result;
}

export async function checkMessageAbility(targetID) {
    const resp = await fetch(`/api/chats/ability?targetID=${encodeURIComponent(targetID)}`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "could not check message permission")
    }

    return Boolean(result.canMessage);
}

export async function getMessages(groupID, offset = 0, userID = 0) {
    const resp = await fetch(`/api/chats?groupID=${groupID}&offset=${offset}&userID=${userID}`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || 'error hapened while sending message')
    }

    return result;
}

export async function getGroupChats(offset) {
    const resp = await fetch(`/api/groups?offset=${offset}&private=0`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "coud not get data")
    }

    return result;
}

export function sendPost(userID = 0, postID = 0, recieverID = 0) {
    if (!userID || userID <= 0) {
        return
    }

    if (!postID || postID <= 0) {
        return
    }

    if (!recieverID || recieverID <= 0) {
        return;
    }

    sendWS({
        type: "post-message",
        data: {
            postID: postID,
            userID: userID,
            recieverID: recieverID
        }
    });
}
export async function markChatRead(groupID) {
    const resp = await fetch('/api/chats/read', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ groupID: Number(groupID) })
    });

    if (!resp.ok) {
        const result = await resp.json().catch(() => ({}));
        throw new Error(result.message || 'could not mark chat as read');
    }
}
