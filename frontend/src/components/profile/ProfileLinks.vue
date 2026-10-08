<script setup>
import { safeHref } from '@/helpers/links';

defineProps({
    about: {
        type: Object,
        required: true
    }
});
</script>

<template>
    <div class="content-section">
        <div class="content-heading">
            <p class="eyebrow">SOCIAL</p>
            <h3>Links</h3>
        </div>

        <div class="links-list">
            <a
                v-if="about.Website"
                :href="safeHref('website', about.Website)"
                target="_blank"
                rel="noopener noreferrer"
                class="link-item"
            >
                <span>Website</span>
                <p class="user-text">{{ about.Website }}</p>
            </a>

            <a
                v-if="about.Linkedin"
                :href="safeHref('linkedin', about.Linkedin)"
                target="_blank"
                rel="noopener noreferrer"
                class="link-item"
            >
                <span>LinkedIn</span>
                <p class="user-text">{{ about.Linkedin }}</p>
            </a>

            <a
                v-if="about.Twitter"
                :href="safeHref('twitter', about.Twitter)"
                target="_blank"
                rel="noopener noreferrer"
                class="link-item"
            >
                <span>Twitter / X</span>
                <p class="user-text">{{ about.Twitter }}</p>
            </a>

            <a
                v-if="about.Instgram"
                :href="safeHref('instagram', about.Instgram)"
                target="_blank"
                rel="noopener noreferrer"
                class="link-item"
            >
                <span>Instagram</span>
                <p class="user-text">{{ about.Instgram }}</p>
            </a>

            <p
                v-if="!about.Website && !about.Linkedin && !about.Twitter && !about.Instgram"
                class="empty"
            >
                No social links provided.
            </p>
        </div>
    </div>
</template>

<style scoped>
.content-section {
    width: 100%;
}

.content-heading {
    margin-bottom: clamp(18px, 3vw, 25px);
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
    letter-spacing: 2px;
}

h3 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: clamp(20px, 3vw, 25px);
}

.links-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.link-item {
    display: grid;
    grid-template-columns: minmax(90px, 130px) 1fr;
    gap: 20px;
    padding: 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    text-decoration: none;
    transition: 0.15s ease;
}

.link-item:hover {
    transform: translate(-2px, -2px);
    box-shadow: 3px 3px var(--main-color);
}

.link-item span {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 10px);
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.link-item p {
    margin: 0;
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: clamp(11px, 1.7vw, 13px);
    word-break: break-all;
}

.empty {
    margin: 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

@media (max-width: 500px) {
    .link-item {
        grid-template-columns: 1fr;
        gap: 6px;
    }
}
</style>
