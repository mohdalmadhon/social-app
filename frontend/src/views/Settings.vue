<script setup>
import { ref } from 'vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import { addNotification } from '@/data/notifications';
import { activePage } from '@/data/chatState';
import { THEMES, getThemeCookie, setTheme } from '@/helpers/common/theme';
import { deleteAccount } from '@/api/users/settings';

activePage.value = 'settings';

const selectedTheme = ref(getThemeCookie());

const showDeleteConfirm = ref(false);
const confirmText = ref('');
const deleting = ref(false);

function selectTheme(theme) {
    selectedTheme.value = theme;
    setTheme(theme);
    addNotification('Theme updated', 'success');
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
                <div class="page-heading">
                    <p class="eyebrow">SETTINGS</p>
                    <h1>Settings</h1>
                </div>

                <section class="settings-section">
                    <h2>Appearance</h2>

                    <div class="theme-options">
                        <button
                            v-for="theme in THEMES"
                            :key="theme.value"
                            type="button"
                            class="theme-option"
                            :class="{ selected: selectedTheme === theme.value }"
                            @click="selectTheme(theme.value)"
                        >
                            <span class="swatch" :class="`swatch-${theme.value}`"></span>

                            <span class="option-info">
                                <span class="option-name">{{ theme.label }}</span>
                            </span>

                            <span class="radio">
                                <span v-if="selectedTheme === theme.value"></span>
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
            </main>
        </div>
    </div>
</template>

<style scoped>
.facebook-layout {
    min-height: 100vh;
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

.theme-options {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.theme-option {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    padding: 14px 16px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
}

.theme-option.selected {
    background: var(--input-focus);
    color: #fff;
    box-shadow: 3px 3px var(--main-color);
}

.swatch {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 2px solid var(--main-color);
}

.swatch-light {
    background: #f4f4f2;
}

.swatch-dark {
    background: #121212;
}

.swatch-blue {
    background: #5b9cff;
}

.swatch-pink {
    background: #e88ca8;
}

.option-info {
    flex: 1;
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

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .settings-page {
        padding: 20px 15px 50px;
    }
}
</style>
