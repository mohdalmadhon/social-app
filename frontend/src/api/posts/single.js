export async function getSinglePost(postID) {
    const params = new URLSearchParams({
        postID: String(postID)
    });

    const response = await fetch(`/api/post/single?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
    });

    const result = await response.json();

    if (!response.ok || !result.status) {
        throw new Error(result.message || 'Could not load post');
    }

    return {
        post: result.data,
        userId: result.userId ?? null
    };
}
