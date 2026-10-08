import { safeSearch } from '@/helpers/limits';
export async function searchGroupMentions(groupID, search = '') {
    const params = new URLSearchParams({
        groupID: String(groupID),
        search: safeSearch(search)
    });

    const response = await fetch(`/api/group/mentions?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'Could not load members');
    }

    return result.data || [];
}
