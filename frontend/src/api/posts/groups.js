export async function addGroupPost(data) {
    const formData = new FormData();

    formData.append('content', data.content);
    formData.append(
        'allowComments',
        data.allowComments
    );
    formData.append(
        'groupID',
        data.groupID
    );
    formData.append(
        'location',
        data.location || ''
    );
    formData.append(
        'taggedPeople',
        JSON.stringify(data.taggedPeople || [])
    );

    if (data.image) {
        formData.append(
            'image',
            data.image
        );
    }

    const resp = await fetch(
        '/api/group/posts',
        {
            method: 'POST',
            credentials: 'include',
            body: formData
        }
    );

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(
            result.message ||
            'Could not send post'
        );
    }

    console.log(result)
    return result;
}

export async function getGroupMembers(searchValue = "", targetId, offset = 0) {
    const params = new URLSearchParams({
        search: searchValue,
        targetid: targetId,
        offset: offset.toString()
    });

    const resp = await fetch(`/api/group/users?${params.toString()}`, {
        method: "GET",
        credentials: "include"
    });
    
    if (!resp.ok) {
        throw new Error("could not fetch data");
    }

    const result = await resp.json();

    if (!result.status) {
        throw new Error("could not fetch data");
    }
    return result;
} 

export async function getGroupPost(groupID, postID) {
    const params = new URLSearchParams({
        groupID: String(groupID),
        postID: String(postID)
    });

    const response = await fetch(`/api/group/post?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await response.json();

    if (!response.ok) {
        throw new Error(result.message || 'Could not fetch post');
    }

    return result.data;
}

export async function insertPostReaction(postID, value) {

    const response = await fetch(`/api/group/post/reaction`, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
            postID: postID,
            value: value
        })
    });

    const result = await response.json();

    if (!response.ok) {
        throw new Error(result.message || 'Could not update reaction');
    }

    return result;
}