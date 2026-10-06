<script setup>
import { onUnmounted, ref, watch } from 'vue';
import { fetchGroupEventVotes } from '@/api/groups/events';
import { throttle } from '@/helpers/throttle';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    eventId: {
        type: [Number, String],
        default: null
    },
    eventTitle: {
        type: String,
        default: ''
    }
});

const emit = defineEmits(['close']);

const votesLimit = 10;

const activeResponse = ref(1);
const voters = ref([]);
const votesOffset = ref(0);
const votesHasMore = ref(true);
const loadingVotes = ref(false);
const votesError = ref('');
const votesContainer = ref(null);

let votesRequestID = 0;

async function loadVoters(reset = false) {
    if (!props.eventId) return;
    if (loadingVotes.value) return;
    if (!reset && !votesHasMore.value) return;

    const requestID = ++votesRequestID;

    loadingVotes.value = true;
    votesError.value = '';

    if (reset) {
        votesOffset.value = 0;
        votesHasMore.value = true;
        voters.value = [];
    }

    try {
        const data = await fetchGroupEventVotes(
            props.eventId,
            activeResponse.value,
            votesOffset.value,
            votesLimit
        );

        if (requestID !== votesRequestID) return;

        voters.value = [...voters.value, ...data];
        votesOffset.value += data.length;

        if (data.length < votesLimit) {
            votesHasMore.value = false;
        }
    } catch (err) {
        if (requestID !== votesRequestID) return;

        votesError.value = err.message || 'Could not load votes';
        votesHasMore.value = false;
    } finally {
        if (requestID === votesRequestID) {
            loadingVotes.value = false;
        }
    }
}

const handleVotesScroll = throttle(() => {
    const container = votesContainer.value;

    if (!container) return;

    const distance =
        container.scrollHeight - container.scrollTop - container.clientHeight;

    if (distance < 120) {
        loadVoters();
    }
}, 300);

function selectResponse(value) {
    if (activeResponse.value === value) return;

    activeResponse.value = value;
    loadVoters(true);
}

function closeVotesDialog() {
    votesRequestID++;
    emit('close');
}

function voterName(voter) {
    return `${voter.user?.firstName ?? ''} ${voter.user?.lastName ?? ''}`.trim();
}

watch(
    () => [props.show, props.eventId],
    ([visible]) => {
        if (visible) {
            activeResponse.value = 1;
            loadVoters(true);
        } else {
            votesRequestID++;
            voters.value = [];
            votesOffset.value = 0;
            votesHasMore.value = true;
            loadingVotes.value = false;
            votesError.value = '';
        }
    }
);

watch(votesContainer, (newEl, oldEl) => {
    if (oldEl) oldEl.removeEventListener('scroll', handleVotesScroll);
    if (newEl) newEl.addEventListener('scroll', handleVotesScroll);
});

onUnmounted(() => {
    if (votesContainer.value) {
        votesContainer.value.removeEventListener('scroll', handleVotesScroll);
    }
});
</script>

<template>
    <Teleport to="body">
        <div v-if="show" class="votes-overlay" @click.self="closeVotesDialog">
            <section class="votes-dialog">
                <header class="votes-header">
                    <div class="votes-heading">
                        <span>VOTES</span>
                        <strong>{{ eventTitle }}</strong>
                    </div>

                    <button type="button" class="votes-close" @click="closeVotesDialog">×</button>
                </header>

                <div class="votes-tabs">
                    <button type="button" :class="{ active: activeResponse === 1 }" @click="selectResponse(1)">
                        Going
                    </button>

                    <button type="button" :class="{ active: activeResponse === 0 }" @click="selectResponse(0)">
                        Not going
                    </button>
                </div>

                <div ref="votesContainer" class="votes-list">
                    <div v-if="votesError" class="votes-state error">{{ votesError }}</div>

                    <div v-else-if="!voters.length && !loadingVotes" class="votes-state">
                        No votes yet.
                    </div>

                    <div v-for="voter in voters" :key="`${voter.user?.ID}-${voter.response}`" class="voter-row">
                        <img v-if="voter.user?.avatar" :src="`/uploads/${voter.user.avatar}`" class="voter-avatar">
                        <div v-else class="voter-avatar fallback">
                            {{ (voter.user?.firstName || '?').charAt(0) }}
                        </div>

                        <div class="voter-info">
                            <strong>{{ voterName(voter) }}</strong>
                            <span>{{ voter.response === 1 ? 'Going' : 'Not going' }}</span>
                        </div>
                    </div>

                    <div v-if="loadingVotes" class="votes-state">Loading...</div>
                </div>
            </section>
        </div>
    </Teleport>
</template>

<style scoped>
.votes-overlay {
    position: fixed;
    inset: 0;
    z-index: 1001;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(0, 0, 0, 0.6);
}

.votes-dialog {
    display: flex;
    width: 100%;
    max-width: 460px;
    height: min(560px, 88vh);
    height: min(560px, 88dvh);
    flex-direction: column;
    overflow: hidden;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.votes-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 18px;
    border-bottom: 2px solid var(--main-color);
}

.votes-heading {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 3px;
}

.votes-heading span {
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.votes-heading strong {
    overflow: hidden;
    color: var(--font-color);
    font-size: 14px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.votes-close {
    width: 30px;
    height: 30px;
    flex-shrink: 0;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    font-size: 16px;
    line-height: 1;
}

.votes-tabs {
    display: flex;
    border-bottom: 2px solid var(--main-color);
}

.votes-tabs button {
    flex: 1;
    padding: 11px 10px;
    border: 0;
    border-right: 2px solid var(--main-color);
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
}

.votes-tabs button:last-child {
    border-right: 0;
}

.votes-tabs button.active {
    background: var(--main-color);
    color: var(--bg-color);
}

.votes-list {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 12px 16px;
}

.voter-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 0;
    border-bottom: 1px solid var(--page-background);
}

.voter-row:last-child {
    border-bottom: 0;
}

.voter-avatar {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    object-fit: cover;
    background: var(--page-background);
}

.voter-avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--main-color);
    color: var(--bg-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 700;
    text-transform: uppercase;
}

.voter-info {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 2px;
}

.voter-info strong {
    color: var(--font-color);
    font-size: 13px;
}

.voter-info span {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.votes-state {
    padding: 24px 0;
    color: var(--font-color-sub);
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.votes-state.error {
    color: #e5484d;
}

@media (max-width: 520px) {
    .votes-overlay {
        padding: 0;
    }

    .votes-dialog {
        max-width: none;
        height: 100%;
        border-radius: 0;
        box-shadow: none;
    }
}
</style>
