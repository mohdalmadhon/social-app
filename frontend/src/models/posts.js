export class Reaction {
    value;
    userID;
    postID;
    
    constructor(value, postID) {
        this.value = value;
        this.postID = postID;
    }

    getData() {
        return {
            value: this.value,
            userID: this.userID,
            postID: this.postID
        }
    }
}

export class Comment {
    constructor(
        id,
        content,
        postId,
        replyTo,
        votes,
        createdAt,
        user,
        replies = 0,
        imagePath = ''
    ) {
        this.ID = id;
        this.content = content;
        this.postId = postId;
        this.replyTo = replyTo;
        this.votes = votes;
        this.createdAt = createdAt;
        this.user = user;
        this.replies = replies;
        this.imagePath = imagePath;
        this.loadedReplies = null;
        this.showReplies = false;
        this.pending = false;
    }
}

export function createComment(data) {
    return new Comment(
        data.id,
        data.content,
        data.postId,
        data.replyTo ?? null,
        data.votes ?? 0,
        data.createdAt,
        data.user,
        data.replies ?? 0,
        data.imagePath ?? ''
    );
}

export class Post {
    constructor(
        ReactionValue = 0,
        allowComments = false,
        avatarPath = '',
        content = '',
        createdAt = '',
        firstName = '',
        groupId = 0,
        id = 0,
        imagePath = '',
        lastName = '',
        location = '',
        relationship = '',
        taggedPeople = [],
        userId = 0,
        username = '',
        visibility = '',
        visibilityUser = '',
        likeCount = 0,
        disLikeCount = 0
    ) {
        this.ReactionValue = ReactionValue
        this.allowComments = allowComments
        this.avatarPath = avatarPath
        this.content = content
        this.createdAt = createdAt
        this.firstName = firstName
        this.groupId = groupId
        this.id = id
        this.imagePath = imagePath
        this.lastName = lastName
        this.location = location
        this.relationship = relationship
        this.taggedPeople = taggedPeople
        this.userId = userId
        this.username = username
        this.visibility = visibility
        this.visibilityUser = visibilityUser
        this.likeCount = likeCount
        this.disLikeCount = disLikeCount
    }
}

export class Posts {
    constructor(posts = []) {
        this.posts = posts
    }

    addPost(post) {
        this.posts.push(post)
    }

    removePost(postId) {
        this.posts = this.posts.filter(post => post.id !== postId)
    }

    getPost(postId) {
        return this.posts.find(post => post.id === postId)
    }

    clear() {
        this.posts = []
    }
}
