let current = null;

export function claimPlayback(video) {
    if (current && current !== video) {
        current.pause();
    }

    current = video;
}

export function releasePlayback(video) {
    if (current === video) {
        current = null;
    }
}
