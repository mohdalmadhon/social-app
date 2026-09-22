<script setup>
import { ref, watch } from 'vue';
import { getPostGroups } from '@/api/common/friends';

const props = defineProps({
    modelValue: {
        type: Number,
        default: 'public'
    },
    groupId: {
        type: Number,
        default: null
    }
});

const emit = defineEmits(['update:modelValue', 'update:groupId', 'update:groupIDValue']);

const groups = ref([]);
const loadingGroups = ref(false);
const groupsLoaded = ref(false);

// Holds the value to persist: 0 for public, -1 for private, or the
// selected group's ID when privacy is set to "group".
const groupIDValue = ref(0);

watch(() => props.modelValue, (value) => {
    if (value === 'public') {
        groupIDValue.value = 0;
    } else if (value === 'followers') {
        groupIDValue.value = 0;
    } else if (value === 'group') {
        groupIDValue.value = props.groupId ?? null;
    } else {
        // private toggle (-1)
        groupIDValue.value = -1;
    }
}, { immediate: true });

watch(() => props.groupId, (value) => {
    if (props.modelValue === 'group') {
        groupIDValue.value = value;
    }
});

watch(groupIDValue, (value) => {
    emit('update:groupIDValue', value);
}, { immediate: true });

function togglePrivate(event) {
    emit(
        'update:modelValue',
        event.target.checked ? -1 : 0
    );
}

function selectPrivacy(value) {
    if (value == 'group' && !groupsLoaded.value) {
        getGroups()
    }
    emit('update:modelValue', value);
}

function selectGroup(id) {
    groupIDValue.value = id;
    emit('update:groupId', id);
}

defineExpose({ groupIDValue });

async function getGroups() {
    loadingGroups.value = true;

    try {
        const result = await getPostGroups();
        groups.value = Array.isArray(result) ? result : Object.values(result || {});
        groupsLoaded.value = true;
    } catch {
        groups.value = [];
    } finally {
        loadingGroups.value = false;
    }
}
</script>

<template>
    <div class="privacy-container">
        <div class="hide-comments-row">
            <div class="text-group">
                <p class="option-title">Post Privacy</p>
                <p class="option-subtitle">Manage post visibility</p>
            </div>

            <label class="switch">
                <input type="checkbox" :checked="modelValue !== 'public'" @change="togglePrivate">
                <span class="slider"></span>
            </label>
        </div>

        <div v-if="modelValue !== 'public'" class="privacy-options">
            <button type="button" class="privacy-option" :class="{ selected: modelValue === 'followers' }"
                @click="selectPrivacy('followers')">
                <div class="option-icon">👥</div>

                <div class="option-info">
                    <p class="option-name">Followers only</p>
                    <p class="option-description">
                        Only your followers can see this post
                    </p>
                </div>

                <span class="radio">
                    <span v-if="modelValue === 'followers'"></span>
                </span>
            </button>

            <button type="button" class="privacy-option" :class="{ selected: modelValue === 'group' }"
                @click="selectPrivacy('group')">
                <div class="option-icon">👥</div>

                <div class="option-info">
                    <p class="option-name">Group</p>
                    <p class="option-description">
                        Only members of a selected group can see this post
                    </p>
                </div>

                <span class="radio">
                    <span v-if="modelValue === 'group'"></span>
                </span>
            </button>

            <div v-if="modelValue === 'group'" class="group-list">
                <div v-if="loadingGroups" class="group-state-message">
                    Loading groups...
                </div>

                <div v-else-if="!groups.length" class="group-state-message">
                    No groups found.
                </div>

                <button v-for="group in groups" v-else :key="group.ID" type="button" class="group-item"
                    :class="{ selected: groupIDValue === group.ID }" @click="selectGroup(group.ID)">
                    <span class="group-name">{{ group.name }}</span>
                
                    <span class="radio radio--small">
                        <span v-if="groupIDValue === group.ID"></span>
                    </span>
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.privacy-container {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.hide-comments-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    padding: 16px 18px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
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
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.switch {
    position: relative;
    flex-shrink: 0;
    width: 50px;
    height: 24px;
}

.switch input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
}

.slider {
    position: absolute;
    inset: 0;
    border: 2px solid var(--main-color);
    border-radius: 20px;
    background: var(--page-background);
    cursor: pointer;
    transition: 0.2s ease;
}

.slider::before {
    position: absolute;
    content: "";
    width: 16px;
    height: 16px;
    left: 2px;
    top: 2px;
    border-radius: 50%;
    background: var(--main-color);
    transition: 0.2s ease;
}

.switch input:checked+.slider {
    background: var(--input-focus);
}

.switch input:checked+.slider::before {
    transform: translateX(26px);
    background: white;
}

.privacy-options {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.privacy-option {
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
    transition: 0.15s ease;
}

.privacy-option:hover {
    transform: translateX(2px);
}

.privacy-option.selected {
    background: var(--input-focus);
    color: white;
}

.option-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    font-size: 14px;
}

.option-info {
    flex: 1;
}

.option-name {
    margin: 0 0 3px;
    font-size: 13px;
    font-weight: 700;
}

.option-description {
    margin: 0;
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    opacity: 0.75;
}

.radio {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    flex-shrink: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
}

.radio span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--input-focus);
}

.group-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-left: 20px;
    padding: 10px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
}

.group-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    width: 100%;
    padding: 10px 12px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
    transition: 0.15s ease;
}

.group-item:hover {
    transform: translateX(2px);
}

.group-item.selected {
    background: var(--input-focus);
    color: white;
}

.group-name {
    font-size: 12px;
    font-weight: 700;
}

.radio--small {
    width: 14px;
    height: 14px;
}

.radio--small span {
    width: 6px;
    height: 6px;
}

.group-state-message {
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    color: var(--font-color-sub);
    padding: 4px 2px;
}
</style>