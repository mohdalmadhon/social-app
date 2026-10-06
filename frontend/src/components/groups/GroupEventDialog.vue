<script setup>
import { ref, watch } from 'vue';
import { createGroupEvent } from '@/api/groups/events';

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

const emit = defineEmits(['close', 'created']);

const title = ref('');
const description = ref('');
const eventTime = ref('');
const submitting = ref(false);
const minTime = ref('');
const formError = ref('');

function currentLocalTime() {
    const now = new Date();
    now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
    return now.toISOString().slice(0, 16);
}

function resetEventForm() {
    minTime.value = currentLocalTime();
    title.value = '';
    description.value = '';
    eventTime.value = '';
    formError.value = '';
    submitting.value = false;
}

function closeEventDialog() {
    emit('close');
}

async function submitEvent() {
    if (submitting.value) return;

    const eventTitle = title.value.trim();
    const eventDescription = description.value.trim();
    const eventDate = eventTime.value;

    if (!eventTitle) {
        formError.value = 'Title is required';
        return;
    }

    if (!eventDate) {
        formError.value = 'Day and time are required';
        return;
    }

    if (new Date(eventDate).getTime() <= Date.now()) {
        formError.value = 'Event cannot be in the past';
        return;
    }

    submitting.value = true;
    formError.value = '';

    try {
        const created = await createGroupEvent({
            groupId: Number(props.groupId),
            title: eventTitle,
            description: eventDescription,
            eventTime: eventDate
        });

        emit('created', created);
        resetEventForm();
    } catch (err) {
        formError.value = err.message || 'Could not create event';
    } finally {
        submitting.value = false;
    }
}

watch(
    () => props.show,
    value => {
        if (value) resetEventForm();
    }
);
</script>

<template>
    <Teleport to="body">
        <div v-if="show" class="event-dialog-overlay" @click.self="closeEventDialog">
            <section class="event-dialog">
                <header class="event-dialog-header">
                    <h2>New event</h2>
                    <button type="button" class="event-dialog-close" @click="closeEventDialog">×</button>
                </header>

                <form class="event-dialog-form" @submit.prevent="submitEvent">
                    <label class="event-field">
                        <span>Title</span>
                        <input v-model="title" type="text" maxlength="100" placeholder="Event title">
                    </label>

                    <label class="event-field">
                        <span>Description</span>
                        <textarea v-model="description" maxlength="1000" rows="4"
                            placeholder="What is this event about?"></textarea>
                    </label>

                    <label class="event-field">
                        <span>Day / Time</span>
                        <input v-model="eventTime" type="datetime-local" :min="minTime">
                    </label>

                    <p v-if="formError" class="event-form-error">{{ formError }}</p>

                    <div class="event-form-actions">
                        <button type="button" class="ghost-button" @click="closeEventDialog">Cancel</button>
                        <button type="submit" class="primary-button" :disabled="submitting">
                            {{ submitting ? 'Creating...' : 'Create event' }}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    </Teleport>
</template>

<style scoped>
.event-dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    overflow-y: auto;
    background: rgba(0, 0, 0, 0.6);
}

.event-dialog {
    width: 100%;
    max-width: 520px;
    max-height: calc(100vh - 48px);
    max-height: calc(100dvh - 48px);
    overflow-y: auto;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.event-dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 2px solid var(--main-color);
}

.event-dialog-header h2 {
    margin: 0;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
}

.event-dialog-close {
    width: 30px;
    height: 30px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    font-size: 16px;
    line-height: 1;
}

.event-dialog-form {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 20px;
}

.event-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.event-field span {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
}

.event-field input,
.event-field textarea {
    width: 100%;
    padding: 10px 12px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    outline: none;
    background: var(--page-background);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    resize: vertical;
}

.event-field input:focus,
.event-field textarea:focus {
    background: var(--bg-color);
}

.event-form-error {
    margin: 0;
    color: #e5484d;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.event-form-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 10px;
}

.ghost-button,
.primary-button {
    flex: 0 1 auto;
    padding: 10px 18px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
}

.ghost-button {
    background: var(--bg-color);
    color: var(--font-color);
}

.primary-button {
    background: var(--input-focus);
    color: #fff;
    box-shadow: 3px 3px var(--main-color);
}

.primary-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

@media (max-width: 520px) {
    .event-dialog-overlay {
        padding: 12px;
    }

    .event-form-actions button {
        flex: 1 1 100%;
    }
}
</style>
