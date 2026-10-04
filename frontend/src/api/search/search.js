import { checkSessionResponse } from '@/helpers/auth/auth';
import { router } from '@/router/router';

async function search(type, query, offset) {
    const params = new URLSearchParams({
        search: query,
        offset: String(offset)
    });

    const resp = await fetch(`/api/search/${type}?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    if (!checkSessionResponse(resp)) {
        router.replace('/login');
        return { data: [], hasMore: false };
    }

    let result = null;

    try {
        result = await resp.json();
    } catch {
        result = null;
    }

    if (!resp.ok || !result?.status) {
        throw new Error(result?.message || 'could not search, please try again');
    }

    return {
        data: result.data || [],
        hasMore: Boolean(result.hasMore)
    };
}

export function searchGroups(query, offset = 0) {
    return search('groups', query, offset);
}

export function searchUsers(query, offset = 0) {
    return search('users', query, offset);
}

export function searchPosts(query, offset = 0) {
    return search('posts', query, offset);
}

export async function searchShares(vlaue = '', postID = -1) {
    const resp = await fetch(`/api/user/follow-followers?search=${vlaue}&postID=${postID}`, {
        method: "GET",
        credentials: 'include'
    });
    
    const result = await resp.json();
    if (!resp.ok) {
        throw new Error(result.message || 'could not get users');
    }

    return result;
}

export async function searchShareProfile(search = '') {
    const resp = await fetch(`/api/share/profile?search=${encodeURIComponent(search)}`, {
        method: "GET",
        credentials: 'include'
    });

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not get users');
    }

    return result;
}
