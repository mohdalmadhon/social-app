<script setup>
import { getSuggestedLocations } from '@/api/common/location';
import { getLocationData } from '@/helpers/common/locationHelpers';
import { onBeforeUnmount, ref } from 'vue';

defineProps({
    modelValue: {
        type: String,
        default: ''
    }
});

const emit = defineEmits(['update:modelValue']);

const MIN_SEARCH_LENGTH = 2;
const SEARCH_DEBOUNCE = 500;

const search = ref('');
const locations = ref([]);
const selectedLocation = ref(null);
const loading = ref(false);
const searched = ref(false);
const errorMessage = ref('');
const open = ref(false);
const activeIndex = ref(-1);

let searchTimeout = null;
let requestNumber = 0;

function resetResults() {
    locations.value = [];
    searched.value = false;
    errorMessage.value = '';
    activeIndex.value = -1;
    loading.value = false;
}

function handleLocationSearch() {
    clearTimeout(searchTimeout);

    selectedLocation.value = null;

    const current = ++requestNumber;
    const value = search.value.trim();

    if (value.length < MIN_SEARCH_LENGTH) {
        resetResults();
        open.value = false;
        return;
    }

    loading.value = true;
    searched.value = false;
    errorMessage.value = '';
    open.value = true;

    searchTimeout = setTimeout(async () => {
        try {
            const result = await getSuggestedLocations(value);

            if (current !== requestNumber) {
                return;
            }

            locations.value = getLocationData(
                Array.isArray(result) ? result : []
            );

            errorMessage.value = '';
        } catch (error) {
            if (current !== requestNumber) {
                return;
            }

            console.error(error);

            locations.value = [];
            errorMessage.value = 'Could not load locations';
        } finally {
            if (current === requestNumber) {
                loading.value = false;
                searched.value = true;
                activeIndex.value = locations.value.length ? 0 : -1;
            }
        }
    }, SEARCH_DEBOUNCE);
}

function selectLocation(location) {
    clearTimeout(searchTimeout);

    requestNumber++;

    selectedLocation.value = location;

    search.value = location.label;

    resetResults();

    open.value = false;
}

function addLocation() {
    if (!selectedLocation.value) {
        return;
    }

    const { label, lat, lon } = selectedLocation.value;

    emit('update:modelValue', `${label}:${lat}:${lon}`);

    search.value = '';
    selectedLocation.value = null;

    resetResults();

    open.value = false;
}

function removeLocation() {
    emit('update:modelValue', '');
}

function openSuggestions() {
    if (locations.value.length || errorMessage.value || searched.value) {
        open.value = true;
    }
}

function closeSuggestions() {
    open.value = false;
}

function handleKeydown(event) {
    if (event.key === 'ArrowDown' && open.value && locations.value.length) {
        event.preventDefault();

        activeIndex.value = (activeIndex.value + 1) % locations.value.length;

        return;
    }

    if (event.key === 'ArrowUp' && open.value && locations.value.length) {
        event.preventDefault();

        activeIndex.value =
            (activeIndex.value - 1 + locations.value.length) %
            locations.value.length;

        return;
    }

    if (event.key === 'Enter') {
        const highlighted = locations.value[activeIndex.value];

        if (open.value && highlighted) {
            event.preventDefault();
            selectLocation(highlighted);
            return;
        }

        if (selectedLocation.value) {
            event.preventDefault();
            addLocation();
        }

        return;
    }

    if (event.key === 'Escape') {
        closeSuggestions();
    }
}

onBeforeUnmount(() => {
    clearTimeout(searchTimeout);

    requestNumber++;
});
</script>

<template>

    <div class="location-picker">

        <label for="post-location">
            Add location
        </label>

        <div class="location-row">

            <div class="location-input">

                <input
                    id="post-location"
                    type="text"
                    placeholder="Search a city, state, country or place"
                    autocomplete="off"
                    v-model="search"
                    @input="handleLocationSearch"
                    @keydown="handleKeydown"
                    @focus="openSuggestions"
                    @blur="closeSuggestions"
                >

                <div
                    v-if="open && (loading || locations.length || errorMessage || searched)"
                    class="location-suggestions"
                >

                    <div
                        v-if="loading"
                        class="location-status"
                    >
                        Searching...
                    </div>

                    <div
                        v-else-if="errorMessage"
                        class="location-status location-status-error"
                    >
                        {{ errorMessage }}
                    </div>

                    <div
                        v-else-if="!locations.length"
                        class="location-status"
                    >
                        No locations found. Try a city, state, country or place name.
                    </div>

                    <button
                        v-for="(location, index) in locations"
                        v-show="!loading && !errorMessage"
                        :key="location.key"
                        type="button"
                        class="location-suggestion"
                        :class="{ active: index === activeIndex }"
                        @mousedown.prevent
                        @mouseenter="activeIndex = index"
                        @click="selectLocation(location)"
                    >
                        <span class="location-kind">
                            {{ location.kindLabel }}
                        </span>

                        <span class="location-text">
                            <strong>{{ location.title }}</strong>

                            <small v-if="location.subtitle">
                                {{ location.subtitle }}
                            </small>
                        </span>
                    </button>

                </div>

            </div>

            <button
                type="button"
                :disabled="!selectedLocation"
                @click="addLocation"
            >
                Add
            </button>

        </div>

        <div
            v-if="modelValue"
            class="location-chip"
        >
            <span>{{ modelValue.split(':')[0].trim() }}</span>

            <button
                type="button"
                @click="removeLocation"
            >
                ×
            </button>
        </div>

    </div>

</template>

<style scoped>

.location-picker {
    display: flex;
    flex-direction: column;
    gap: 7px;
    min-width: 0;
}

label {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.location-row {
    display: flex;
    gap: 10px;
    min-width: 0;
}

.location-input {
    position: relative;
    flex: 1;
    min-width: 0;
}

.location-row input {
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

.location-row input:focus {
    box-shadow: 3px 3px var(--main-color);
}

.location-row > button {
    flex-shrink: 0;
    padding: 0 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 3px 3px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
    transition: transform 0.1s ease, box-shadow 0.1s ease, opacity 0.15s ease;
}

.location-row > button:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.location-row > button:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.location-row > button:disabled {
    opacity: 0.55;
    cursor: default;
}

.location-suggestions {
    position: absolute;
    top: calc(100% + 5px);
    left: 0;
    right: 0;
    z-index: 10;
    max-height: 300px;
    overflow-x: hidden;
    overflow-y: auto;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.location-suggestion {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px 14px;
    border: 0;
    border-bottom: 1px solid var(--main-color);
    background: var(--bg-color);
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
}

.location-suggestion:last-child {
    border-bottom: 0;
}

.location-suggestion:hover,
.location-suggestion.active {
    background: var(--input-focus);
    color: white;
}

.location-kind {
    flex-shrink: 0;
    min-width: 54px;
    max-width: 90px;
    padding: 3px 7px;
    border: 1px solid currentColor;
    border-radius: 10px;
    overflow: hidden;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
}

.location-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    min-width: 0;
}

.location-text strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    font-weight: 600;
}

.location-text small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    opacity: 0.75;
    font-size: 11px;
}

.location-status {
    padding: 12px 14px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.location-status-error {
    color: #d93025;
}

.location-chip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    width: fit-content;
    max-width: 100%;
    margin-top: 4px;
    padding: 7px 10px 7px 14px;
    border: 2px solid var(--main-color);
    border-radius: 20px;
    background: var(--page-background);
    color: var(--font-color);
    font-size: 12px;
    font-weight: 600;
}

.location-chip span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.location-chip button {
    flex-shrink: 0;
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