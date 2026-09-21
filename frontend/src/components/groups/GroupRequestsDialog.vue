```vue
<script setup>
import { ref, watch } from 'vue';
import { throttle } from '@/helpers/throttle';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    groupId: {
        type: [Number, String],
        required: true
    }
});

const emit = defineEmits(['close']);

const requests = ref([]);
const offset = ref(0);
const hasMore = ref(true);
const loading = ref(false);
const processing = ref(new Set());
const requestsScroll = ref(null);

const limit = 10;

async function getRequests(reset = false) {
    if (loading.value) return;
    if (!reset && !hasMore.value) return;

    if (reset) {
        offset.value = 0;
        hasMore.value = true;
        requests.value = [];
    }

    loading.value = true;

    try {
        const response = await fetch(
            `/api/groups/requests?groupID=${Number(props.groupId)}&offset=${offset.value}`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const result = await response.json();
        console.log(result)
        if (!response.ok || !result.status) {
            throw new Error(
                result.message || 'Could not get requests'
            );
        }

        const newRequests = result.data || [];

        const existingIds = new Set(
            requests.value.map(user => Number(user.ID))
        );

        const filteredRequests = newRequests.filter(
            user => !existingIds.has(Number(user.ID))
        );

        requests.value.push(...filteredRequests);

        offset.value += newRequests.length;

        if (newRequests.length < limit) {
            hasMore.value = false;
        }
    } catch (error) {
        console.error(error);
    } finally {
        loading.value = false;
    }
}

async function handleRequest(userID, code) {
    if (processing.value.has(userID)) {
        return;
    }

    processing.value.add(userID);
    processing.value = new Set(processing.value);

    try {
        const response = await fetch(
            '/api/groups/requests',
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    groupID: Number(props.groupId),
                    userID: Number(userID),
                    code
                })
            }
        );

        const result = await response.json();

        if (!response.ok || !result.status) {
            throw new Error(
                result.message || 'Could not handle request'
            );
        }

        requests.value = requests.value.filter(
            user => Number(user.ID) !== Number(userID)
        );
    } catch (error) {
        console.error(error);
    } finally {
        processing.value.delete(userID);
        processing.value = new Set(processing.value);
    }
}

function acceptRequest(userID) {
    handleRequest(userID, 1);
}

function rejectRequest(userID) {
    handleRequest(userID, -1);
}

const handleScroll = throttle(() => {
    const container = requestsScroll.value;

    if (!container) return;

    const distance =
        container.scrollHeight -
        container.scrollTop -
        container.clientHeight;

    if (distance < 150) {
        getRequests();
    }
}, 300);

function close() {
    emit('close');
}

watch(
    () => props.show,
    value => {
        if (value) {
            getRequests(true);
        }
    }
);
</script>

<template>
    <div
        v-if="show"
        class="dialog-overlay"
        @click.self="close"
    >
        <div class="requests-dialog">
            <div class="dialog-header">
                <h2>Group Requests</h2>

                <button
                    type="button"
                    class="close-button"
                    @click="close"
                >
                    ×
                </button>
            </div>

            <div
                ref="requestsScroll"
                class="requests-scroll"
                @scroll="handleScroll"
            >
                <div
                    v-if="loading && !requests.length"
                    class="message"
                >
                    Loading requests...
                </div>

                <div
                    v-else-if="!requests.length"
                    class="message"
                >
                    No pending requests.
                </div>

                <div
                    v-for="user in requests"
                    :key="user.ID"
                    class="request-item"
                >
                    <div class="user-avatar">
                        <img
                            v-if="user.avatar"
                            :src="`/uploads/${user.avatar}`"
                            :alt="`${user.firstName} ${user.lastName}`"
                        >

                    </div>

                    <div class="user-info">
                        <div class="user-name">
                            {{ user.firstName }} {{ user.lastName }}
                        </div>
                    </div>

                    <div class="request-actions">
                        <button
                            type="button"
                            class="accept-button"
                            :disabled="processing.has(user.ID)"
                            @click="acceptRequest(user.ID)"
                        >
                            Accept
                        </button>

                        <button
                            type="button"
                            class="reject-button"
                            :disabled="processing.has(user.ID)"
                            @click="rejectRequest(user.ID)"
                        >
                            Reject
                        </button>
                    </div>
                </div>

                <div
                    v-if="loading && requests.length"
                    class="message"
                >
                    Loading more...
                </div>

                <div
                    v-if="!hasMore && requests.length"
                    class="message"
                >
                    No more requests.
                </div>
            </div>
        </div>
    </div>
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

.requests-dialog {
    width: min(620px, 100%);
    max-height: 80vh;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    color: var(--font-color);
    box-shadow: 8px 8px var(--main-color);
}

.dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 18px 20px;
    border-bottom: 2px solid var(--main-color);
}

.dialog-header h2 {
    margin: 0;
    font-family: "JetBrains Mono", monospace;
    font-size: 16px;
}

.close-button {
    width: 34px;
    height: 34px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
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

.requests-scroll {
    max-height: calc(80vh - 75px);
    overflow-y: auto;
    padding: 14px;
}

.request-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    margin-bottom: 10px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
}

.user-avatar {
    width: 46px;
    height: 46px;
    flex: 0 0 46px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: var(--bg-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 700;
}

.user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.user-info {
    min-width: 0;
    flex: 1;
}

.user-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 600;
}

.request-actions {
    display: flex;
    gap: 8px;
}

.accept-button,
.reject-button {
    min-width: 75px;
    padding: 7px 10px;
    border: 2px solid var(--main-color);
    border-radius: 4px;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
}

.accept-button {
    background: var(--main-color);
    color: var(--bg-color);
}

.reject-button {
    background: var(--bg-color);
    color: var(--font-color);
}

.accept-button:hover {
    opacity: 0.85;
}

.reject-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.accept-button:disabled,
.reject-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.message {
    padding: 24px 10px;
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    opacity: 0.7;
}

@media (max-width: 600px) {
    .dialog-overlay {
        padding: 10px;
    }

    .requests-dialog {
        width: 100%;
        box-shadow: 5px 5px var(--main-color);
    }

    .request-item {
        align-items: flex-start;
        flex-wrap: wrap;
    }

    .user-info {
        padding-top: 4px;
    }

    .request-actions {
        width: 100%;
        margin-left: 58px;
    }

    .accept-button,
    .reject-button {
        flex: 1;
    }
}
</style>