export class GroupComment {
    constructor(
        id,
        content,
        postId,
        replyTo,
        votes,
        createdAt,
        user,
        replies = 0
    ) {
        this.ID = id;
        this.content = content;
        this.postId = postId;
        this.replyTo = replyTo;
        this.votes = votes;
        this.createdAt = createdAt;
        this.user = user;
        this.replies = replies;
        this.loadedReplies = null;
        this.showReplies = false;
        this.pending = false;
    }
}

export function createGroupComment(data) {
    return new GroupComment(
        data.id,
        data.content,
        data.postId,
        data.replyTo ?? null,
        data.votes ?? 0,
        data.createdAt,
        data.user,
        data.replies ?? 0
    );
}
