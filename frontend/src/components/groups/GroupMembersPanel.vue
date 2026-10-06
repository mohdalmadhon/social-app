<script setup>
import { ref, watch, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import Groupssearch from './Groupssearch.vue';
import GroupRequestsDialog from './GroupRequestsDialog.vue';
import { searchInvites } from '@/api/chats/search';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    groupID: {
        type: [Number, String],
        required: true
    },
    members: {
        type: Array,
        default: () => []
    },
    isOwner: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['close']);

const router = useRouter();

const search = ref('');
const memberList = ref([...props.members]);
const offset = ref(props.members.length);
const limit = 20;
const loading = ref(false);
const hasMore = ref(true);

const showInviteDialog = ref(false);
const showConfirmDialog = ref(false);
const showRequestsDialog = ref(false);

const inviteSearch = ref('');
const inviteUsers = ref([]);
const selectedUsers = ref([]);
const inviteLoading = ref(false);
const inviteSending = ref(false);
const inviteError = ref('');
const inviteSuccess = ref('');

let debounceTimer = null;
let scrollTimer = null;
let inviteSearchTimer = null;
let requestNumber = 0;
let inviteRequestNumber = 0;

function getMemberId(member) {
    return member.ID ?? member.id;
}

function getMemberName(member) {
    return `${member.firstName || member.FirstName || ''} ${member.lastName || member.LastName || ''}`.trim();
}

function getMemberAvatar(member) {
    const avatar = member.avatar ?? member.Avatar;

    if (!avatar) {
        return '';
    }

    if (avatar.startsWith('/')) {
        return avatar;
    }

    return `/uploads/${avatar}`;
}

function initials(name) {
    if (!name) {
        return '?';
    }

    return name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map(word => word[0]?.toUpperCase())
        .join('');
}

async function getMembers(reset = false) {
    if (loading.value) {
        return;
    }

    if (!reset && !hasMore.value) {
        return;
    }

    loading.value = true;

    const currentRequest = ++requestNumber;

    try {
        const currentOffset = reset ? 0 : offset.value;

        const params = new URLSearchParams({
            groupID: String(props.groupID),
            offset: String(currentOffset),
            search: search.value.trim()
        });

        const response = await fetch(
            `/api/group/search?${params.toString()}`
        );

        if (!response.ok) {
            throw new Error('Failed to get members');
        }

        const result = await response.json();

        if (currentRequest !== requestNumber) {
            return;
        }

        if (!result.status) {
            throw new Error(
                result.message || 'Failed to get members'
            );
        }

        const newMembers = result.data || [];

        if (reset) {
            memberList.value = newMembers;
            offset.value = newMembers.length;
        } else {
            memberList.value.push(...newMembers);
            offset.value += newMembers.length;
        }

        hasMore.value = newMembers.length === limit;
    } catch (error) {
        console.error(error);
    } finally {
        if (currentRequest === requestNumber) {
            loading.value = false;
        }
    }
}

function handleSearch(value) {
    search.value = value;
}

function handleScroll(event) {
    if (scrollTimer) {
        return;
    }

    scrollTimer = setTimeout(() => {
        scrollTimer = null;

        const element = event.currentTarget;

        const distanceFromBottom =
            element.scrollHeight -
            element.scrollTop -
            element.clientHeight;

        if (distanceFromBottom <= 80) {
            getMembers();
        }
    }, 200);
}

function openProfile(member) {
    const id = getMemberId(member);

    if (!id) {
        return;
    }

    emit('close');

    router.push(`/user?id=${id}`);
}

function openRequests() {
    showRequestsDialog.value = true;
}

function closeRequests() {
    showRequestsDialog.value = false;
}

function openInvite() {
    showInviteDialog.value = true;
    showConfirmDialog.value = false;
    inviteSearch.value = '';
    inviteUsers.value = [];
    selectedUsers.value = [];
    inviteError.value = '';
    inviteSuccess.value = '';

    searchInviteUsers();
}

function closeInvite() {
    if (inviteSending.value) {
        return;
    }

    showInviteDialog.value = false;
    showConfirmDialog.value = false;
    inviteSearch.value = '';
    inviteUsers.value = [];
    selectedUsers.value = [];
    inviteError.value = '';
    inviteSuccess.value = '';
}

function getInviteUserId(user) {
    return user.ID ?? user.id;
}

function getInviteUserName(user) {
    return `${user.firstName || user.FirstName || ''} ${user.lastName || user.LastName || ''}`.trim();
}

function getInviteUserAvatar(user) {
    const avatar = user.avatar ?? user.Avatar;

    if (!avatar) {
        return '';
    }

    if (avatar.startsWith('/')) {
        return avatar;
    }

    return `/uploads/${avatar}`;
}

function isSelected(user) {
    const id = getInviteUserId(user);

    return selectedUsers.value.some(
        selected => getInviteUserId(selected) === id
    );
}

function toggleUser(user) {
    const id = getInviteUserId(user);

    if (!id) {
        return;
    }

    const index = selectedUsers.value.findIndex(
        selected => getInviteUserId(selected) === id
    );

    if (index === -1) {
        selectedUsers.value.push(user);
    } else {
        selectedUsers.value.splice(index, 1);
    }
}

async function searchInviteUsers() {
    const currentRequest = ++inviteRequestNumber;

    inviteLoading.value = true;
    inviteError.value = '';

    try {
        const result = await searchInvites(
            inviteSearch.value.trim(),
            props.groupID
        );

        if (currentRequest !== inviteRequestNumber) {
            return;
        }

        inviteUsers.value = result.data || [];
    } catch (error) {
        if (currentRequest !== inviteRequestNumber) {
            return;
        }

        console.error(error);
        inviteUsers.value = [];
        inviteError.value =
            error.message || 'Could not search users';
    } finally {
        if (currentRequest === inviteRequestNumber) {
            inviteLoading.value = false;
        }
    }
}

function handleInviteSearch(value) {
    inviteSearch.value = value;

    clearTimeout(inviteSearchTimer);

    inviteSearchTimer = setTimeout(() => {
        searchInviteUsers();
    }, 300);
}

function openConfirmDialog() {
    if (!selectedUsers.value.length) {
        return;
    }

    inviteError.value = '';
    showConfirmDialog.value = true;
}

function closeConfirmDialog() {
    if (inviteSending.value) {
        return;
    }

    showConfirmDialog.value = false;
}

async function confirmInvite() {
    if (!selectedUsers.value.length || inviteSending.value) {
        return;
    }

    inviteSending.value = true;
    inviteError.value = '';

    const userIDs = selectedUsers.value
        .map(user => Number(getInviteUserId(user)))
        .filter(
            id => Number.isInteger(id) && id > 0
        );

    if (!userIDs.length) {
        inviteError.value = 'No valid users selected';
        inviteSending.value = false;
        return;
    }

    try {
        const response = await fetch(
            '/api/group/invite',
            {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    groupID: Number(props.groupID),
                    userIDs
                })
            }
        );

        const result = await response.json();

        if (!response.ok || !result.status) {
            throw new Error(
                result.message || 'Could not send invitation'
            );
        }

        showConfirmDialog.value = false;
        const parts = [];

        if (result.requested) {
            parts.push(`${result.requested} invited`);
        }

        inviteSuccess.value =
            parts.join(', ') || result.message || 'Invite sent';

        setTimeout(() => {
            showInviteDialog.value = false;
            inviteSuccess.value = '';
            selectedUsers.value = [];
            inviteUsers.value = [];
            inviteSearch.value = '';
        }, 800);
    } catch (error) {
        console.error(error);

        inviteError.value =
            error.message || 'Could not send invitation';
    } finally {
        inviteSending.value = false;
    }
}

watch(search, () => {
    clearTimeout(debounceTimer);

    debounceTimer = setTimeout(() => {
        offset.value = 0;
        hasMore.value = true;
        getMembers(true);
    }, 300);
});

watch(
    () => props.members,
    value => {
        if (!search.value.trim()) {
            memberList.value = [...value];
            offset.value = value.length;
            hasMore.value = value.length >= limit;
        }
    }
);

watch(
    () => props.show,
    visible => {
        if (visible && memberList.value.length === 0) {
            getMembers(true);
        }
    }
);

onBeforeUnmount(() => {
    clearTimeout(debounceTimer);
    clearTimeout(scrollTimer);
    clearTimeout(inviteSearchTimer);
    inviteRequestNumber++;
});
</script>

<template>
    <div
        v-if="show"
        class="members-panel"
    >
        <div class="members-panel-header">
            <h3 class="members-title">
                {{ memberList.length }}
                {{ memberList.length === 1 ? 'Member' : 'Members' }}
            </h3>

            <div class="header-actions">
                <button v-if="isOwner"
                    class="requests-button"
                    type="button"
                    @click="openRequests"
                >
                    Requests
                </button>

                <button
                    class="invite-button"
                    type="button"
                    @click="openInvite"
                >
                    + Invite
                </button>

                <button
                    class="close-button"
                    type="button"
                    @click="$emit('close')"
                >
                    ×
                </button>
            </div>
        </div>

        <Groupssearch
            :model-value="search"
            placeholder="Search members..."
            @update:model-value="handleSearch"
        />

        <div
            class="members-list"
            @scroll="handleScroll"
        >
            <div
                v-for="member in memberList"
                :key="getMemberId(member)"
                class="member-row"
                @click="openProfile(member)"
            >
                <div class="member-avatar">
                    <img
                        v-if="getMemberAvatar(member)"
                        :src="getMemberAvatar(member)"
                        :alt="getMemberName(member)"
                        class="member-avatar-img"
                    >

                    <span
                        v-else
                        class="member-avatar-fallback"
                    >
                        {{ initials(getMemberName(member)) }}
                    </span>
                </div>

                <div class="member-info">
                    <span class="member-name">
                        {{ getMemberName(member) }}
                    </span>
                </div>
            </div>

            <div
                v-if="loading"
                class="loading-members"
            >
                Loading...
            </div>

            <p
                v-if="!loading && !memberList.length"
                class="no-members"
            >
                No members found.
            </p>
        </div>

        <div
            v-if="showInviteDialog"
            class="dialog-overlay"
            @click.self="closeInvite"
        >
            <div class="invite-dialog">
                <div class="dialog-header">
                    <h3>Invite Members</h3>

                    <button
                        type="button"
                        class="dialog-close"
                        @click="closeInvite"
                    >
                        ×
                    </button>
                </div>

                <Groupssearch
                    :model-value="inviteSearch"
                    placeholder="Search users..."
                    @update:model-value="handleInviteSearch"
                />

                <div class="selected-count">
                    {{ selectedUsers.length }}
                    {{ selectedUsers.length === 1 ? 'user' : 'users' }}
                    selected
                </div>

                <div class="invite-users-list">
                    <button
                        v-for="user in inviteUsers"
                        :key="getInviteUserId(user)"
                        type="button"
                        class="invite-user"
                        :class="{ selected: isSelected(user) }"
                        @click="toggleUser(user)"
                    >
                        <div class="invite-user-avatar">
                            <img
                                v-if="getInviteUserAvatar(user)"
                                :src="getInviteUserAvatar(user)"
                                :alt="getInviteUserName(user)"
                            >

                            <span v-else>
                                {{ initials(getInviteUserName(user)) }}
                            </span>
                        </div>

                        <div class="invite-user-info">
                            <span class="invite-user-name">
                                {{ getInviteUserName(user) }}
                            </span>
                        </div>

                        <div
                            class="select-check"
                            :class="{ checked: isSelected(user) }"
                        >
                            {{ isSelected(user) ? '✓' : '' }}
                        </div>
                    </button>

                    <div
                        v-if="inviteLoading"
                        class="dialog-loading"
                    >
                        Searching...
                    </div>

                    <div
                        v-if="!inviteLoading && inviteError"
                        class="dialog-error"
                    >
                        {{ inviteError }}
                    </div>

                    <div
                        v-if="
                            !inviteLoading &&
                            !inviteError &&
                            !inviteUsers.length
                        "
                        class="dialog-empty"
                    >
                        No one available to invite. Only friends, followers and people you follow whose group invite settings allow it are listed.
                    </div>
                </div>

                <div
                    v-if="inviteSuccess"
                    class="dialog-success"
                >
                    {{ inviteSuccess }}
                </div>

                <div class="dialog-footer">
                    <button
                        type="button"
                        class="cancel-button"
                        @click="closeInvite"
                    >
                        Cancel
                    </button>

                    <button
                        type="button"
                        class="confirm-invite-button"
                        :disabled="
                            !selectedUsers.length ||
                            inviteSending ||
                            !!inviteSuccess
                        "
                        @click="openConfirmDialog"
                    >
                        Invite

                        <span v-if="selectedUsers.length">
                            ({{ selectedUsers.length }})
                        </span>
                    </button>
                </div>
            </div>
        </div>

        <div
            v-if="showConfirmDialog"
            class="dialog-overlay confirm-overlay"
            @click.self="closeConfirmDialog"
        >
            <div class="confirm-dialog">
                <div class="confirm-icon">
                    ?
                </div>

                <h3>Confirm Invitation</h3>

                <p>
                    Are you sure you want to invite
                    <strong>{{ selectedUsers.length }}</strong>
                    {{ selectedUsers.length === 1 ? 'user' : 'users' }}
                    to this group?
                </p>

                <div class="confirm-users">
                    <span
                        v-for="user in selectedUsers"
                        :key="getInviteUserId(user)"
                        class="confirm-user"
                    >
                        {{ getInviteUserName(user) }}
                    </span>
                </div>

                <div
                    v-if="inviteError"
                    class="dialog-error"
                >
                    {{ inviteError }}
                </div>

                <div class="confirm-actions">
                    <button
                        type="button"
                        class="cancel-button"
                        :disabled="inviteSending"
                        @click="closeConfirmDialog"
                    >
                        Cancel
                    </button>

                    <button
                        type="button"
                        class="confirm-invite-button"
                        :disabled="inviteSending"
                        @click="confirmInvite"
                    >
                        {{ inviteSending ? 'Sending...' : 'Confirm Invite' }}
                    </button>
                </div>
            </div>
        </div>

        <GroupRequestsDialog
            :show="showRequestsDialog"
            :group-id="groupID"
            @close="closeRequests"
        />
    </div>
</template>

<style scoped>
.members-panel {
    position: absolute;
    z-index: 20;
    top: calc(100% + 10px);
    right: 24px;
    width: 320px;
    max-width: calc(100vw - 48px);
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.members-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
}

.members-title {
    margin: 0;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 16px;
    font-weight: 700;
}

.header-actions {
    display: flex;
    align-items: center;
    gap: 7px;
}

.requests-button,
.invite-button {
    min-height: 28px;
    padding: 5px 9px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
}

.requests-button {
    background: var(--bg-color);
    color: var(--main-color);
}

.requests-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.invite-button {
    background: var(--main-color);
    color: var(--bg-color);
}

.invite-button:hover {
    background: var(--bg-color);
    color: var(--main-color);
}

.close-button {
    width: 26px;
    height: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 14px;
    line-height: 1;
    cursor: pointer;
}

.close-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.members-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 320px;
    overflow-y: auto;
}

.member-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 5px;
    border-radius: 6px;
    cursor: pointer;
}

.member-row:hover {
    background: var(--input-focus);
}

.member-avatar {
    flex-shrink: 0;
    width: 38px;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
}

.member-avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.member-avatar-fallback {
    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 600;
}

.member-info {
    flex: 1;
    min-width: 0;
}

.member-name {
    color: var(--font-color);
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.loading-members,
.no-members {
    padding: 8px;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.dialog-overlay {
    position: fixed;
    z-index: 100;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgba(0, 0, 0, 0.7);
}

.invite-dialog {
    width: 420px;
    max-width: 100%;
    max-height: 80vh;
    max-height: 80dvh;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 20px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 8px 8px var(--main-color);
}

.dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.dialog-header h3 {
    margin: 0;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 18px;
}

.dialog-close {
    width: 28px;
    height: 28px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: transparent;
    color: var(--main-color);
    font-size: 16px;
    cursor: pointer;
}

.dialog-close:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.selected-count {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.invite-users-list {
    min-height: 100px;
    max-height: 320px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    overflow-y: auto;
}

.invite-user {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px;
    border: 2px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
}

.invite-user:hover {
    background: var(--input-focus);
}

.invite-user.selected {
    border-color: var(--main-color);
    background: var(--input-focus);
}

.invite-user-avatar {
    width: 38px;
    height: 38px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color-alt);
    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.invite-user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.invite-user-info {
    flex: 1;
    min-width: 0;
}

.invite-user-name {
    color: var(--font-color);
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.select-check {
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 4px;
    color: var(--bg-color);
    font-size: 13px;
    font-weight: 700;
}

.select-check.checked {
    background: var(--main-color);
}

.dialog-loading,
.dialog-empty,
.dialog-error,
.dialog-success {
    padding: 10px;
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.dialog-loading,
.dialog-empty {
    color: var(--font-color-sub);
}

.dialog-error {
    color: #e57373;
}

.dialog-success {
    color: #66bb6a;
}

.dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding-top: 4px;
}

.cancel-button,
.confirm-invite-button {
    padding: 8px 14px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
}

.cancel-button {
    background: transparent;
    color: var(--font-color);
}

.cancel-button:hover:not(:disabled) {
    background: var(--bg-color-alt);
}

.confirm-invite-button {
    background: var(--main-color);
    color: var(--bg-color);
}

.confirm-invite-button:hover:not(:disabled) {
    opacity: 0.85;
}

.cancel-button:disabled,
.confirm-invite-button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

.confirm-dialog {
    width: 380px;
    max-width: 100%;
    padding: 24px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 8px 8px var(--main-color);
    text-align: center;
}

.confirm-icon {
    width: 42px;
    height: 42px;
    margin: 0 auto 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 18px;
    font-weight: 700;
}

.confirm-dialog h3 {
    margin: 0 0 10px;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 18px;
}

.confirm-dialog p {
    margin: 0;
    color: var(--font-color-sub);
    font-size: 13px;
    line-height: 1.5;
}

.confirm-users {
    max-height: 120px;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 6px;
    margin: 16px 0;
    overflow-y: auto;
}

.confirm-user {
    padding: 5px 8px;
    border: 1px solid var(--bg-color-alt);
    border-radius: 5px;
    color: var(--font-color);
    font-size: 11px;
}

.confirm-actions {
    display: flex;
    justify-content: center;
    gap: 8px;
    margin-top: 18px;
}

@media (max-width: 650px) {
    .members-panel {
        right: 12px;
        left: 12px;
        width: auto;
    }

    .header-actions {
        gap: 4px;
    }

    .requests-button,
    .invite-button {
        padding: 5px 7px;
        font-size: 9px;
    }

    .invite-dialog,
    .confirm-dialog {
        width: 100%;
    }
}
</style>