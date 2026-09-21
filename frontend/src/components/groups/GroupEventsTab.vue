<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue';
import GroupEventVotesDialog from './GroupEventVotesDialog.vue';
import { fetchGroupEvents, respondGroupEvent } from '@/api/groups/events';
import { addNotification } from '@/data/notifications';
import { throttle } from '@/helpers/throttle';

const props = defineProps({
    groupId: {
        type: [Number, String],
        required: true
    }
});

const eventsLimit = 10;

const events = ref([]);
const eventsOffset = ref(0);
const eventsHasMore = ref(true);
const loadingEvents = ref(false);
const eventsError = ref('');
const respondingEvent = ref(null);
const selectedEvent = ref(null);
const showVotesDialog = ref(false);

let eventsRequestID = 0;

function formatEventTime(value) {
    if (!value) return '';

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) return value;

    return date.toLocaleString();
}

async function loadGroupEvents(reset = false) {
    if (loadingEvents.value) return;
    if (!reset && !eventsHasMore.value) return;

    const requestID = ++eventsRequestID;

    loadingEvents.value = true;
    eventsError.value = '';

    if (reset) {
        eventsOffset.value = 0;
        eventsHasMore.value = true;
        events.value = [];
    }

    try {
        const result = await fetchGroupEvents(
            props.groupId,
            eventsOffset.value,
            eventsLimit
        );

        if (requestID !== eventsRequestID) return;

        events.value = [...events.value, ...result.events];
        eventsOffset.value += result.events.length;

        if (result.events.length < eventsLimit) {
            eventsHasMore.value = false;
        }
    } catch (err) {
        if (requestID !== eventsRequestID) return;

        eventsError.value = err.message || 'Could not load events';
        eventsHasMore.value = false;
    } finally {
        if (requestID === eventsRequestID) {
            loadingEvents.value = false;
        }
    }
}

const handleEventsScroll = throttle(() => {
    const scrolled = window.innerHeight + window.scrollY;
    const pageHeight = document.documentElement.scrollHeight;

    if (scrolled >= pageHeight - 300) {
        loadGroupEvents();
    }
}, 300);

async function answerEvent(event, value) {
    if (respondingEvent.value === event.id) return;

    respondingEvent.value = event.id;

    try {
        const updated = await respondGroupEvent(event.id, value);

        const index = events.value.findIndex(item => item.id === event.id);

        if (index !== -1) {
            events.value[index] = updated;
        }
    } catch (err) {
        addNotification(err.message || 'Could not save response', 'error');
    } finally {
        respondingEvent.value = null;
    }
}

function openVotes(event) {
    selectedEvent.value = event;
    showVotesDialog.value = true;
}

function closeVotes() {
    showVotesDialog.value = false;
    selectedEvent.value = null;
}

function addCreatedEvent(event) {
    if (!event) return;

    events.value = [event, ...events.value.filter(item => item.id !== event.id)];
    eventsOffset.value += 1;
}

function reloadEvents() {
    loadGroupEvents(true);
}

defineExpose({
    addCreatedEvent,
    reloadEvents
});

watch(
    () => props.groupId,
    () => loadGroupEvents(true)
);

onMounted(() => {
    loadGroupEvents(true);
    window.addEventListener('scroll', handleEventsScroll);
});

onUnmounted(() => {
    eventsRequestID++;
    window.removeEventListener('scroll', handleEventsScroll);
});
</script>

<template>
    <section class="tab-panel">
        <div v-if="eventsError" class="empty-state error">{{ eventsError }}</div>

        <div v-else-if="!events.length && !loadingEvents" class="empty-state">
            No events scheduled yet.
        </div>

        <div v-else class="events-list">
            <article v-for="event in events" :key="event.id" class="event-card">
                <header class="event-card-head">
                    <div class="event-date">{{ formatEventTime(event.eventTime) }}</div>

                    <div class="event-creator">
                        <img v-if="event.creator?.avatar" :src="`/uploads/${event.creator.avatar}`"
                            class="creator-avatar">
                        <div v-else class="creator-avatar fallback">
                            {{ (event.creator?.firstName || '?').charAt(0) }}
                        </div>

                        <span>
                            {{ event.creator?.firstName }} {{ event.creator?.lastName }}
                        </span>
                    </div>
                </header>

                <div class="event-info">
                    <h3 class="event-title">{{ event.title }}</h3>

                    <p v-if="event.description" class="event-description">
                        {{ event.description }}
                    </p>
                </div>

                <div class="event-counts">
                    <span>{{ event.goingCount }} going</span>
                    <span>{{ event.notGoingCount }} not going</span>
                </div>

                <footer class="event-actions">
                    <button type="button" class="event-button" :class="{ active: event.userResponse === 1 }"
                        :disabled="respondingEvent === event.id" @click="answerEvent(event, 1)">
                        Going
                    </button>

                    <button type="button" class="event-button" :class="{ active: event.userResponse === 0 }"
                        :disabled="respondingEvent === event.id" @click="answerEvent(event, 0)">
                        Not going
                    </button>

                    <button type="button" class="event-button ghost" @click="openVotes(event)">
                        See votes
                    </button>
                </footer>
            </article>

            <div v-if="loadingEvents" class="loading-more">Loading events...</div>
        </div>

        <GroupEventVotesDialog :show="showVotesDialog" :event-id="selectedEvent?.id ?? null"
            :event-title="selectedEvent?.title ?? ''" @close="closeVotes" />
    </section>
</template>

<style scoped>
.tab-panel {
    width: 100%;
}

.empty-state {
    padding: 40px 20px;
    text-align: center;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
}

.empty-state.error {
    color: #e5484d;
}

.events-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.event-card {
    display: flex;
    flex-direction: column;
    gap: 12px;

    padding: 16px 18px;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);
}

.event-card-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
}

.event-date {
    padding: 8px 12px;

    border: 2px solid var(--main-color);
    border-radius: 6px;

    background: var(--input-focus);
    color: #fff;

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
}

.event-creator {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.creator-avatar {
    width: 28px;
    height: 28px;
    flex-shrink: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    object-fit: cover;
    background: var(--page-background);
}

.creator-avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--main-color);
    color: var(--bg-color);
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
}

.event-info {
    min-width: 0;
}

.event-title {
    margin: 0 0 6px;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 18px;
    font-weight: 700;
    overflow-wrap: anywhere;
}

.event-description {
    margin: 0;
    color: var(--font-color-sub);
    font-size: 13px;
    line-height: 1.5;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}

.event-counts {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.event-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
}

.event-button {
    flex: 1 1 auto;
    min-width: 100px;
    padding: 9px 14px;

    border: 2px solid var(--main-color);
    border-radius: 6px;

    background: var(--bg-color);
    color: var(--font-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;

    transition: background 0.15s, color 0.15s;
}

.event-button:hover:not(:disabled) {
    background: var(--main-color);
    color: var(--bg-color);
}

.event-button.active {
    background: var(--input-focus);
    border-color: var(--input-focus);
    color: #fff;
}

.event-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.loading-more {
    padding: 14px;
    color: var(--font-color-sub);
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

@media (max-width: 650px) {
    .event-card {
        padding: 14px;
        box-shadow: 4px 4px var(--main-color);
    }

    .event-title {
        font-size: 16px;
    }

    .event-button {
        min-width: 0;
        flex: 1 1 100%;
    }
}
</style>
