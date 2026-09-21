export async function createGroup(data) {
    const formData = new FormData();

    formData.append('title', data.title);
    formData.append('description', data.description || '');
    formData.append('userIDs', JSON.stringify(data.userIDs || []));

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
