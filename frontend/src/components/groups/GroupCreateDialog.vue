<script setup>
import { reactive, ref } from 'vue';

import AvatarPicker from '@/components/groups/AvatarPicker.vue';
import AddMembers from '@/components/groups/AddMembers.vue';

import { addNotification } from '@/data/notifications';
import { createGroup } from '@/api/groups/groups';

defineProps({
    show: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['close', 'created']);

const titleLimit = 50;
const descriptionLimit = 300;

const group = reactive({
    title: '',
    description: '',
    avatar: null,
    members: []
});

const creating = ref(false);
const titleTouched = ref(false);

function resetGroup() {
    group.title = '';
    group.description = '';
    group.avatar = null;
    group.members = [];
    titleTouched.value = false;
}

function validateGroup() {
    const title = group.title.trim();

    if (!title) {
        return 'Group name is required';
    }

    if (title.length > titleLimit) {
        return `Group name cannot exceed ${titleLimit} characters`;
    }

    if (group.description.length > descriptionLimit) {
        return `Description cannot exceed ${descriptionLimit} characters`;
    }

    return null;
}

function closeDialog() {
    if (creating.value) {
        return;
    }

    emit('close');
    resetGroup();
}

async function handleSubmit() {
    titleTouched.value = true;

    const error = validateGroup();

    if (error) {
        addNotification(error, 'error');
        return;
    }

    creating.value = true;

    try {
        const userIDs = group.members
            .map(member => Number(member.ID))
            .filter(id => Number.isInteger(id) && id > 0);

        console.log('Users being sent:', userIDs);

        const result = await createGroup({
            title: group.title.trim(),
            description: group.description.trim(),
            avatar: group.avatar,
            users: userIDs
        });

        addNotification('Group created!', 'success');
        emit('created', result.data);
        creating.value = false;
        closeDialog();
    } catch (err) {
        addNotification(err.message || 'Could not create group', 'error');
    } finally {
        creating.value = false;
    }
}
</script>

<template>
    <div v-if="show" class="group-modal-overlay" @click.self="closeDialog">
        <form class="group-modal" @submit.prevent="handleSubmit">
            <div class="modal-scroll">
                <div class="modal-header">
                    <h1>Create group</h1>

                    <button type="button" class="close-button" :disabled="creating" @click="closeDialog">
                        ×
                    </button>
                </div>

                <AvatarPicker v-model="group.avatar" />

                <div class="field-group">
                    <label for="group-title">Group name</label>

                    <input id="group-title" v-model="group.title" type="text" :maxlength="titleLimit"
                        placeholder="Group name" autocomplete="off" @blur="titleTouched = true">

                    <p v-if="titleTouched && !group.title.trim()" class="validation-error">
                        Group name is required
                    </p>
                </div>

                <div class="field-group">
                    <label for="group-description">Description</label>

                    <textarea id="group-description" v-model="group.description" :maxlength="descriptionLimit"
                        placeholder="What is this group about?"></textarea>
                </div>

                <AddMembers v-model="group.members" />
            </div>

            <div class="navigation-buttons">
                <button class="cancel-button" type="button" :disabled="creating" @click="closeDialog">
                    Cancel
                </button>

                <button class="submit-button" type="submit" :disabled="creating">
                    {{ creating ? 'Creating...' : 'Create group' }}
                </button>
            </div>
        </form>
    </div>
</template>

<style scoped>
.group-modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 30px;
    background: rgba(0, 0, 0, 0.55);
    overflow-y: auto;
}

.group-modal {
    width: 100%;
    max-width: 600px;
    max-height: calc(100vh - 60px);
    max-height: calc(100dvh - 60px);
    display: flex;
    flex-direction: column;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    box-sizing: border-box;
    overflow: hidden;
}

.modal-scroll {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 22px;
    padding: 35px 40px;
}

.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    padding-bottom: 18px;
    border-bottom: 2px solid var(--main-color);
}

.modal-header h1 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 28px;
}

.close-button {
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--main-color);
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
}

.close-button:hover:not(:disabled) {
    background: var(--input-focus);
    color: white;
}

.close-button:disabled {
    opacity: 0.5;
    cursor: default;
}

.field-group {
    display: flex;
    flex-direction: column;
    gap: 7px;
}

label {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

input,
textarea {
    width: 100%;
    padding: 12px 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    outline: none;
    background: var(--bg-color);
    color: var(--font-color);
    box-sizing: border-box;
    transition: box-shadow 0.15s ease;
}

textarea {
    min-height: 100px;
    resize: vertical;
    line-height: 1.5;
    font-family: inherit;
}

input:focus,
textarea:focus {
    box-shadow: 3px 3px var(--main-color);
}

.validation-error {
    margin: 0;
    color: #d93025;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
}

.navigation-buttons {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
    padding: 18px 40px;
    border-top: 2px solid var(--main-color);
    background: var(--bg-color);
}

.cancel-button,
.submit-button {
    padding: 13px 30px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    box-shadow: 4px 4px var(--main-color);
    cursor: pointer;
    transition:
        transform 0.1s ease,
        box-shadow 0.1s ease;
}

.cancel-button {
    background: var(--bg-color);
    color: var(--main-color);
}

.submit-button {
    margin-left: auto;
    background: var(--input-focus);
    color: white;
}

.cancel-button:hover:not(:disabled),
.submit-button:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.cancel-button:active:not(:disabled),
.submit-button:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.cancel-button:disabled,
.submit-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

@media (max-width: 800px) {
    .group-modal-overlay {
        padding: 15px;
    }

    .group-modal {
        max-height: calc(100vh - 30px);
        max-height: calc(100dvh - 30px);
    }

    .modal-scroll {
        padding: 28px 24px;
    }

    .navigation-buttons {
        padding: 16px 24px;
    }
}

@media (max-width: 550px) {
    .group-modal {
        border-radius: 12px;
    }

    .modal-scroll {
        padding: 25px 20px;
    }

    .navigation-buttons {
        flex-direction: column-reverse;
        padding: 14px 20px;
    }

    .cancel-button,
    .submit-button {
        width: 100%;
        margin-left: 0;
        text-align: center;
    }

    .modal-header h1 {
        font-size: 24px;
    }
}
</style>
