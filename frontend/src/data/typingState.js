import { reactive } from 'vue';

export const typingUsers = reactive({});
export const groupTyping = reactive({});

const timers = new Map();
const groupTimers = new Map();
const TYPING_TIMEOUT = 5000;

export function setTyping(userID, typing) {
    const key = String(userID);

    clearTimeout(timers.get(key));
    timers.delete(key);

    if (!typing) {
        delete typingUsers[key];
        return;
    }

    typingUsers[key] = true;

    timers.set(
        key,
        setTimeout(() => {
            delete typingUsers[key];
            timers.delete(key);
        }, TYPING_TIMEOUT)
    );
}

export function isUserTyping(userID) {
    return Boolean(typingUsers[String(userID)]);
}

function clearGroupTyper(groupKey, userKey) {
    const typers = groupTyping[groupKey];

    if (!typers) {
        return;
    }

    delete typers[userKey];

    if (!Object.keys(typers).length) {
        delete groupTyping[groupKey];
    }
}

export function setGroupTyping(groupID, userID, typing) {
    const groupKey = String(groupID);
    const userKey = String(userID);
    const timerKey = `${groupKey}:${userKey}`;

    clearTimeout(groupTimers.get(timerKey));
    groupTimers.delete(timerKey);

    if (!typing) {
        clearGroupTyper(groupKey, userKey);
        return;
    }

    if (!groupTyping[groupKey]) {
        groupTyping[groupKey] = {};
    }

    groupTyping[groupKey][userKey] = true;

    groupTimers.set(
        timerKey,
        setTimeout(() => {
            clearGroupTyper(groupKey, userKey);
            groupTimers.delete(timerKey);
        }, TYPING_TIMEOUT)
    );
}

export function groupTypingCount(groupID) {
    return Object.keys(groupTyping[String(groupID)] || {}).length;
}
