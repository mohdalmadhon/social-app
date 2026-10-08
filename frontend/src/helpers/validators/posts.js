import { LIMITS, charCount } from '@/helpers/limits';

export function validatePost(data) {
    if (typeof data.content !== 'string') {
        return {
            field: 'content',
            message: 'invalid content'
        };
    }

    if (charCount(data.content) > LIMITS.postContent) {
        return {
            field: 'content',
            message: `content cannot be more than ${LIMITS.postContent} characters`
        };
    }

    if (data.content.trim().length === 0) {
        return {
            field: 'content',
            message: 'content cannot be empty'
        };
    }

    if (data.allowComments !== 0 && data.allowComments !== 1) {
        return {
            field: 'allowComments',
            message: 'invalid allow comments code'
        };
    }

    if (typeof data.groupID !== 'number' || data.groupID < -1) {
        return {
            field: 'groupID',
            message: 'invalid group code'
        };
    }

    if (data.location && charCount(data.location) > LIMITS.postLocation) {
        return {
            field: 'location',
            message: 'location is too long'
        };
    }

    if (data.location) {
        const locationParts = data.location.split(':');

        if (
            locationParts.length !== 3 ||
            !locationParts[0].trim() ||
            !locationParts[1].trim() ||
            !locationParts[2].trim() ||
            Number.isNaN(Number(locationParts[1])) ||
            Number.isNaN(Number(locationParts[2]))
        ) {
            return {
                field: 'location',
                message: 'invalid location'
            };
        }
    }

    if (
        !Array.isArray(data.taggedPeople) ||
        !data.taggedPeople.every(
            id => typeof id === 'number' && Number.isFinite(id)
        )
    ) {
        return {
            field: 'tags',
            message: 'invalid tagged people'
        };
    }

    if (data.taggedPeople.length > LIMITS.postTags) {
        return {
            field: 'tags',
            message: `you can tag at most ${LIMITS.postTags} people`
        };
    }

    if (data.image && data.image.size > LIMITS.postMediaSize) {
        return {
            field: 'image',
            message: 'file must be smaller than 50MB'
        };
    }

    if (data.image) {
        const allowedTypes = [
            'image/png',
            'image/jpeg',
            'image/gif',
            'video/mp4'
        ];

        if (!allowedTypes.includes(data.image.type)) {
            return {
                field: 'image',
                message: 'image must be png, jpg or gif\n or video of type mp4'
            };
        }

        const extension = data.image.name
            .split('.')
            .pop()
            .toLowerCase();

        if (!['png', 'jpg', 'jpeg', 'gif', 'mp4'].includes(extension)) {
            return {
                field: 'image',
                message: 'image must be png, jpg or gif'
            };
        }
    }

    return {
        field: null,
        message: null
    };
}