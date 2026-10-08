<script setup>
import { LIMITS } from '@/helpers/limits';
defineProps({
    title: {
        type: String,
        default: ''
    },
    description: {
        type: String,
        default: ''
    }
});

defineEmits(['update:title', 'update:description']);
</script>

<template>
    <div class="post-content-form">

        <div class="field-group">
            <label for="post-description">Content</label>

            <textarea
                id="post-description"
                :maxlength="LIMITS.postContent"
                placeholder="post content"
                required
                :value="description"
                @input="$emit('update:description', $event.target.value)"
            ></textarea>
        </div>
    </div>
</template>

<style scoped>
.post-content-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
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
    transition: box-shadow 0.15s ease;
}

textarea {
    min-height: 130px;
    resize: vertical;
    line-height: 1.5;
}

input:focus,
textarea:focus {
    box-shadow: 3px 3px var(--main-color);
}
</style>
