<script setup>
import { getFriends } from '@/api/common/friends';
import { getGroupMembers } from '@/api/posts/groups';
import { activePage } from '@/data/chatState';
import { addNotification } from '@/data/notifications';
import { ref } from 'vue';

const props = defineProps({
    modelValue: {
        type: Array,
        default: () => []
    }
});

const emit = defineEmits(['update:modelValue']);

const search = ref('');
const friends = ref([]);
let timeout;

function isGroupPage() {
    return activePage.value?.startsWith('group:');
}

function getCurrentGroupID() {
    if (!isGroupPage()) {
        return null;
    }

    return Number(activePage.value.split(':')[1]);
}

function handleSearchFriends() {
    clearTimeout(timeout);

    timeout = setTimeout(async () => {
        const value = search.value.trim();

        if (!value) {
            friends.value = [];
            return;
        }

        try {
            let result;

            if (isGroupPage()) {
                const groupID = getCurrentGroupID();

                if (!groupID) {
                    friends.value = [];
                    return;
                }

                result = await getGroupMembers(groupID, value);
            } else {
                result = await getFriends(value);
            }

            friends.value = Object.entries(result.data)
                .map(([id, person]) => ({
                    id: Number(id),
                    ...person
                }))
                .filter(
                    person =>
                        !props.modelValue.some(
                            selected => selected.id === person.id
                        )
                );
        } catch (err) {
            console.error(err);
            addNotification('could not fetch people');
            friends.value = [];
        }
    }, 500);
}

function addPerson(person) {
    if (!person) {
        return;
    }

    if (
        props.modelValue.some(
            existing => existing.id === person.id
        )
    ) {
        return;
    }

    emit('update:modelValue', [
        ...props.modelValue,
        person
    ]);

    friends.value = friends.value.filter(
        friend => friend.id !== person.id
    );

    search.value = '';
    friends.value = [];
}

function removePerson(id) {
    emit(
        'update:modelValue',
        props.modelValue.filter(
            person => person.id !== id
        )
    );
}
</script>

<template>
    <div class="tag-people">
        <label for="post-tag" style="font-weight: 900;">
            <strong>Tag people</strong>
        </label>

        <label for="post-tag" style="font-size: 8px;">
            You can only tag people you are friends with
        </label>

        <div class="tag-row">
            <div class="search-container">
                <input
                    id="post-tag"
                    type="text"
                    placeholder="Search for people"
                    v-model="search"
                    @input="handleSearchFriends"
                >

                <div
                    v-if="friends.length"
                    class="friend-list"
                >
                    <button
                        v-for="person in friends"
                        :key="person.id"
                        type="button"
                        class="friend"
                        @click="addPerson(person)"
                    >
                        <div class="friend-info">
                            <img
                                v-if="person.avatar"
                                :src="`/uploads/${person.avatar}`"
                                alt=""
                                class="friend-avatar"
                            >

                            <span>
                                {{ person.firstName }}
                                {{ person.lastName }}
                            </span>
                        </div>
                    </button>
                </div>
            </div>
        </div>

        <div
            v-if="modelValue.length"
            class="tag-list"
        >
            <span
                v-for="person in modelValue"
                :key="person.id"
                class="tag-chip"
            >
                <div class="friend-info">
                    <img
                        v-if="person.avatar"
                        :src="`/uploads/${person.avatar}`"
                        alt=""
                        class="friend-avatar"
                    >

                    <span>
                        {{ person.firstName }}
                        {{ person.lastName }}
                    </span>
                </div>

                <button
                    type="button"
                    @click="removePerson(person.id)"
                >
                    ×
                </button>
            </span>
        </div>
    </div>
</template>

<style scoped>
.friend-info {
    display: flex;
    align-items: center;
    gap: 10px;
}

.friend-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
}

.tag-people {
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

.tag-row {
    display: flex;
    gap: 10px;
}

.search-container {
    position: relative;
    width: 100%;
}

.tag-row input {
    width: 100%;
    box-sizing: border-box;
    padding: 12px 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    outline: none;
    background: var(--bg-color);
    color: var(--font-color);
    transition: box-shadow 0.15s ease;
}

.tag-row input:focus {
    box-shadow: 3px 3px var(--main-color);
}

.friend-list {
    position: absolute;
    z-index: 10;
    top: calc(100% + 5px);
    left: 0;
    width: 100%;
    max-height: 200px;
    overflow-y: auto;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.friend {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px;
    border: 0;
    border-bottom: 1px solid var(--main-color);
    background: transparent;
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
}

.friend:last-child {
    border-bottom: 0;
}

.friend:hover {
    background: var(--input-focus);
}

.friend img {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
}

.tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 4px;
}

.tag-chip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px 7px 14px;
    border: 2px solid var(--main-color);
    border-radius: 20px;
    background: var(--page-background);
    color: var(--font-color);
    font-size: 12px;
    font-weight: 600;
}

.tag-chip button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border: 0;
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-size: 12px;
    line-height: 1;
    cursor: pointer;
}
</style>