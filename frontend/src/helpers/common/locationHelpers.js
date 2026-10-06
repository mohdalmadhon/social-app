const GOOGLE_MAPS_LINK = (lat, lon) => {
    return `https://www.google.com/maps/search/?api=1&query=${lat},${lon}`;
};

const COUNTRY_TYPES = ['country'];

const STATE_TYPES = [
    'state',
    'region',
    'province',
    'state_district',
    'county',
    'territory'
];

const CITY_TYPES = [
    'city',
    'town',
    'village',
    'municipality',
    'hamlet',
    'city_district',
    'borough',
    'suburb',
    'quarter',
    'neighbourhood'
];

const GENERIC_PLACE_TYPES = [
    'yes',
    'house',
    'building',
    'residential',
    'road',
    'service',
    'unclassified'
];

const MAX_LABEL_LENGTH = 150;

function humanize(value) {
    const text = String(value || '').replace(/_/g, ' ').trim();

    if (!text) {
        return '';
    }

    return text.charAt(0).toUpperCase() + text.slice(1);
}

function getKind(location) {
    const type = location.addresstype || location.type || '';

    if (COUNTRY_TYPES.includes(type)) {
        return 'country';
    }

    if (STATE_TYPES.includes(type)) {
        return 'state';
    }

    if (CITY_TYPES.includes(type)) {
        return 'city';
    }

    return 'place';
}

function getKindLabel(kind, location) {
    if (kind === 'country') {
        return 'Country';
    }

    if (kind === 'state') {
        return 'State';
    }

    if (kind === 'city') {
        return 'City';
    }

    const type = location.type || '';

    if (!type || GENERIC_PLACE_TYPES.includes(type)) {
        return 'Place';
    }

    return humanize(type);
}

function uniqueParts(parts, title) {
    const seen = new Set([String(title || '').toLowerCase()]);
    const result = [];

    parts.forEach(part => {
        const text = String(part || '').trim();
        const key = text.toLowerCase();

        if (!text || seen.has(key)) {
            return;
        }

        seen.add(key);
        result.push(text);
    });

    return result;
}

function cleanLabel(value) {
    return String(value || '')
        .replace(/:/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
        .slice(0, MAX_LABEL_LENGTH)
        .trim();
}

export function getLocationData(locations = []) {
    const result = [];
    const seenLabels = new Set();

    locations.forEach(location => {
        if (!location || location.lat === undefined || location.lon === undefined) {
            return;
        }

        const address = location.address || {};

        const locality =
            address.city ||
            address.town ||
            address.village ||
            address.municipality ||
            address.hamlet ||
            address.city_district ||
            address.county ||
            '';

        const state =
            address.state ||
            address.region ||
            address.province ||
            address.state_district ||
            '';

        const country = address.country || '';

        const area =
            address.suburb ||
            address.neighbourhood ||
            address.quarter ||
            address.borough ||
            '';

        const kind = getKind(location);

        const title =
            location.name ||
            String(location.display_name || '').split(',')[0].trim() ||
            locality ||
            state ||
            country;

        const context = kind === 'place'
            ? [area, locality, state, country]
            : [locality, state, country];

        const subtitleParts = uniqueParts(context, title);

        const label = cleanLabel([title, ...subtitleParts].join(', '));

        if (!label) {
            return;
        }

        const labelKey = `${kind}:${label.toLowerCase()}`;

        if (seenLabels.has(labelKey)) {
            return;
        }

        seenLabels.add(labelKey);

        result.push({
            key: location.place_id ?? `${location.lat},${location.lon}`,
            lat: location.lat,
            lon: location.lon,
            link: GOOGLE_MAPS_LINK(location.lat, location.lon),
            kind,
            kindLabel: getKindLabel(kind, location),
            title,
            subtitle: subtitleParts.join(', '),
            label,
            city: locality || state,
            state,
            country
        });
    });

    return result;
}

export function normalizeProfilePost(raw) {
    return {
        id: raw.id,
        userId: raw.userId,
        firstName: raw.firstName,
        lastName: raw.lastName,
        username: raw.username,
        avatarPath: raw.avatarPath,
        content: raw.content,
        imagePath: raw.imagePath,
        allowComments: raw.allowComments,
        location: raw.location,
        groupId: raw.groupId,
        groupName: raw.GroupName ?? raw.groupName ?? '',
        createdAt: raw.createdAt,
        relationship: raw.relationship,
        visibility: raw.visibility,
        visibilityUser: raw.visibilityUser,
        taggedPeople: raw.taggedPeople ?? [],
        likeCount: raw.likeCount ?? 0,
        dislikeCount: raw.disLikeCount ?? 0,
        commentCount: raw.commentCount ?? 0,
        reactionValue: raw.ReactionValue ?? raw.reactionValue ?? 0
    };
}