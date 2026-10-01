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
    const resp = await fetch(`/api/groups/search?offset=${offset}&private=1&search=${searchValue}`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || "coud not get data")
    }

    console.log(result)
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

    console.log(result)
    return result;
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