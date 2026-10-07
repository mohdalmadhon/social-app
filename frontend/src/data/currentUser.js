import { ref } from 'vue';

export const sessionUserId = ref(null);

export function setSessionUserId(value) {
    const id = Number(value);

    sessionUserId.value = Number.isFinite(id) && id > 0 ? id : null;
}
