export async function addPost(post = {}, image = null) {
    const formData = new FormData();

    formData.append('content', post.content || '');
    formData.append('allowComments', String(post.allowComments));
    formData.append('privatePost', String(post.privatePost));
    formData.append('groupID', String(post.groupID));
    formData.append('location', post.location || '');
    formData.append(
        'taggedPeople',
        JSON.stringify(post.taggedPeople || [])
    );

    if (image) {
        formData.append('image', image);
    }

    const resp = await fetch('/api/post', {
        method: 'POST',
        credentials: 'include',
        body: formData
    });

    if (!resp.ok) {
        throw new Error('could not send post');
    }

    return await resp.json();
}

export async function getUserPosts(userID = "", offset = 0, limit = 9, type = "") {
    const params = new URLSearchParams({
        offset: String(offset),
        limit: String(limit),
        targetID: String(userID)
    });

    if (type) {
        params.set('type', type);
    }

    const resp = await fetch(`/api/user/posts?${params.toString()}`, {
        method: "GET",
        credentials: 'include',
    });

    const result = await resp.json()
    if (!resp.ok) {
        throw new Error('Error: ' + (result.message || 'could not get data'))
    }

    return result;
}

export async function getHomeVideos(offset = 0, limit = 5) {
    const params = new URLSearchParams({
        type: 'videos',
        offset: String(offset),
        limit: String(limit)
    });

    const resp = await fetch(`/api/posts?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not get videos');
    }

    return result;
}

export async function deletePost(postID) {
    const resp = await fetch(`/api/post?postId=${postID}`, {
        method: 'DELETE',
        credentials: 'include'
    });

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not delete post');
    }

    return result;
}

export async function viewPost(postID) {
    const resp = await fetch('/api/posts/seen', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(postID)
    });

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'Could not mark post as seen');
    }

    return result;
}