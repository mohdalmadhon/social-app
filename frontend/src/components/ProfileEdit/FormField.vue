<script setup>
defineProps({
    label: String,
    id: String,
    modelValue: String,
    type: {
        type: String,
        default: 'text'
    },
    placeholder: String,
    maxlength: {
        type: Number,
        default: 75
    },
    error: {
        type: String,
        default: ''
    }
})

defineEmits(['update:modelValue'])
</script>

<template>
    <div class="form-field">
        <label :for="id">{{ label }}</label>

        <textarea
            v-if="type === 'textarea'"
            :id="id"
            :placeholder="placeholder"
            :maxlength="maxlength"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
        ></textarea>

        <input
            v-else
            :id="id"
            :type="type"
            :placeholder="placeholder"
            :maxlength="maxlength"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
        >

        <p v-if="error" class="field-error">{{ error }}</p>
    </div>
</template>

<style scoped>
.field-error {
    margin: 0;
    color: #d93025;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    overflow-wrap: anywhere;
}

.form-field {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
}

label {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
    font-weight: 600;
    letter-spacing: 1px;
}

input,
textarea {
    width: 100%;
    padding: clamp(10px, 1.8vw, 12px) clamp(10px, 2vw, 14px);
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: clamp(12px, 1.8vw, 13px);
    box-sizing: border-box;
}

textarea {
    min-height: 110px;
    resize: vertical;
}

input:focus,
textarea:focus {
    outline: none;
    border-color: var(--input-focus);
    box-shadow: 3px 3px var(--input-focus);
}

@media (max-width: 550px) {
    input:focus,
    textarea:focus {
        box-shadow: 2px 2px var(--input-focus);
    }
}
</style>