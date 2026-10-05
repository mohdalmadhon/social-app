<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    people: {
        type: Array,
        default: () => []
    }
});

const emit = defineEmits(['close']);

const router = useRouter();

const visiblePeople = computed(() => {
    return props.people.filter(person => {
        return person && (person.id != null || person.ID != null);
    });
});

function getPersonId(person) {
    return person?.id ?? person?.ID ?? null;
}

function getFirstName(person) {
    return person?.firstName ?? person?.FirstName ?? '';
}

function getLastName(person) {
    return person?.lastName ?? person?.LastName ?? '';
}

function getUsername(person) {
    return person?.username ?? person?.UserName ?? '';
}

function getFullName(person) {
    const firstName = getFirstName(person);
    const lastName = getLastName(person);

    const fullName = `${firstName} ${lastName}`.trim();

    if (fullName) {
        return fullName;
    }

    return getUsername(person);
}

function getAvatar(person) {
    return person?.avatarPath ?? person?.Avatar ?? 'avatars/default.png';
}

function getAvatarUrl(person) {
    const avatar = getAvatar(person);

    if (!avatar) {
        return '/uploads/avatars/default.png';
    }

    if (
        avatar.startsWith('http://') ||
        avatar.startsWith('https://') ||
        avatar.startsWith('/')
    ) {
        return avatar;
    }

    return `/uploads/${avatar}`;
}

function openProfile(person) {
    const id = getPersonId(person);

    if (id == null) {
        return;
    }

    emit('close');

    router.push({
        path: '/user',
        query: {
            id: String(id)
        }
    });
}

function closeDialog() {
    emit('close');
}

function handleBackdropClick(event) {
    if (event.target === event.currentTarget) {
        closeDialog();
    }
}
</script>

<template>
    <Teleport to="body">
        <div
            v-if="show"
            class="dialog-overlay"
            @click="handleBackdropClick"
        >
            <div class="tagged-dialog">
                <div class="dialog-header">
                    <h2>Tagged people</h2>

                    <button
                        type="button"
                        class="close-button"
                        @click="closeDialog"
                    >
                        ×
                    </button>
                </div>

                <div v-if="visiblePeople.length" class="people-list">
                    <button
                        v-for="person in visiblePeople"
                        :key="getPersonId(person)"
                        type="button"
                        class="person-row"
                        @click="openProfile(person)"
                    >
                        <img
                            :src="getAvatarUrl(person)"
                            :alt="getFullName(person)"
                            class="person-avatar"
                        />

                        <div class="person-info">
                            <span class="person-name">
                                {{ getFullName(person) }}
                            </span>

                            <span
                                v-if="getUsername(person)"
                                class="person-username"
                            >
                                @{{ getUsername(person) }}
                            </span>
                        </div>
                    </button>
                </div>

                <div v-else class="empty-state">
                    No tagged people.
                </div>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgba(0, 0, 0, 0.65);
}

.tagged-dialog {
    width: 100%;
    max-width: 420px;
    max-height: 80vh;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 8px 8px var(--main-color);
}

.dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 18px;
    border-bottom: 2px solid var(--main-color);
}

.dialog-header h2 {
    margin: 0;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 15px;
    font-weight: 700;
}

.close-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    font-size: 22px;
    line-height: 1;
    cursor: pointer;
}

.close-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.people-list {
    max-height: calc(80vh - 70px);
    overflow-y: auto;
    padding: 8px;
}

.person-row {
    display: flex;
    align-items: center;
    width: 100%;
    gap: 12px;
    padding: 10px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
}

.person-row:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.person-avatar {
    flex-shrink: 0;
    width: 44px;
    height: 44px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    object-fit: cover;
}

.person-info {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 3px;
}

.person-name {
    overflow: hidden;
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.person-username {
    overflow: hidden;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    opacity: 0.7;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.empty-state {
    padding: 30px 20px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    text-align: center;
}

@media (max-width: 500px) {
    .dialog-overlay {
        align-items: flex-end;
        padding: 10px;
    }

    .tagged-dialog {
        max-width: none;
        box-shadow: 4px 4px var(--main-color);
    }
}
</style>