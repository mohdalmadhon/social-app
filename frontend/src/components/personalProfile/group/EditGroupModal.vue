<script setup>
import { ref, computed, watch, onMounted } from 'vue';

import { updatePostGroup, getFriends } from '@/api/common/friends.js';
import { avatarUrl, fullName, userKey } from './groupHelpers.js';
import ConfirmModal from './ConfirmModal.vue';

const props = defineProps({
    group: {
        type: Object,
        required: true
    }
});

onMounted(() => {
    console.log(props.group)
})
const emit = defineEmits(['close', 'saved']);

function toMember(user) {
    return {
        ID: user.ID,
        firstName: user.firstName,
        lastName: user.lastName,
        avatar: user.avatar
    };
}

const editName = ref(props.group.Name || '');
const currentMembers = ref((props.group.Users || []).map(toMember));
const saving = ref(false);
const error = ref('');

const searchQuery = ref('');
const searchResults = ref([]);
const searching = ref(false);
let searchDebounce = null;

const showRemoveConfirm = ref(false);
const memberPendingRemoval = ref(null);
const memberPendingRemovalIndex = ref(-1);

async function runSearch(query) {
    searching.value = true;

    try {
        const result = await getFriends(query);
        searchResults.value = Object.entries(result.data || {}).map(
            ([key, value]) => ({
                ...value,
                ID: key
            })
        );
    } catch {
        searchResults.value = [];
    } finally {
        searching.value = false;
    }
}

watch(searchQuery, (value) => {
    clearTimeout(searchDebounce);

    if (!value.trim()) {
        searchResults.value = [];
        return;
    }

    searchDebounce = setTimeout(() => {
        runSearch(value.trim());
    }, 300);
});

const availableToAdd = computed(() => {
    const currentMemberIds = new Set(
        currentMembers.value.map((member) => String(member.ID))
    );

    return searchResults.value.filter(
        (user) => !currentMemberIds.has(String(user.ID))
    );
});

function addMember(user) {
    const alreadyMember = currentMembers.value.some(
        (member) => String(member.ID) === String(user.ID)
    );

    if (alreadyMember) {
        return;
    }

    currentMembers.value.push(toMember(user));
}

function requestRemoveMember(user, index) {
    memberPendingRemoval.value = user;
    memberPendingRemovalIndex.value = index;
    showRemoveConfirm.value = true;
}

function cancelRemoveMember() {
    showRemoveConfirm.value = false;
    memberPendingRemoval.value = null;
    memberPendingRemovalIndex.value = -1;
}

function confirmRemoveMember() {
    if (memberPendingRemovalIndex.value !== -1) {
        currentMembers.value.splice(memberPendingRemovalIndex.value, 1);
    }

    cancelRemoveMember();
}

async function handleSave() {
    error.value = '';

    if (!editName.value.trim()) {
        error.value = 'Group name is required.';
        return;
    }

    if (!currentMembers.value.length) {
        error.value = 'Select at least one member.';
        return;
    }

    saving.value = true;

    try {
        await updatePostGroup({
            groupId: props.group.ID,
            name: editName.value.trim(),
            users: currentMembers.value.map((member) => Number(member.ID))
        });

        emit('saved');
    } catch {
        error.value = 'Failed to save group. Please try again.';
    } finally {
        saving.value = false;
    }
}
</script>

<template>
    <div class="modal-overlay" @click.self="emit('close')">
        <div class="modal">
            <div class="modal-header">
                <h3>Edit Group</h3>

                <button type="button" class="close-button" @click="emit('close')">
                    ×
                </button>
            </div>

            <div class="modal-body">
                <label class="field">
                    <span class="field-label">Group name</span>

                    <input v-model="editName" type="text" placeholder="e.g. goats" class="text-input" />
                </label>

                <div class="field">
                    <span class="field-label">Current members</span>

                    <div v-if="!currentMembers.length" class="state-message">
                        No members yet.
                    </div>

                    <div v-else class="friends-list">
                        <div v-for="(user, index) in currentMembers" :key="userKey(user) || index"
                            class="friend-row friend-row--selected">
                            <img v-if="avatarUrl(user.avatar)" :src="avatarUrl(user.avatar)" :alt="fullName(user)"
                                class="friend-avatar" />

                            <div v-else class="friend-avatar friend-avatar--fallback"></div>

                            <span class="friend-name">
                                {{ fullName(user) }}
                            </span>

                            <button type="button" class="remove-button" @click="requestRemoveMember(user, index)">
                                Remove
                            </button>
                        </div>
                    </div>
                </div>

                <div class="field">
                    <span class="field-label">Add members</span>

                    <input v-model="searchQuery" type="text" placeholder="Search friends by name..."
                        class="text-input" />

                    <div v-if="searching" class="state-message">
                        Searching...
                    </div>

                    <div v-else-if="
                        searchQuery.trim() &&
                        !availableToAdd.length
                    " class="state-message">
                        No matching friends found.
                    </div>

                    <div v-else-if="availableToAdd.length" class="friends-list">
                        <div v-for="user in availableToAdd" :key="userKey(user)" class="friend-row">
                            <img v-if="avatarUrl(user.avatar)" :src="avatarUrl(user.avatar)" :alt="fullName(user)"
                                class="friend-avatar" />

                            <div v-else class="friend-avatar friend-avatar--fallback"></div>

                            <span class="friend-name">
                                {{ fullName(user) }}
                            </span>

                            <button type="button" class="add-button" @click="addMember(user)">
                                Add
                            </button>
                        </div>
                    </div>
                </div>

                <p v-if="error" class="error-message">
                    {{ error }}
                </p>
            </div>

            <div class="modal-footer">
                <button type="button" class="secondary-button" @click="emit('close')" :disabled="saving">
                    Cancel
                </button>

                <button type="button" class="primary-button" @click="handleSave" :disabled="saving">
                    {{ saving ? 'Saving...' : 'Save' }}
                </button>
            </div>
        </div>
    </div>

    <ConfirmModal v-if="showRemoveConfirm" title="Remove Member" confirm-label="Remove"
        @cancel="cancelRemoveMember" @confirm="confirmRemoveMember">
        <p>
            Remove
            <strong>{{ fullName(memberPendingRemoval) }}</strong>
            from this group?
        </p>
    </ConfirmModal>
</template>

<style scoped>
* {
    box-sizing: border-box;
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, .4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal {
    width: 100%;
    max-width: 640px;
    max-height: 90vh;
    max-height: 90dvh;
    display: flex;
    flex-direction: column;
    background: #fff;
    border: 3px solid #1a1a1a;
    border-radius: 12px;
    box-shadow: 6px 6px 0 #1a1a1a;
}

.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 3px solid #1a1a1a;
}

.modal-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 800;
    color: #1a1a1a;
}

.close-button {
    border: none;
    background: none;
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    color: #1a1a1a;
}

.modal-body {
    padding: 24px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.field {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.field-label {
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 800;
    color: #55554f;
    text-transform: uppercase;
}

.text-input {
    padding: 12px 14px;
    border: 2px solid #1a1a1a;
    border-radius: 8px;
    font-size: 15px;
}

.friends-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 320px;
    overflow-y: auto;
    padding: 10px;
    border: 2px solid #1a1a1a;
    border-radius: 8px;
}

.friend-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px;
    border: 2px solid transparent;
    border-radius: 8px;
    font-size: 14px;
}

.friend-row--selected {
    background: #dbeafe;
    border-color: #2563eb;
}

.friend-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 2px solid #1a1a1a;
    object-fit: cover;
    flex-shrink: 0;
}

.friend-avatar--fallback {
    background: #e5e5e5;
}

.friend-name {
    flex: 1;
    color: #1a1a1a;
}

.add-button,
.remove-button {
    padding: 7px 14px;
    border: 2px solid #1a1a1a;
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 800;
    cursor: pointer;
    box-shadow: 2px 2px 0 #1a1a1a;
    transition: transform .1s ease, box-shadow .1s ease;
}

.add-button:hover,
.remove-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0 #1a1a1a;
}

.add-button:active,
.remove-button:active {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0 #1a1a1a;
}

.add-button {
    background: #2563eb;
    color: #fff;
}

.remove-button {
    background: #ef4444;
    color: #fff;
}

.state-message {
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    color: #55554f;
}

.error-message {
    color: #ef4444;
    font-size: 12px;
    font-family: "JetBrains Mono", monospace;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 16px 20px;
    border-top: 3px solid #1a1a1a;
}

.primary-button,
.secondary-button {
    padding: 10px 18px;
    border: 3px solid #1a1a1a;
    border-radius: 8px;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 800;
    cursor: pointer;
    box-shadow: 4px 4px 0 #1a1a1a;
    transition: transform .1s ease, box-shadow .1s ease;
}

.primary-button:hover,
.secondary-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px 0 #1a1a1a;
}

.primary-button:active,
.secondary-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 #1a1a1a;
}

.primary-button {
    background: #2563eb;
    color: #fff;
}

.secondary-button {
    background: #fff;
    color: #1a1a1a;
}

.primary-button:disabled,
.secondary-button:disabled {
    opacity: .6;
    cursor: not-allowed;
}
</style>