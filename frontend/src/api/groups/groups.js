export async function createGroup(data) {
    const formData = new FormData();

    formData.append('title', data.title);
    formData.append('description', data.description || '');
    formData.append('users', JSON.stringify(data.users || []));

    if (data.avatar) {
        formData.append('avatar', data.avatar);
    }

    const resp = await fetch('/api/groups', {
        method: 'POST',
        credentials: 'include',
        body: formData
    });

    const result = await resp.json();

    if (!resp.ok || !result.status) {
        throw new Error(result.message || 'Could not create group');
    }

    return result;
}

export async function kickMember(groupID, userID) {
    const resp = await fetch('/api/group/kick', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            groupID: Number(groupID),
            userID: Number(userID)
        })
    });

    const result = await resp.json();

    if (!resp.ok || !result.status) {
        throw new Error(result.message || 'Could not remove member');
    }

    return result;
}

export async function leaveGroup(groupID) {
    const resp = await fetch('/api/group/leave', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            groupID: Number(groupID)
        })
    });

    const result = await resp.json();

    if (!resp.ok || !result.status) {
        throw new Error(result.message || 'Could not leave group');
    }

    return result;
}
