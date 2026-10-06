import { reactive } from 'vue';

export const typingUsers = reactive({});

const timers = new Map();
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
