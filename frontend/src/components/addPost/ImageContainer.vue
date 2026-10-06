```vue
<script setup>
import { ref, watch, onBeforeUnmount } from 'vue';

const props = defineProps({
    modelValue: {
        type: File,
        default: null
    }
});

const emit = defineEmits(['update:modelValue']);

const fileInput = ref(null);
const preview = ref('');
const videoUrl = ref('');
const isVideo = ref(false);
const isPlaying = ref(false);

function openFilePicker() {
    fileInput.value?.click();
}

function createVideoThumbnail(file) {
    return new Promise((resolve, reject) => {
        const video = document.createElement('video');
        const videoObjectUrl = URL.createObjectURL(file);

        video.src = videoObjectUrl;
        video.muted = true;
        video.playsInline = true;
        video.preload = 'metadata';

        video.addEventListener('loadeddata', () => {
            video.currentTime = 0.1;
        });

        video.addEventListener('seeked', () => {
            const canvas = document.createElement('canvas');

            canvas.width = video.videoWidth;
            canvas.height = video.videoHeight;

            const context = canvas.getContext('2d');

            if (!context) {
                URL.revokeObjectURL(videoObjectUrl);
                reject(new Error('Could not create canvas context'));
                return;
            }

            context.drawImage(
                video,
                0,
                0,
                canvas.width,
                canvas.height
            );

            canvas.toBlob(
                blob => {
                    URL.revokeObjectURL(videoObjectUrl);

                    if (!blob) {
                        reject(new Error('Could not create thumbnail'));
                        return;
                    }

                    resolve(URL.createObjectURL(blob));
                },
                'image/jpeg',
                0.85
            );
        });

        video.addEventListener('error', () => {
            URL.revokeObjectURL(videoObjectUrl);
            reject(new Error('Could not load video'));
        });
    });
}

async function handleFileChange(event) {
    const file = event.target.files?.[0];

    if (!file) {
        return;
    }

    const allowedTypes = [
        'image/png',
        'image/jpeg',
        'image/gif',
        'video/mp4'
    ];

    const allowedExtensions = [
        'png',
        'jpg',
        'jpeg',
        'gif',
        'mp4'
    ];

    const extension = file.name
        .split('.')
        .pop()
        .toLowerCase();

    if (
        !allowedTypes.includes(file.type) ||
        !allowedExtensions.includes(extension)
    ) {
        event.target.value = '';
        emit('update:modelValue', null);
        return;
    }

    cleanupPreview();

    if (file.type === 'video/mp4') {
        isVideo.value = true;
        videoUrl.value = URL.createObjectURL(file);

        try {
            preview.value = await createVideoThumbnail(file);
        } catch (error) {
            console.error('Video thumbnail error:', error);

            cleanupPreview();
            event.target.value = '';
            emit('update:modelValue', null);
            return;
        }
    } else {
        isVideo.value = false;
        preview.value = URL.createObjectURL(file);
    }

    emit('update:modelValue', file);
}

function playVideo() {
    if (!isVideo.value) {
        return;
    }

    isPlaying.value = true;
}

function stopVideo() {
    isPlaying.value = false;
}

function cleanupPreview() {
    if (preview.value) {
        URL.revokeObjectURL(preview.value);
    }

    if (videoUrl.value) {
        URL.revokeObjectURL(videoUrl.value);
    }

    preview.value = '';
    videoUrl.value = '';
    isVideo.value = false;
    isPlaying.value = false;
}

function removeImage() {
    cleanupPreview();

    emit('update:modelValue', null);

    if (fileInput.value) {
        fileInput.value.value = '';
    }
}

watch(
    () => props.modelValue,
    file => {
        if (!file) {
            cleanupPreview();
        }
    }
);

onBeforeUnmount(() => {
    cleanupPreview();
});
</script>

<template>
    <div class="add-picture">
        <input
            ref="fileInput"
            type="file"
            accept=".png,.jpg,.jpeg,.gif,.mp4"
            hidden
            @change="handleFileChange"
        >

        <div
            v-if="!preview"
            class="picture-upload"
            @click="openFilePicker"
        >
            <div class="upload-icon">+</div>

            <strong>Add picture or video</strong>

            <span>
                Click to choose an image or MP4 video
            </span>
        </div>

        <div v-else class="picture-preview">

            <div
                v-if="isVideo && isPlaying"
                class="video-container"
            >
                <video
                    :src="videoUrl"
                    controls
                    autoplay
                    playsinline
                    class="video-player"
                ></video>

                <button
                    type="button"
                    class="close-video"
                    @click="stopVideo"
                >
                    Close
                </button>
            </div>

            <div
                v-else
                class="preview-container"
                @click="playVideo"
            >
                <img
                    :src="preview"
                    :alt="isVideo ? 'Selected video thumbnail' : 'Selected picture'"
                >

                <div
                    v-if="isVideo"
                    class="video-overlay"
                >
                    <div class="play-icon">
                        ▶
                    </div>

                    <span>Play video</span>
                </div>
            </div>

            <div class="picture-actions">
                <button
                    type="button"
                    @click="openFilePicker"
                >
                    Change
                </button>

                <button
                    type="button"
                    @click="removeImage"
                >
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

.preview-container {
    position: relative;
    width: 100%;
    overflow: hidden;
    border-radius: 10px;
    cursor: pointer;
}

.preview-container img {
    width: 100%;
    max-height: 400px;
    object-fit: contain;
    border-radius: 10px;
    display: block;
}

.video-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: rgba(0, 0, 0, 0.25);
    color: white;
}

.play-icon {
    width: 65px;
    height: 65px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.75);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    padding-left: 4px;
}

.video-overlay span {
    font-size: 14px;
    font-weight: bold;
    background: rgba(0, 0, 0, 0.75);
    padding: 5px 10px;
    border-radius: 5px;
}

.video-container {
    position: relative;
    width: 100%;
}

.video-player {
    width: 100%;
    max-height: 500px;
    display: block;
    border-radius: 10px;
    background: black;
}

.close-video {
    margin-top: 8px;
    border: none;
    padding: 8px 14px;
    border-radius: 6px;
    cursor: pointer;
    font-weight: bold;
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
