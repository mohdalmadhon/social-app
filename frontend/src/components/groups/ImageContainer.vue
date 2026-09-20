<script setup>
import { ref, watch } from 'vue';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    modelValue: {
        type: File,
        default: null
    }
});

const emit = defineEmits(['update:modelValue']);

const fileInput = ref(null);
const preview = ref('');

const allowedTypes = [
    'image/png',
    'image/jpeg',
    'image/gif'
];

const allowedExtensions = ['png', 'jpg', 'jpeg', 'gif'];

function openFilePicker() {
    fileInput.value?.click();
}

function handleFileChange(event) {
    const file = event.target.files?.[0];

    event.target.value = '';

    if (!file) {
        return;
    }

    try {
        const extension = (file.name.split('.').pop() || '').toLowerCase();

        if (
            !allowedTypes.includes(file.type) ||
            !allowedExtensions.includes(extension)
        ) {
            addNotification(
                'Please upload a PNG, JPG or GIF image',
                'error'
            );
            emit('update:modelValue', null);
            return;
        }

        if (preview.value) {
            URL.revokeObjectURL(preview.value);
        }

        preview.value = URL.createObjectURL(file);

        emit('update:modelValue', file);
    } catch (err) {
        addNotification('Could not read the selected image', 'error');
        emit('update:modelValue', null);
    }
}

function removeImage() {
    if (preview.value) {
        URL.revokeObjectURL(preview.value);
    }

    preview.value = '';
    emit('update:modelValue', null);

    if (fileInput.value) {
        fileInput.value.value = '';
    }
}

watch(
    () => props.modelValue,
    file => {
        if (!file) {
            if (preview.value) {
                URL.revokeObjectURL(preview.value);
            }

            preview.value = '';
            return;
        }

        if (!preview.value) {
            preview.value = URL.createObjectURL(file);
        }
    }
);
</script>

<template>
    <div class="add-picture">
        <input ref="fileInput" type="file" accept=".png,.jpg,.jpeg,.gif" hidden @change="handleFileChange">

        <div v-if="!preview" class="picture-upload" @click="openFilePicker">
            <div class="upload-icon">+</div>
            <strong>Add picture</strong>
            <span>Click to choose an image</span>
        </div>

        <div v-else class="picture-preview">
            <img :src="preview" alt="Selected picture">

            <div class="picture-actions">
                <button type="button" @click="openFilePicker">
                    Change
                </button>

                <button type="button" @click="removeImage">
                    Remove
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.add-picture {
    width: 100%;
}

.picture-upload {
    min-height: 180px;
    border: 2px dashed var(--main-color);
    border-radius: 10px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    cursor: pointer;
    transition: 0.2s;
}

.picture-upload:hover {
    background: var(--input-focus);
}

.upload-icon {
    width: 45px;
    height: 45px;
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    font-weight: bold;
}

.picture-upload span {
    font-size: 12px;
    opacity: 0.7;
}

.picture-preview {
    position: relative;
    width: 100%;
}

.picture-preview img {
    width: 100%;
    max-height: 220px;
    object-fit: contain;
    border-radius: 10px;
    display: block;
}

.picture-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
}

.picture-actions button {
    border: none;
    padding: 8px 14px;
    border-radius: 6px;
    cursor: pointer;
    font-weight: bold;
}
</style>