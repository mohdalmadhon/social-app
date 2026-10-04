import { buildCommentRequest } from '@/helpers/commentMedia';

export async function getGroupComments(postId, replyTo = null) {
    const params = new URLSearchParams();

    params.append('postId', postId.toString());

    if (replyTo !== null) {
        params.append('replyTo', replyTo.toString());
    }

    const response = await fetch(
        `/api/group/post/comment?${params.toString()}`,
        {
            method: 'GET',
            credentials: 'include'
        }
    );

    const data = await response.json();

    if (!response.ok || !data.status) {
        throw new Error(
            data.message || 'Failed to load comments'
        );
    }

    return {
        comments: data.comments || [],
        userId: data.userId ?? null
    };
}

export async function addGroupComment(
    postId,
    content,
    replyTo = null,
    image = null
) {
    const response = await fetch(
        '/api/group/post/comment',
        buildCommentRequest(postId, content, replyTo, image)
    );

    const data = await response.json();

    if (!response.ok || !data.status) {
        throw new Error(
            data.message || 'Failed to add comment'
        );
    }

    return data.comment;
}

export async function deleteGroupComment(commentId) {
    const response = await fetch(
        `/api/group/post/comment?commentId=${commentId}`,
        {
            method: 'DELETE',
            credentials: 'include'
        }
    );

    const data = await response.json();

    if (!response.ok || !data.status) {
        throw new Error(
            data.message || 'Failed to delete comment'
        );
    }

    return data;
}

export async function voteGroupComment(commentId, vote) {
    const response = await fetch(
        `/api/group/post/comment/vote?commentId=${commentId}&vote=${vote}`,
        {
            method: 'POST',
            credentials: 'include'
        }
    );

    const data = await response.json();

    if (!response.ok || !data.status) {
        throw new Error(
            data.message || 'Failed to vote on comment'
        );
    }

    return data;
}

export async function getGroupPosts(groupID, offset = 0) {
    const params = new URLSearchParams({
        groupID: String(groupID),
        offset: String(offset)
    });

    const response = await fetch(
        `/api/group/posts?${params.toString()}`,
        {
            method: 'GET',
            credentials: 'include'
        }
    );

    const data = await response.json();

    if (!response.ok || !data.status) {
        throw new Error(
            data.message || 'Failed to load group posts'
        );
    }

    return {
        posts: data.data || [],
        userId: data.userId ?? null
    };
}
