<script setup>
import { ref, computed } from 'vue';

const activeTab = ref('information');

const props = defineProps({
    about: {
        type: Object,
        required: true
    },
    email: {
        type: String,
        default: ''
    },
    dob: {
        type: String,
        default: ''
    },
    visibility: {
        type: Object,
        default: null
    }
});

const showEmail = computed(() => !!props.email && (!props.visibility || props.visibility.email));

const showDob = computed(() => {
    if (!props.dob || props.dob.startsWith('0001')) {
        return false;
    }
    return !props.visibility || props.visibility.dob;
});

const formattedDob = computed(() => {
    const date = new Date(props.dob);
    if (isNaN(date.getTime())) {
        return props.dob;
    }
    return date.toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        timeZone: 'UTC'
    });
});

const additionalHidden = computed(() => !!props.visibility && !props.visibility.additionalInfo);

const additionalAudience = computed(() => {
    if (!props.visibility) {
        return '';
    }
    return props.visibility.additionalInfoAudience === 'friends' ? 'friends' : 'followers';
});

const tabs = [
    {
        id: 'information',
        label: 'Information'
    },
    {
        id: 'links',
        label: 'Social Links'
    }
];
</script>

<template>
    <section class="about-section">
        <div class="section-heading">
            <p class="eyebrow">PROFILE</p>
            <h2>About</h2>
        </div>

        <div class="about-card">
            <nav class="about-navigation">
                <button
                    v-for="tab in tabs"
                    :key="tab.id"
                    type="button"
                    :class="{ active: activeTab === tab.id }"
                    @click="activeTab = tab.id"
                >
                    {{ tab.label }}
                </button>
            </nav>

            <main class="about-content">
                <div
                    v-if="activeTab === 'information'"
                    class="content-section"
                >
                    <div class="content-heading">
                        <p class="eyebrow">PERSONAL</p>
                        <h3>Information</h3>
                    </div>

                    <div class="info-list">
                        <div
                            v-if="showEmail"
                            class="info-item"
                        >
                            <span>Email</span>
                            <p>{{ props.email }}</p>
                        </div>

                        <div
                            v-if="showDob"
                            class="info-item"
                        >
                            <span>Date of birth</span>
                            <p>{{ formattedDob }}</p>
                        </div>

                        <p
                            v-if="additionalHidden"
                            class="empty"
                        >
                            Additional info is only visible to {{ additionalAudience }}.
                        </p>

                        <div
                            v-if="!additionalHidden && props.about.work"
                            class="info-item"
                        >
                            <span>Work</span>
                            <p>{{ props.about.work }}</p>
                        </div>

                        <div
                            v-if="!additionalHidden && props.about.education"
                            class="info-item"
                        >
                            <span>Education</span>
                            <p>{{ props.about.education }}</p>
                        </div>

                        <div
                            v-if="!additionalHidden && props.about.hobbies"
                            class="info-item"
                        >
                            <span>Hobbies</span>
                            <p>{{ props.about.hobbies }}</p>
                        </div>

                        <div
                            v-if="!additionalHidden && props.about.intrests"
                            class="info-item"
                        >
                            <span>Interests</span>
                            <p>{{ props.about.intrests }}</p>
                        </div>

                        <div
                            v-if="!additionalHidden && props.about.travel"
                            class="info-item"
                        >
                            <span>Travel</span>
                            <p>{{ props.about.travel }}</p>
                        </div>

                        <p
                            v-if="
                                !additionalHidden &&
                                !showEmail &&
                                !showDob &&
                                !props.about.work &&
                                !props.about.education &&
                                !props.about.hobbies &&
                                !props.about.intrests &&
                                !props.about.travel
                            "
                            class="empty"
                        >
                            No information provided.
                        </p>
                    </div>
                </div>

                <div
                    v-if="activeTab === 'links'"
                    class="content-section"
                >
                    <div class="content-heading">
                        <p class="eyebrow">SOCIAL</p>
                        <h3>Links</h3>
                    </div>

                    <div class="links-list">
                        <p
                            v-if="additionalHidden"
                            class="empty"
                        >
                            Social links are only visible to {{ additionalAudience }}.
                        </p>

                        <a
                            v-if="props.about.website"
                            :href="props.about.website"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="link-item"
                        >
                            <span>Website</span>
                            <p>{{ props.about.website }}</p>
                        </a>

                        <a
                            v-if="props.about.linkedin"
                            :href="props.about.linkedin"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="link-item"
                        >
                            <span>LinkedIn</span>
                            <p>{{ props.about.linkedin }}</p>
                        </a>

                        <a
                            v-if="props.about.twitter"
                            :href="props.about.twitter"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="link-item"
                        >
                            <span>Twitter / X</span>
                            <p>{{ props.about.twitter }}</p>
                        </a>

                        <a
                            v-if="props.about.instgram"
                            :href="props.about.instgram"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="link-item"
                        >
                            <span>Instagram</span>
                            <p>{{ props.about.instgram }}</p>
                        </a>

                        <p
                            v-if="
                                !additionalHidden &&
                                !props.about.website &&
                                !props.about.linkedin &&
                                !props.about.twitter &&
                                !props.about.instgram
                            "
                            class="empty"
                        >
                            No social links provided.
                        </p>
                    </div>
                </div>
            </main>
        </div>
    </section>
</template>

<style scoped>
.about-section {
    width: 100%;
    scroll-margin-top: 100px;
}

.section-heading {
    margin-bottom: clamp(14px, 2.5vw, 20px);
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
    letter-spacing: 2px;
}

h2 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: clamp(20px, 4vw, 29px);
}

.about-card {
    display: grid;
    grid-template-columns: minmax(150px, 200px) 1fr;
    min-height: 360px;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);
    overflow: hidden;
}

.about-navigation {
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding: clamp(12px, 2vw, 18px);
    border-right: 2px solid var(--main-color);
}

.about-navigation button {
    width: 100%;
    padding: 12px 10px;
    border: 2px solid transparent;
    border-radius: 5px;
    background: transparent;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(9px, 1.4vw, 11px);
    font-weight: 600;
    text-align: left;
    cursor: pointer;
    transition: 0.15s ease;
}

.about-navigation button:hover {
    border-color: var(--main-color);
    color: var(--font-color);
}

.about-navigation button.active {
    border-color: var(--main-color);
    background: var(--input-focus);
    color: white;
    box-shadow: 3px 3px var(--main-color);
}

.about-content {
    min-width: 0;
    padding: clamp(18px, 3vw, 28px);
}

.content-section {
    width: 100%;
}

.content-heading {
    margin-bottom: clamp(18px, 3vw, 25px);
}

.info-list {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.info-item {
    display: grid;
    grid-template-columns: minmax(90px, 130px) 1fr;
    gap: 20px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--main-color);
}

.info-item span,
.link-item span {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 10px);
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.info-item p {
    margin: 0;
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: clamp(12px, 1.8vw, 14px);
    line-height: 1.5;
    word-break: break-word;
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

@media (max-width: 600px) {
    .about-card {
        grid-template-columns: 1fr;
    }

    .about-navigation {
        flex-direction: row;
        border-right: none;
        border-bottom: 2px solid var(--main-color);
        overflow-x: auto;
    }

    .about-navigation button {
        width: auto;
        min-width: max-content;
        text-align: center;
    }
}

@media (max-width: 500px) {
    .info-item,
    .link-item {
        grid-template-columns: 1fr;
        gap: 6px;
    }
}
</style>

