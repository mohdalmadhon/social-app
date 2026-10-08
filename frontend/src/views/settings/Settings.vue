<script setup>
import { LIMITS } from '@/helpers/limits';
import { ref, onMounted } from 'vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import { addNotification } from '@/data/notifications';
import { activePage } from '@/data/chatState';
import { THEMES, getThemeCookie, setTheme } from '@/helpers/common/theme';
import {
    deleteAccount,
    changePreferences,
    getPreferences,
    getNotificationPreferences,
    changeNotificationPreference
} from '@/api/users/settings';

activePage.value = 'settings';

const TABS = [
    { value: 'general', label: 'General' },
    { value: 'preferences', label: 'Preferences' },
    { value: 'privacy', label: 'Privacy' },
    { value: 'notifications', label: 'Notifications' },
];

const NOTIFICATION_MODES = [
    { value: 'any', label: 'Everyone' },
    { value: 'friends', label: 'Friends only' },
    { value: 'none', label: 'Off' },
];

const NOTIFICATION_TYPES = [
    { type: 'follow', title: 'New followers', subtitle: 'When someone starts following you.' },
    { type: 'post_reaction', title: 'Post likes and dislikes', subtitle: 'When someone reacts to your post.' },
    { type: 'comment', title: 'Comments and replies', subtitle: 'When someone comments on your post or replies to your comment.' },
    { type: 'comment_like', title: 'Comment likes', subtitle: 'When someone likes your comment.' },
    { type: 'mention', title: 'Mentions and tags', subtitle: 'When someone mentions or tags you in a post or comment.' },
    { type: 'event', title: 'Group events', subtitle: 'When you are invited to an event or someone responds to yours.' },
    { type: 'message', title: 'Chat messages', subtitle: 'Pop-up alerts for new messages while you are on another page.' },
];

const VISIBILITY_OPTIONS = [
    {
        value: 'any',
        label: 'Everyone',
        description: 'Anyone who can view your profile can see it.',
        icon: '∞',
    },
    {
        value: 'none',
        label: 'No one',
        description: 'Only you can see it.',
        icon: '⊘',
    },
];

const ADDITIONAL_INFO_OPTIONS = [
    {
        value: 'friends',
        label: 'Friends',
        description: 'Only your friends can see your additional info.',
        icon: '♥',
    },
    {
        value: 'followers',
        label: 'Followers',
        description: 'Only people who follow you can see your additional info.',
        icon: '⇄',
    },
    {
        value: 'any',
        label: 'Any',
        description: 'Anyone who can view your profile can see your additional info.',
        icon: '∞',
    },
];

const PRIVACY_SECTIONS = [
    {
        type: 'email',
        title: 'Email',
        subtitle: 'Choose who can see your email on your profile.',
        options: VISIBILITY_OPTIONS,
    },
    {
        type: 'dob',
        title: 'Date of birth',
        subtitle: 'Choose who can see your date of birth on your profile.',
        options: VISIBILITY_OPTIONS,
    },
    {
        type: 'additional',
        title: 'Additional info',
        subtitle: 'Work, education, hobbies, interests, travel and social links.',
        options: ADDITIONAL_INFO_OPTIONS,
    },
];

const CHAT_OPTIONS = [
    {
        value: 'following-followers',
        label: 'Following or followers',
        description: 'Anyone you follow or who follows you can message you.',
        icon: '⇄',
    },
    {
        value: 'friends',
        label: 'Friends only',
        description: 'Only your friends can start a chat with you.',
        icon: '♥',
    },
    {
        value: 'following',
        label: 'Following only',
        description: 'Only people you follow can message you.',
        icon: '→',
    },
    {
        value: 'friends-following',
        label: 'Following or friends',
        description: 'People you follow and your friends can message you.',
        icon: '★',
    },
    {
        value: 'any',
        label: 'Any',
        description: 'Everyone can start a chat with you.',
        icon: '∞',
    },
    {
        value: 'none',
        label: 'None',
        description: 'Nobody can start a chat with you.',
        icon: '⊘',
    },
];

const GROUP_INVITE_OPTIONS = [
    {
        value: 'following',
        label: 'Following & followers',
        description: 'Friends, people you follow and your followers can invite you to groups.',
        icon: '→',
    },
    {
        value: 'friends',
        label: 'Friends',
        description: 'Only your friends can invite you to groups.',
        icon: '♥',
    },
    {
        value: 'none',
        label: 'None',
        description: 'Nobody can invite you to groups.',
        icon: '⊘',
    },
];

const activeTab = ref('general');

const selectedTheme = ref(getThemeCookie());
const selectedChat = ref('following-followers');
const savingChat = ref(false);
const allowPreviousSenders = ref(false);
const savingPreviousSenders = ref(false);
const selectedGroupInvite = ref('following');
const savingGroupInvite = ref(false);

const privacy = ref({
    email: 'none',
    dob: 'none',
    additional: 'any',
});
const savingPrivacy = ref(false);

const notificationPrefs = ref({});
const savingNotification = ref(false);

async function selectNotificationMode(type, mode) {
    if (savingNotification.value || mode === notificationPrefs.value[type]) {
        return;
    }

    const previous = notificationPrefs.value[type];
    notificationPrefs.value[type] = mode;
    savingNotification.value = true;

    try {
        await changeNotificationPreference(type, mode);
        addNotification('Notification preference updated', 'success');
    } catch (err) {
        notificationPrefs.value[type] = previous;
        addNotification(err.message || 'Could not update notification preference', 'error');
    } finally {
        savingNotification.value = false;
    }
}

onMounted(async () => {
    try {
        notificationPrefs.value = await getNotificationPreferences();
    } catch (err) {
        addNotification(err.message || 'Could not load notification preferences', 'error');
    }

    try {
        const prefs = await getPreferences();
        selectedChat.value = prefs.chat;
        allowPreviousSenders.value = Boolean(prefs.allowPreviousSenders);
        selectedGroupInvite.value = prefs.groupInvite || 'following';
        privacy.value = {
            email: prefs.email,
            dob: prefs.dob,
            additional: prefs.additionalInfo,
        };
    } catch (err) {
        addNotification(err.message || 'Could not load preferences', 'error');
    }
});

async function selectPrivacy(type, value) {
    if (savingPrivacy.value || value === privacy.value[type]) {
        return;
    }

    const previous = privacy.value[type];
    privacy.value[type] = value;
    savingPrivacy.value = true;

    try {
        await changePreferences(type, value);
        addNotification('Privacy updated', 'success');
    } catch (err) {
        privacy.value[type] = previous;
        addNotification(err.message || 'Could not update privacy', 'error');
    } finally {
        savingPrivacy.value = false;
    }
}

const showDeleteConfirm = ref(false);
const confirmText = ref('');
const deleting = ref(false);

function selectTheme(theme) {
    selectedTheme.value = theme;
    setTheme(theme);
    addNotification('Theme updated', 'success');
}

async function selectChat(value) {
    if (savingChat.value || value === selectedChat.value) {
        return;
    }

    const previous = selectedChat.value;
    selectedChat.value = value;
    savingChat.value = true;

    try {
        await changePreferences('chat', value);
        addNotification('Chat preference updated', 'success');
    } catch (err) {
        selectedChat.value = previous;
        addNotification(err.message || 'Could not update preferences', 'error');
    } finally {
        savingChat.value = false;
    }
}

async function togglePreviousSenders() {
    if (savingPreviousSenders.value) {
        return;
    }

    const previous = allowPreviousSenders.value;
    const next = !previous;

    allowPreviousSenders.value = next;
    savingPreviousSenders.value = true;

    try {
        await changePreferences('previoussenders', next ? '1' : '0');
        addNotification('Chat preference updated', 'success');
    } catch (err) {
        allowPreviousSenders.value = previous;
        addNotification(err.message || 'Could not update preferences', 'error');
    } finally {
        savingPreviousSenders.value = false;
    }
}

async function selectGroupInvite(value) {
    if (savingGroupInvite.value || value === selectedGroupInvite.value) {
        return;
    }

    const previous = selectedGroupInvite.value;
    selectedGroupInvite.value = value;
    savingGroupInvite.value = true;

    try {
        await changePreferences('groupinvite', value);
        addNotification('Group invite preference updated', 'success');
    } catch (err) {
        selectedGroupInvite.value = previous;
        addNotification(err.message || 'Could not update preferences', 'error');
    } finally {
        savingGroupInvite.value = false;
    }
}

function openDeleteConfirm() {
    showDeleteConfirm.value = true;
    confirmText.value = '';
}

function cancelDeleteConfirm() {
    showDeleteConfirm.value = false;
    confirmText.value = '';
}

async function confirmDeleteAccount() {
    if (confirmText.value.trim().toUpperCase() !== 'DELETE') {
        return;
    }

    deleting.value = true;

    try {
        await deleteAccount();
    } catch (err) {
        addNotification(err.message || 'Could not delete account', 'error');
    } finally {
        deleting.value = false;
    }
}
</script>

<template>
    <div class="facebook-layout">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="settings-page">
                <BackToHome />

                <div class="page-heading">
                    <p class="eyebrow">SETTINGS</p>
                    <h1>Settings</h1>
                </div>

                <div class="tabs" role="tablist">
                    <button
                        v-for="tab in TABS"
                        :key="tab.value"
                        type="button"
                        role="tab"
                        class="tab"
                        :class="{ active: activeTab === tab.value }"
                        :aria-selected="activeTab === tab.value"
                        @click="activeTab = tab.value"
                    >
                        {{ tab.label }}
                    </button>
                </div>

                <template v-if="activeTab === 'general'">
                    <section class="settings-section">
                        <h2>Appearance</h2>

                        <div class="theme-options">
                            <button
                                v-for="theme in THEMES"
                                :key="theme.value"
                                type="button"
                                class="theme-option"
                                :class="[`screen-${theme.value}`, { selected: selectedTheme === theme.value }]"
                                @click="selectTheme(theme.value)"
                            >
                                <span class="monitor">
                                    <span class="monitor-top">
                                        <span class="top-logo"></span>
                                        <span class="top-search"></span>
                                        <span class="top-avatar"></span>
                                    </span>
                                    <span class="monitor-body">
                                        <span class="monitor-side">
                                            <span class="side-item active"></span>
                                            <span class="side-item"></span>
                                            <span class="side-item"></span>
                                            <span class="side-item"></span>
                                        </span>
                                        <span class="monitor-content">
                                            <span class="mini-post">
                                                <span class="post-head">
                                                    <span class="post-avatar"></span>
                                                    <span class="post-name"></span>
                                                </span>
                                                <span class="line long"></span>
                                                <span class="line"></span>
                                                <span class="post-image"></span>
                                            </span>
                                        </span>
                                    </span>
                                </span>
                                <span class="monitor-stand"></span>

                                <span class="option-footer">
                                    <span class="option-name">{{ theme.label }}</span>
                                    <span class="radio">
                                        <span v-if="selectedTheme === theme.value"></span>
                                    </span>
                                </span>
                            </button>
                        </div>
                    </section>

                    <section class="settings-section danger-zone">
                        <h2>Danger zone</h2>

                        <div class="danger-row">
                            <div class="text-group">
                                <p class="option-title">Delete account</p>
                                <p class="option-subtitle">
                                    This permanently deletes your account, posts, messages and all
                                    related data. This cannot be undone.
                                </p>
                            </div>

                            <button type="button" class="delete-btn" @click="openDeleteConfirm">
                                Delete account
                            </button>
                        </div>

                        <div v-if="showDeleteConfirm" class="confirm-box">
                            <p class="confirm-text">
                                Type <strong>DELETE</strong> to permanently remove your account.
                            </p>

                            <input
                                v-model="confirmText"
                                :maxlength="LIMITS.deleteConfirm"
                                type="text"
                                placeholder="DELETE"
                                autocomplete="off"
                                @keyup.enter="confirmDeleteAccount"
                            />

                            <div class="confirm-actions">
                                <button type="button" class="cancel-btn" @click="cancelDeleteConfirm">
                                    Cancel
                                </button>

                                <button
                                    type="button"
                                    class="delete-btn"
                                    :disabled="confirmText.trim().toUpperCase() !== 'DELETE' || deleting"
                                    @click="confirmDeleteAccount"
                                >
                                    {{ deleting ? 'Deleting...' : 'Permanently delete' }}
                                </button>
                            </div>
                        </div>
                    </section>
                </template>

                <template v-else-if="activeTab === 'preferences'">
                    <section class="settings-section">
                        <div class="section-head">
                            <div>
                                <h2>Chat</h2>
                                <p class="section-subtitle">Choose who can start a chat with you.</p>
                            </div>
                            <span v-if="savingChat" class="saving">Saving...</span>
                        </div>

                        <div class="choice-grid">
                            <button
                                v-for="option in CHAT_OPTIONS"
                                :key="option.value"
                                type="button"
                                class="choice-card"
                                :class="{ selected: selectedChat === option.value }"
                                :disabled="savingChat"
                                @click="selectChat(option.value)"
                            >
                                <span class="choice-top">
                                    <span class="choice-icon">{{ option.icon }}</span>
                                    <span class="radio">
                                        <span v-if="selectedChat === option.value"></span>
                                    </span>
                                </span>
                                <span class="choice-name">{{ option.label }}</span>
                                <span class="choice-desc">{{ option.description }}</span>
                            </button>
                        </div>

                        <div class="switch-row">
                            <div class="switch-text">
                                <span class="switch-title">Let people who messaged me before message me again</span>
                                <span class="switch-desc">Anyone who has already messaged you can keep messaging you, whatever your chat setting above is.</span>
                            </div>

                            <button
                                type="button"
                                role="switch"
                                class="switch"
                                :class="{ on: allowPreviousSenders }"
                                :aria-checked="allowPreviousSenders"
                                :disabled="savingPreviousSenders"
                                @click="togglePreviousSenders"
                            >
                                <span class="switch-knob"></span>
                            </button>
                        </div>
                    </section>

                    <section class="settings-section">
                        <div class="section-head">
                            <div>
                                <h2>Group invites</h2>
                                <p class="section-subtitle">Choose who can invite you to groups.</p>
                            </div>
                            <span v-if="savingGroupInvite" class="saving">Saving...</span>
                        </div>

                        <div class="choice-grid">
                            <button
                                v-for="option in GROUP_INVITE_OPTIONS"
                                :key="option.value"
                                type="button"
                                class="choice-card"
                                :class="{ selected: selectedGroupInvite === option.value }"
                                :disabled="savingGroupInvite"
                                @click="selectGroupInvite(option.value)"
                            >
                                <span class="choice-top">
                                    <span class="choice-icon">{{ option.icon }}</span>
                                    <span class="radio">
                                        <span v-if="selectedGroupInvite === option.value"></span>
                                    </span>
                                </span>
                                <span class="choice-name">{{ option.label }}</span>
                                <span class="choice-desc">{{ option.description }}</span>
                            </button>
                        </div>
                    </section>
                </template>

                <template v-else-if="activeTab === 'privacy'">
                    <section
                        v-for="section in PRIVACY_SECTIONS"
                        :key="section.type"
                        class="settings-section"
                    >
                        <div class="section-head">
                            <div>
                                <h2>{{ section.title }}</h2>
                                <p class="section-subtitle">{{ section.subtitle }}</p>
                            </div>
                            <span v-if="savingPrivacy" class="saving">Saving...</span>
                        </div>

                        <div class="choice-grid">
                            <button
                                v-for="option in section.options"
                                :key="option.value"
                                type="button"
                                class="choice-card"
                                :class="{ selected: privacy[section.type] === option.value }"
                                :disabled="savingPrivacy"
                                @click="selectPrivacy(section.type, option.value)"
                            >
                                <span class="choice-top">
                                    <span class="choice-icon">{{ option.icon }}</span>
                                    <span class="radio">
                                        <span v-if="privacy[section.type] === option.value"></span>
                                    </span>
                                </span>
                                <span class="choice-name">{{ option.label }}</span>
                                <span class="choice-desc">{{ option.description }}</span>
                            </button>
                        </div>
                    </section>
                </template>

                <template v-else-if="activeTab === 'notifications'">
                    <section class="settings-section">
                        <div class="section-head">
                            <div>
                                <h2>Notifications</h2>
                                <p class="section-subtitle">
                                    Choose which notifications you receive and from whom.
                                </p>
                            </div>
                            <span v-if="savingNotification" class="saving">Saving...</span>
                        </div>

                        <div class="notification-pref-list">
                            <div
                                v-for="item in NOTIFICATION_TYPES"
                                :key="item.type"
                                class="notification-pref"
                            >
                                <div class="text-group">
                                    <p class="option-title">{{ item.title }}</p>
                                    <p class="option-subtitle">{{ item.subtitle }}</p>
                                </div>

                                <div class="mode-group">
                                    <button
                                        v-for="mode in NOTIFICATION_MODES"
                                        :key="mode.value"
                                        type="button"
                                        class="mode-btn"
                                        :class="{ selected: notificationPrefs[item.type] === mode.value }"
                                        :disabled="savingNotification"
                                        @click="selectNotificationMode(item.type, mode.value)"
                                    >
                                        {{ mode.label }}
                                    </button>
                                </div>
                            </div>
                        </div>
                    </section>
                </template>
            </main>
        </div>
    </div>
</template>

<style scoped>
.facebook-layout {
    min-height: 100vh;
    min-height: 100dvh;
}

.page-layout {
    display: flex;
    padding-top: 64px;
}

.settings-page {
    width: 100%;
    max-width: 900px;
    margin: 0 auto;
    padding: 25px 30px 60px;
    display: flex;
    flex-direction: column;
    gap: 30px;
}

.page-heading .eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.page-heading h1 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 36px;
}

.settings-section h2 {
    margin: 0 0 14px;
    font-family: "Liter", serif;
    font-size: 20px;
    color: var(--font-color);
}

.tabs {
    display: inline-flex;
    align-self: flex-start;
    gap: 4px;
    padding: 4px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.tab {
    padding: 9px 22px;
    border: none;
    border-radius: 5px;
    background: transparent;
    color: var(--font-color);
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
}

.tab:hover:not(.active) {
    background: var(--page-background);
}

.tab.active {
    background: var(--input-focus);
    color: #fff;
}

.section-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
}

.section-head h2 {
    margin-bottom: 6px;
}

.section-subtitle {
    margin: 0 0 18px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.saving {
    padding: 4px 10px;
    border: 2px dashed var(--main-color);
    border-radius: 20px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.choice-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
    gap: 16px;
}

.choice-card {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    padding: 16px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
    transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.choice-card:hover:not(:disabled):not(.selected) {
    transform: translate(-2px, -2px);
    box-shadow: 4px 4px var(--main-color);
}

.choice-card.selected {
    background: var(--input-focus);
    color: #fff;
    box-shadow: 4px 4px var(--main-color);
}

.switch-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-top: 18px;
    padding: 16px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
}

.switch-text {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
}

.switch-title {
    color: var(--font-color);
    font-size: 13px;
    font-weight: 700;
}

.switch-desc {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    line-height: 1.5;
}

.switch {
    position: relative;
    flex-shrink: 0;
    width: 52px;
    height: 28px;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 999px;
    background: var(--page-background);
    cursor: pointer;
    transition: background 0.15s ease;
}

.switch.on {
    background: var(--input-focus);
}

.switch:disabled {
    cursor: not-allowed;
    opacity: 0.75;
}

.switch-knob {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 20px;
    height: 20px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: #fff;
    transition: transform 0.15s ease;
}

.switch.on .switch-knob {
    transform: translateX(24px);
}

.choice-card:disabled {
    cursor: not-allowed;
    opacity: 0.75;
}

.choice-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    margin-bottom: 8px;
}

.choice-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border: 2px solid currentColor;
    border-radius: 10px;
    font-size: 18px;
    line-height: 1;
}

.choice-name {
    font-weight: 700;
    font-size: 15px;
}

.choice-desc {
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    line-height: 1.5;
    opacity: 0.8;
}

.theme-options {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
    gap: 16px;
}

.theme-option {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0;
    padding: 14px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    cursor: pointer;
}

.theme-option.selected {
    background: var(--input-focus);
    color: #fff;
    box-shadow: 3px 3px var(--main-color);
}

.screen-light {
    --s-accent: #5b9cff;
    --s-bg: #f4f4f2;
    --s-bar: #dcdcd8;
    --s-side: #e6e6e2;
    --s-line: #b5b5b0;
    --s-card: #ffffff;
}

.screen-dark {
    --s-accent: #8a7dff;
    --s-bg: #121212;
    --s-bar: #1f1f1f;
    --s-side: #1a1a1a;
    --s-line: #444;
    --s-card: #262626;
}

.screen-blue {
    --s-accent: #ffffff;
    --s-bg: #5b9cff;
    --s-bar: #3f7fe0;
    --s-side: #4d8ff0;
    --s-line: #a9ccff;
    --s-card: #8bb8ff;
}

.screen-pink {
    --s-accent: #ffffff;
    --s-bg: #e88ca8;
    --s-bar: #d06f8d;
    --s-side: #dc7f9b;
    --s-line: #f5c2d2;
    --s-card: #f2a9bf;
}

.monitor {
    width: 100%;
    aspect-ratio: 16 / 10;
    display: flex;
    flex-direction: column;
    border: 3px solid var(--main-color);
    border-radius: 6px;
    overflow: hidden;
    background: var(--s-bg);
}

.monitor-top {
    flex: 0 0 16%;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 6px;
    background: var(--s-bar);
}

.top-logo {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--s-accent);
}

.top-search {
    flex: 1;
    max-width: 45%;
    height: 6px;
    margin: 0 auto 0 4px;
    border-radius: 3px;
    background: var(--s-line);
    opacity: 0.6;
}

.top-avatar {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--s-line);
}

.monitor-body {
    flex: 1;
    display: flex;
    min-height: 0;
}

.monitor-side {
    flex: 0 0 22%;
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding: 6px 5px;
    background: var(--s-side);
}

.side-item {
    height: 5px;
    border-radius: 3px;
    background: var(--s-line);
    opacity: 0.6;
}

.side-item.active {
    background: var(--s-accent);
    opacity: 1;
}

.monitor-content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 6px;
    min-width: 0;
}

.mini-post {
    width: 78%;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 5px;
    border-radius: 4px;
    background: var(--s-card);
    overflow: hidden;
}

.post-head {
    display: flex;
    align-items: center;
    gap: 4px;
}

.post-avatar {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--s-accent);
}

.post-name {
    width: 40%;
    height: 4px;
    border-radius: 2px;
    background: var(--s-line);
}

.line {
    height: 4px;
    width: 55%;
    border-radius: 3px;
    background: var(--s-line);
}

.line.long {
    width: 90%;
}

.post-image {
    flex: 1;
    min-height: 0;
    border-radius: 3px;
    background: var(--s-side);
}

.monitor-stand {
    width: 28%;
    height: 8px;
    background: var(--main-color);
}

.monitor-stand::after {
    content: "";
    display: block;
    width: 160%;
    height: 4px;
    margin: 8px 0 0 -30%;
    border-radius: 2px;
    background: var(--main-color);
}

.option-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    margin-top: 14px;
}

.option-name {
    font-weight: 600;
    font-size: 14px;
}

.radio {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border: 2px solid currentColor;
    border-radius: 50%;
}

.radio span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
}

.danger-zone {
    border: 2px solid #d9534f;
    border-radius: 6px;
    padding: 20px;
    background: var(--bg-color);
}

.danger-zone h2 {
    color: #d9534f;
}

.danger-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    flex-wrap: wrap;
}

.text-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.option-title {
    margin: 0;
    color: var(--font-color);
    font-weight: 600;
    font-size: 14px;
}

.option-subtitle {
    margin: 0;
    max-width: 480px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.delete-btn {
    flex-shrink: 0;
    padding: 10px 18px;
    border: 2px solid #d9534f;
    border-radius: 5px;
    background: #d9534f;
    color: #fff;
    font-weight: 600;
    font-size: 13px;
}

.delete-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.confirm-box {
    margin-top: 18px;
    padding-top: 18px;
    border-top: 2px dashed #d9534f;
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.confirm-text {
    margin: 0;
    font-size: 13px;
    color: var(--font-color);
}

.confirm-box input {
    width: 100%;
    max-width: 260px;
    padding: 10px 12px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--font-color);
}

.confirm-actions {
    display: flex;
    gap: 10px;
}

.cancel-btn {
    padding: 10px 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    font-weight: 600;
    font-size: 13px;
}

.notification-pref-list {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.notification-pref {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
}

.mode-group {
    display: flex;
    flex-shrink: 0;
    gap: 6px;
}

.mode-btn {
    padding: 7px 12px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
    cursor: pointer;
}

.mode-btn.selected {
    background: var(--input-focus);
    color: #fff;
    box-shadow: 2px 2px var(--main-color);
}

.mode-btn:disabled {
    cursor: not-allowed;
    opacity: 0.75;
}

@media (max-width: 650px) {
    .notification-pref {
        flex-direction: column;
        align-items: flex-start;
    }
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .settings-page {
        padding: 20px 15px 50px;
    }
}
</style>