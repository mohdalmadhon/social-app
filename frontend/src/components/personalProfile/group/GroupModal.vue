<script setup>
import { LIMITS, validateGroupTitle } from '@/helpers/limits';
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { addPostGroup, getFriends } from '@/api/common/friends.js';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    group: {
        type: Object,
        default: null
    }
});

const emit = defineEmits(['close', 'save']);

const name = ref(props.group?.name || '');
const search = ref('');
const friends = ref([]);
const selected = ref(props.group?.members ? [...props.group.members] : []);
const loading = ref(false);
const nameTouched = ref(false);

let searchTimeout;

async function getSearchedFriends() {
    loading.value = true;

    try {
        const result = await getFriends(search.value.trim());

        const data = result?.data;

        if (data && typeof data === 'object' && !Array.isArray(data)) {
            friends.value = Object.entries(data).map(([id, friend]) => ({
                ...friend,
                id: Number(id)
            }));
        } else if (Array.isArray(data)) {
            friends.value = data.map(friend => ({
                ...friend,
                id: Number(friend.ID ?? friend.id)
            }));
        } else {
            friends.value = [];
        }

        console.log('friends:', friends.value);
    } catch (error) {
        console.error(error);
        friends.value = [];
    } finally {
        loading.value = false;
    }
}

function handleSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(getSearchedFriends, 300);
}

const filteredFriends = computed(() => friends.value);

const nameError = computed(() => {
    if (!nameTouched.value) return null;
    return validateGroupTitle(name.value) || null;
});

const canSubmit = computed(() => !validateGroupTitle(name.value));

function toggleFriend(friend) {
    const index = selected.value.findIndex(item => Number(item.id) === Number(friend.id));

    if (index !== -1) {
        selected.value.splice(index, 1);
        return;
    }

    selected.value.push(friend);
}

function removeSelected(friend) {
    selected.value = selected.value.filter(
        item => Number(item.id) !== Number(friend.id)
    );
}

function isSelected(friend) {
    return selected.value.some(
        item => Number(item.id) === Number(friend.id)
    );
}

function initials(friend) {
    const first = friend.firstName?.charAt(0) || '';
    const last = friend.lastName?.charAt(0) || '';

    return (first + last).toUpperCase();
}

async function submit() {
    nameTouched.value = true;

    if (!canSubmit.value) {
        return;
    }

    const data = {
        name: name.value.trim(),
        users: selected.value.map(friend => Number(friend.id))
    };

    try {
        const result = await addPostGroup(data);

        if (!result.status) {
            addNotification(result.message, 'error');
            return;
        }

        emit('save', {
            name: name.value.trim(),
            members: [...selected.value]
        });

        addNotification('Group created!', 'success');
    } catch (err) {
        console.error(err);
        addNotification(err?.message || 'Could not create group', 'error');
    }
}

function closeModal() {
    emit('close');
}

function handleKeydown(e) {
    if (e.key === 'Escape') {
        closeModal();
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown);
    getSearchedFriends();
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
    clearTimeout(searchTimeout);
});
</script>

<template>
    <div class="overlay" @click.self="closeModal">
        <div class="modal">
            <header>
                <div>
                    <span class="eyebrow">
                        {{ group ? 'UPDATE COLLECTION' : 'NEW COLLECTION' }}
                    </span>

                    <h2>
                        {{ group ? 'Edit Group' : 'Create Group' }}
                    </h2>

                    <p>
                        {{ group
                            ? 'Update your group and manage its members.'
                            : 'Create a group and organize your friends.' }}
                    </p>
                </div>

                <button type="button" class="close" @click="closeModal">
                    ×
                </button>
            </header>

            <div class="form-group">
                <label>
                    Group name <span class="required">*</span>
                </label>

                <input v-model="name" :maxlength="LIMITS.groupTitle" placeholder="Close Friends" class="input" :class="{ invalid: nameError }"
                    @blur="nameTouched = true" />

                <p v-if="nameError" class="field-error">
                    {{ nameError }}
                </p>
            </div>

            <div class="form-group">
                <label>Add members</label>

                <div class="search-wrapper">
                    <span class="search-icon">⌕</span>

                    <input v-model="search" placeholder="Search your friends..." class="input" @input="handleSearch" />
                </div>
            </div>

            <div class="selected-chips" v-if="selected.length">
                <span v-for="friend in selected" :key="friend.id" class="chip">
                    <span class="chip-avatar">
                        <img v-if="friend.avatar" :src="`/uploads/${friend.avatar}`" :alt="friend.firstName">
                        <span v-else>{{ initials(friend) }}</span>
                    </span>

                    {{ friend.firstName }}

                    <button type="button" class="chip-remove" @click="removeSelected(friend)">
                        ×
                    </button>
                </span>
            </div>

            <div class="friends">
                <div v-if="loading" class="no-friends">
                    Searching friends...
                </div>

                <template v-else>
                    <button v-for="friend in filteredFriends" :key="friend.id" type="button" class="friend"
                        :class="{ selected: isSelected(friend) }" @click="toggleFriend(friend)">
                        <div class="avatar">
                            <img v-if="friend.avatar" :src="`/uploads/${friend.avatar}`"
                                :alt="`${friend.firstName} ${friend.lastName}`">

                            <span v-else>
                                {{ initials(friend) }}
                            </span>
                        </div>

                        <div class="friend-info">
                            <strong>
                                {{ friend.firstName }} {{ friend.lastName }}
                            </strong>
                        </div>

                        <div class="check" :class="{ active: isSelected(friend) }">
                            {{ isSelected(friend) ? '✓' : '' }}
                        </div>
                    </button>

                    <div v-if="!filteredFriends.length" class="no-friends">
                        No friends found.
                    </div>
                </template>
            </div>

            <div class="selected">
                <span class="selected-count">
                    {{ selected.length }}
                </span>

                <span>
                    {{ selected.length === 1 ? 'friend' : 'friends' }}
                    selected
                </span>
            </div>

            <footer>
                <button type="button" class="cancel" @click="closeModal">
                    Cancel
                </button>

                <button type="button" class="submit" @click="submit">
                    {{ group ? 'Save Changes' : 'Create Group' }}
                </button>
            </footer>
        </div>
    </div>
</template>

<style scoped>
* {
    box-sizing: border-box;
}

.overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: grid;
    place-items: center;
    padding: 20px;
    background: rgba(15, 23, 42, .55);
}

.modal {
    width: min(520px, 100%);
    max-height: 90vh;
    max-height: 90dvh;
    overflow-y: auto;
    background: #ffffff;
    border: 3px solid #1a1a1a;
    border-radius: 16px;
    padding: 26px;
    box-shadow: 8px 8px 0 #1a1a1a;
}

header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 15px;
    margin-bottom: 24px;
    padding-bottom: 18px;
    border-bottom: 3px solid #1a1a1a;
}

.eyebrow {
    display: block;
    margin-bottom: 6px;
    color: #2563eb;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 1.5px;
}

header h2 {
    margin: 0 0 6px;
    color: #1a1a1a;
    font-size: 24px;
    font-weight: 800;
}

header p {
    max-width: 360px;
    margin: 0;
    color: #55554f;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    line-height: 1.6;
}

.close {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    flex-shrink: 0;
    border: 3px solid #1a1a1a;
    border-radius: 8px;
    background: #ffffff;
    color: #1a1a1a;
    font-size: 20px;
    line-height: 1;
    box-shadow: 3px 3px 0 #1a1a1a;
    transition: transform .1s ease, box-shadow .1s ease;
    cursor: pointer;
}

.close:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px 0 #1a1a1a;
}

.close:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px 0 #1a1a1a;
}

.form-group {
    margin-bottom: 18px;
}

label {
    display: block;
    margin-bottom: 7px;
    color: #1a1a1a;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: .3px;
}

.required {
    color: #d93025;
}

.input {
    width: 100%;
    padding: 11px 12px;
    border: 3px solid #1a1a1a;
    border-radius: 8px;
    outline: none;
    background: #ffffff;
    color: #1a1a1a;
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    transition: box-shadow .15s ease;
}

.input::placeholder {
    color: #a3a39a;
}

.input:focus {
    box-shadow: 3px 3px 0 #2563eb;
}

.input.invalid {
    border-color: #d93025;
}

.input.invalid:focus {
    box-shadow: 3px 3px 0 #d93025;
}

.field-error {
    margin: 6px 0 0;
    color: #d93025;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
}

.search-wrapper {
    position: relative;
}

.search-wrapper .input {
    padding-left: 34px;
}

.search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 18px;
    opacity: .7;
    pointer-events: none;
}

.selected-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;
}

.chip {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 4px 6px 4px 4px;
    border: 3px solid #1a1a1a;
    border-radius: 30px;
    background: #2563eb;
    color: #ffffff;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 700;
    box-shadow: 3px 3px 0 #1a1a1a;
}

.chip-avatar {
    width: 22px;
    height: 22px;
    display: grid;
    place-items: center;
    overflow: hidden;
    flex-shrink: 0;
    border: 2px solid #1a1a1a;
    border-radius: 50%;
    background: #ffffff;
    color: #2563eb;
    font-size: 9px;
    font-weight: 800;
}

.chip-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.chip-remove {
    width: 17px;
    height: 17px;
    display: grid;
    place-items: center;
    border: 0;
    border-radius: 50%;
    background: rgba(255, 255, 255, .3);
    color: #ffffff;
    font-size: 11px;
    line-height: 1;
    cursor: pointer;
}

.chip-remove:hover {
    background: rgba(255, 255, 255, .5);
}

.friends {
    max-height: 245px;
    overflow-y: auto;
    border: 3px solid #1a1a1a;
    border-radius: 10px;
}

.friend {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 11px;
    padding: 10px 12px;
    border: 0;
    border-bottom: 3px solid #1a1a1a;
    background: #ffffff;
    text-align: left;
    transition: background .1s ease;
    cursor: pointer;
}

.friend:last-child {
    border-bottom: 0;
}

.friend:hover {
    background: #f4f4f0;
}

.friend.selected {
    background: #2563eb;
    color: #ffffff;
}

.avatar {
    width: 38px;
    height: 38px;
    display: grid;
    place-items: center;
    overflow: hidden;
    flex-shrink: 0;
    border: 3px solid #1a1a1a;
    border-radius: 50%;
    background: #ffffff;
    color: #1a1a1a;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 800;
}

.friend.selected .avatar {
    background: #ffffff;
    color: #2563eb;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.friend-info {
    flex: 1;
    min-width: 0;
}

.friend-info strong {
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 700;
}

.check {
    width: 22px;
    height: 22px;
    display: grid;
    place-items: center;
    flex-shrink: 0;
    border: 3px solid #1a1a1a;
    border-radius: 50%;
    font-size: 11px;
    font-weight: 800;
}

.friend.selected .check {
    border-color: #ffffff;
}

.check.active {
    background: #1a1a1a;
    color: #ffffff;
}

.friend.selected .check.active {
    background: #ffffff;
    color: #2563eb;
}

.no-friends {
    padding: 28px 15px;
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    color: #55554f;
}

.selected {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 14px 0 18px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    color: #55554f;
}

.selected-count {
    min-width: 24px;
    height: 24px;
    display: grid;
    place-items: center;
    border: 3px solid #1a1a1a;
    border-radius: 6px;
    font-weight: 800;
    color: #1a1a1a;
}

footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding-top: 18px;
    border-top: 3px solid #1a1a1a;
}

.cancel,
.submit {
    padding: 11px 20px;
    border: 3px solid #1a1a1a;
    border-radius: 8px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 800;
    box-shadow: 4px 4px 0 #1a1a1a;
    transition: transform .1s ease, box-shadow .1s ease;
    cursor: pointer;
}

.cancel {
    background: #ffffff;
    color: #1a1a1a;
}

.submit {
    background: #2563eb;
    color: #ffffff;
}

.cancel:hover,
.submit:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px 0 #1a1a1a;
}

.cancel:active,
.submit:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 #1a1a1a;
}

@media (max-width: 500px) {
    .overlay {
        padding: 10px;
    }

    .modal {
        padding: 20px;
    }

    footer {
        flex-direction: column-reverse;
    }

    .cancel,
    .submit {
        width: 100%;
        text-align: center;
    }
}
</style>