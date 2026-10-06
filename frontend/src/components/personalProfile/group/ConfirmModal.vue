<script setup>
defineProps({
    title: {
        type: String,
        default: 'Are you sure?'
    },
    confirmLabel: {
        type: String,
        default: 'Confirm'
    },
    cancelLabel: {
        type: String,
        default: 'Cancel'
    },
    danger: {
        type: Boolean,
        default: true
    }
});

const emit = defineEmits(['confirm', 'cancel']);
</script>

<template>
    <div class="modal-overlay" @click.self="emit('cancel')">
        <div class="modal confirm-modal">
            <div class="modal-header">
                <h3>{{ title }}</h3>

                <button type="button" class="close-button" @click="emit('cancel')">
                    ×
                </button>
            </div>

            <div class="modal-body">
                <slot />
            </div>

            <div class="modal-footer">
                <button type="button" class="secondary-button" @click="emit('cancel')">
                    {{ cancelLabel }}
                </button>

                <button type="button" class="primary-button" :class="{ 'danger-button': danger }"
                    @click="emit('confirm')">
                    {{ confirmLabel }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
* {
    box-sizing: border-box;
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, .4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1100;
}

.modal {
    width: 100%;
    max-width: 640px;
    max-height: 90vh;
    max-height: 90dvh;
    display: flex;
    flex-direction: column;
    background: #fff;
    border: 3px solid #1a1a1a;
    border-radius: 12px;
    box-shadow: 6px 6px 0 #1a1a1a;
}

.confirm-modal {
    max-width: 340px;
}

.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 3px solid #1a1a1a;
}

.modal-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 800;
    color: #1a1a1a;
}

.close-button {
    border: none;
    background: none;
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    color: #1a1a1a;
}

.modal-body {
    padding: 24px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 16px 20px;
    border-top: 3px solid #1a1a1a;
}

.primary-button,
.secondary-button {
    padding: 10px 18px;
    border: 3px solid #1a1a1a;
    border-radius: 8px;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 800;
    cursor: pointer;
    box-shadow: 4px 4px 0 #1a1a1a;
    transition: transform .1s ease, box-shadow .1s ease;
}

.primary-button:hover,
.secondary-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px 0 #1a1a1a;
}

.primary-button:active,
.secondary-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0 #1a1a1a;
}

.primary-button {
    background: #2563eb;
    color: #fff;
}

.danger-button {
    background: #ef4444;
}

.secondary-button {
    background: #fff;
    color: #1a1a1a;
}
</style>
