<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue';
import { THEMES, getThemeCookie, setTheme } from '@/helpers/common/theme';

const TABS = [
    { value: 'about', label: 'About' },
    { value: 'authors', label: 'Authors' },
    { value: 'features', label: 'Features' }
];

const AUTHORS = [
    {
        name: 'Mohammed Almadhoon',
        gitea: 'malmadhoo',
        github: 'https://github.com/mohdalmadhon/',
        image: 'https://avatars.githubusercontent.com/u/39929536?v=4'
    },
    {
        name: 'Sulaiman Almubarak',
        gitea: 'salmubar',
        github: 'https://github.com/sulubaby',
        image: 'https://avatars.githubusercontent.com/u/222214954?v=4'
    },
    {
        name: 'Mohammed Rajabi',
        gitea: 'mrajabi',
        github: 'https://github.com/mohammedrajabi',
        image: 'https://avatars.githubusercontent.com/u/226603147?v=4'
    }
];

const FEATURES = [
    {
        icon: '<circle cx="12" cy="8" r="4"/><path d="M4 21v-1a6 6 0 0 1 6-6h4a6 6 0 0 1 6 6v1"/>',
        title: 'Profiles',
        text: 'Public or private profiles with an avatar, about section, followers, following and a personal feed of posts.'
    },
    {
        icon: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.1-3.1a2 2 0 0 0-2.8 0L6 21"/>',
        title: 'Posts',
        text: 'Share text, images and videos with location, tagged people and fine-grained privacy, from public to selected friends.'
    },
    {
        icon: '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>',
        title: 'Reactions and comments',
        text: 'Like or dislike posts, comment with images and GIFs, reply in threads and like comments.'
    },
    {
        icon: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
        title: 'Groups',
        text: 'Create groups, invite members, handle join requests, post inside the group and chat in real time.'
    },
    {
        icon: '<rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4"/><path d="M8 2v4"/><path d="M3 10h18"/>',
        title: 'Events',
        text: 'Plan group events and let members respond so everyone knows who is coming.'
    },
    {
        icon: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>',
        title: 'Private chat',
        text: 'One-to-one conversations with emojis, images, shared posts and shared profiles, with live typing status.'
    },
    {
        icon: '<path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/>',
        title: 'Notifications',
        text: 'Live alerts for follows, reactions, comments, mentions, events and messages, with controls for what you receive.'
    },
    {
        icon: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
        title: 'Search and themes',
        text: 'Find people and groups quickly and pick the look that suits you, including light, dark, blue and pink themes.'
    }
];

const STACK = ['Vue 3', 'Vue Router', 'Go', 'SQLite', 'WebSockets', 'SQL migrations'];

const WORDS = ['connect', 'share', 'chat', 'plan', 'belong'];

const DEMO = [
    { from: 'them', text: 'Hey! Are you coming to the group event?' },
    { from: 'me', text: 'Just replied yes 🎉' },
    { from: 'them', text: 'Perfect, see you there!' }
];

const SWATCHES = {
    light: '#ffffff',
    dark: '#121212',
    blue: '#5b9cff',
    pink: '#e88ca8'
};

const CONFETTI_COLORS = ['#ffd23f', '#ff6b9d', '#4ecdc4', '#2d8cf0', '#ff9f43'];

const reducedMotion =
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const activeTab = ref('about');
const sessionState = ref('checking');
const currentTheme = ref(getThemeCookie());
const tabsEl = ref(null);
const typed = ref(reducedMotion ? WORDS[0] : '');
const mouse = ref({ x: 0, y: 0 });
const statValues = ref([0, 0, 0]);
const demoMessages = ref([]);
const demoTyping = ref(false);
const demoWho = ref('them');
const particles = ref([]);
const copiedName = ref('');

const STATS = computed(() => [
    { label: 'Core features', target: FEATURES.length },
    { label: 'Authors', target: AUTHORS.length },
    { label: 'Themes', target: THEMES.length }
]);

const activeIndex = computed(() =>
    TABS.findIndex(tab => tab.value === activeTab.value)
);

const ctaTo = computed(() =>
    sessionState.value === 'valid' ? '/home' : '/login'
);

const ctaLabel = computed(() =>
    sessionState.value === 'valid' ? 'Go to home' : 'Go to login'
);

let typeTimer = null;
let wordIndex = 0;
let deleting = false;
let mouseFrame = null;
let statFrame = null;
let demoRun = 0;
let copyTimer = null;
const demoTimers = [];
const particleTimers = [];
let particleID = 0;

async function checkSession() {
    try {
        const resp = await fetch('/api/session', {
            method: 'GET',
            credentials: 'include'
        });

        sessionState.value = resp.ok ? 'valid' : 'invalid';
    } catch {
        sessionState.value = 'invalid';
    }
}

function pickTheme(theme) {
    setTheme(theme);
    currentTheme.value = theme;
}

function scheduleType(delay) {
    typeTimer = setTimeout(tickType, delay);
}

function tickType() {
    const word = WORDS[wordIndex];

    if (!deleting) {
        typed.value = word.slice(0, typed.value.length + 1);

        if (typed.value === word) {
            deleting = true;
            scheduleType(1500);
            return;
        }

        scheduleType(90);
        return;
    }

    typed.value = word.slice(0, typed.value.length - 1);

    if (!typed.value) {
        deleting = false;
        wordIndex = (wordIndex + 1) % WORDS.length;
        scheduleType(300);
        return;
    }

    scheduleType(45);
}

function onHeroMove(event) {
    if (reducedMotion || mouseFrame) {
        return;
    }

    const el = event.currentTarget;
    const clientX = event.clientX;
    const clientY = event.clientY;

    mouseFrame = requestAnimationFrame(() => {
        mouseFrame = null;

        const rect = el.getBoundingClientRect();

        mouse.value = {
            x: ((clientX - rect.left) / rect.width - 0.5) * 2,
            y: ((clientY - rect.top) / rect.height - 0.5) * 2
        };
    });
}

function onHeroLeave() {
    mouse.value = { x: 0, y: 0 };
}

function countUp() {
    const targets = STATS.value.map(stat => stat.target);

    if (reducedMotion) {
        statValues.value = targets;
        return;
    }

    const duration = 1300;
    const start = performance.now();

    function step(now) {
        const progress = Math.min((now - start) / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3);

        statValues.value = targets.map(target => Math.round(target * eased));

        if (progress < 1) {
            statFrame = requestAnimationFrame(step);
        } else {
            statFrame = null;
        }
    }

    statFrame = requestAnimationFrame(step);
}

function wait(ms, run) {
    return new Promise(resolve => {
        const timer = setTimeout(() => resolve(run === demoRun), ms);
        demoTimers.push(timer);
    });
}

async function playDemo() {
    const run = ++demoRun;

    demoMessages.value = [];
    demoTyping.value = false;

    if (reducedMotion) {
        demoMessages.value = [...DEMO];
        return;
    }

    for (const message of DEMO) {
        demoWho.value = message.from;
        demoTyping.value = true;

        if (!(await wait(message.from === 'me' ? 800 : 1200, run))) {
            return;
        }

        demoTyping.value = false;
        demoMessages.value.push(message);

        if (!(await wait(500, run))) {
            return;
        }
    }
}

function burst(event, count = 24) {
    if (reducedMotion) {
        return;
    }

    let x = event.clientX;
    let y = event.clientY;

    if (!x && !y) {
        const rect = event.currentTarget.getBoundingClientRect();
        x = rect.left + rect.width / 2;
        y = rect.top + rect.height / 2;
    }

    const created = [];

    for (let i = 0; i < count; i++) {
        const angle = Math.random() * Math.PI * 2;
        const distance = 60 + Math.random() * 110;

        created.push({
            id: ++particleID,
            x,
            y,
            dx: Math.round(Math.cos(angle) * distance),
            dy: Math.round(Math.sin(angle) * distance - 40),
            rot: Math.round(Math.random() * 720 - 360),
            size: 6 + Math.round(Math.random() * 6),
            color: CONFETTI_COLORS[Math.floor(Math.random() * CONFETTI_COLORS.length)]
        });
    }

    const ids = new Set(created.map(item => item.id));

    particles.value = [...particles.value, ...created];

    const timer = setTimeout(() => {
        particles.value = particles.value.filter(item => !ids.has(item.id));
    }, 1100);

    particleTimers.push(timer);
}

function onTilt(event) {
    if (reducedMotion) {
        return;
    }

    const el = event.currentTarget;
    const rect = el.getBoundingClientRect();
    const px = (event.clientX - rect.left) / rect.width - 0.5;
    const py = (event.clientY - rect.top) / rect.height - 0.5;

    el.style.setProperty('--rx', `${(-py * 5).toFixed(2)}deg`);
    el.style.setProperty('--ry', `${(px * 5).toFixed(2)}deg`);
}

function resetTilt(event) {
    const el = event.currentTarget;

    el.style.setProperty('--rx', '0deg');
    el.style.setProperty('--ry', '0deg');
}

async function copyGitea(author) {
    try {
        await navigator.clipboard.writeText(author.gitea);
    } catch {
        return;
    }

    copiedName.value = author.name;

    if (copyTimer) {
        clearTimeout(copyTimer);
    }

    copyTimer = setTimeout(() => {
        copiedName.value = '';
        copyTimer = null;
    }, 1500);
}

function goToAuthors() {
    activeTab.value = 'authors';

    nextTick(() => {
        tabsEl.value?.scrollIntoView({
            behavior: reducedMotion ? 'auto' : 'smooth',
            block: 'start'
        });
    });
}

function initials(name) {
    return name
        .split(' ')
        .filter(Boolean)
        .slice(0, 2)
        .map(part => part.charAt(0).toUpperCase())
        .join('');
}

function onTabKey(event) {
    const last = TABS.length - 1;
    let next = activeIndex.value;

    if (event.key === 'ArrowRight') {
        next = activeIndex.value === last ? 0 : activeIndex.value + 1;
    } else if (event.key === 'ArrowLeft') {
        next = activeIndex.value === 0 ? last : activeIndex.value - 1;
    } else if (event.key === 'Home') {
        next = 0;
    } else if (event.key === 'End') {
        next = last;
    } else {
        return;
    }

    event.preventDefault();
    activeTab.value = TABS[next].value;
    document.getElementById(`tab-${TABS[next].value}`)?.focus();
}

watch(activeTab, value => {
    if (value === 'about') {
        nextTick(playDemo);
    }
});

onMounted(() => {
    checkSession();
    countUp();
    playDemo();

    if (!reducedMotion) {
        scheduleType(600);
    }
});

onBeforeUnmount(() => {
    demoRun++;

    if (typeTimer) {
        clearTimeout(typeTimer);
    }

    if (copyTimer) {
        clearTimeout(copyTimer);
    }

    if (mouseFrame) {
        cancelAnimationFrame(mouseFrame);
    }

    if (statFrame) {
        cancelAnimationFrame(statFrame);
    }

    demoTimers.forEach(clearTimeout);
    particleTimers.forEach(clearTimeout);
});
</script>

<template>
    <div class="welcome-page">
        <div class="confetti-layer" aria-hidden="true">
            <span
                v-for="piece in particles"
                :key="piece.id"
                class="confetti"
                :style="{
                    left: piece.x + 'px',
                    top: piece.y + 'px',
                    width: piece.size + 'px',
                    height: piece.size + 'px',
                    background: piece.color,
                    '--dx': piece.dx + 'px',
                    '--dy': piece.dy + 'px',
                    '--rot': piece.rot + 'deg'
                }"
            ></span>
        </div>

        <header class="welcome-top">
            <button type="button" class="brand" @click="burst($event, 30)">
                <span class="brand-mark" aria-hidden="true">✦</span>
                Social Network
            </button>

            <nav class="top-actions">
                <div class="theme-switch" role="group" aria-label="Choose theme">
                    <button
                        v-for="theme in THEMES"
                        :key="theme.value"
                        type="button"
                        class="swatch"
                        :class="{ active: currentTheme === theme.value }"
                        :style="{ background: SWATCHES[theme.value] }"
                        :title="theme.label"
                        :aria-label="`Use ${theme.label} theme`"
                        :aria-pressed="currentTheme === theme.value"
                        @click="pickTheme(theme.value)"
                    ></button>
                </div>

                <RouterLink
                    v-if="sessionState !== 'checking'"
                    :to="ctaTo"
                    class="btn primary"
                >
                    {{ ctaLabel }}
                </RouterLink>

                <span v-else class="btn placeholder" aria-hidden="true"></span>
            </nav>
        </header>

        <main class="welcome-main">
            <section
                class="hero"
                :style="{ '--mx': mouse.x, '--my': mouse.y }"
                @mousemove="onHeroMove"
                @mouseleave="onHeroLeave"
            >
                <div class="shapes" aria-hidden="true">
                    <span class="shape circle" style="--d: 26"></span>
                    <span class="shape square" style="--d: -34"></span>
                    <span class="shape triangle" style="--d: 18"></span>
                    <span class="shape ring" style="--d: -22"></span>
                    <span class="shape plus" style="--d: 30">+</span>
                </div>

                <p class="eyebrow">WELCOME</p>

                <h1 aria-label="A place to connect, share and talk with the people who matter.">
                    <span aria-hidden="true">
                        A place to
                        <span class="word">{{ typed }}<span class="caret"></span></span>
                        with the people who matter.
                    </span>
                </h1>

                <div class="hero-actions">
                    <RouterLink
                        v-if="sessionState !== 'checking'"
                        :to="ctaTo"
                        class="btn primary big"
                    >
                        {{ ctaLabel }}
                    </RouterLink>

                    <button type="button" class="btn big" @click="goToAuthors">
                        Meet the authors
                    </button>
                </div>
            </section>

            <section class="stats" aria-label="Project at a glance">
                <div
                    v-for="(stat, index) in STATS"
                    :key="stat.label"
                    class="stat"
                >
                    <strong>{{ statValues[index] }}</strong>
                    <span>{{ stat.label }}</span>
                </div>
            </section>

            <div
                ref="tabsEl"
                class="tabs"
                role="tablist"
                aria-label="Welcome sections"
            >
                <button
                    v-for="tab in TABS"
                    :id="`tab-${tab.value}`"
                    :key="tab.value"
                    type="button"
                    role="tab"
                    class="tab"
                    :class="{ active: activeTab === tab.value }"
                    :aria-selected="activeTab === tab.value"
                    :aria-controls="`panel-${tab.value}`"
                    :tabindex="activeTab === tab.value ? 0 : -1"
                    @click="activeTab = tab.value"
                    @keydown="onTabKey"
                >
                    {{ tab.label }}
                </button>
            </div>

            <Transition name="swap" mode="out-in">
                <section
                    v-if="activeTab === 'about'"
                    id="panel-about"
                    key="about"
                    class="panel about-grid"
                    role="tabpanel"
                    aria-labelledby="tab-about"
                >
                    <article class="card about-card">
                        <h2>About the project</h2>

                        <p>
                            Social Network is a full-stack social platform where people
                            build a profile, follow each other, publish posts and take
                            part in groups. Everything updates in real time, so chats,
                            notifications and typing indicators feel instant.
                        </p>

                        <p>
                            Privacy is part of the design. Profiles can be public or
                            private, every post has its own audience, and each person
                            decides who can message them and which notifications they
                            want to receive.
                        </p>

                        <p>
                            The frontend is a single-page Vue application that talks to a
                            Go backend over a JSON API and WebSockets, with data stored
                            in SQLite and managed through versioned migrations.
                        </p>

                        <div class="chips" aria-label="Technology stack">
                            <span v-for="item in STACK" :key="item" class="chip">
                                {{ item }}
                            </span>
                        </div>
                    </article>

                    <aside class="card demo-card" aria-label="Chat preview">
                        <header class="demo-header">
                            <span class="demo-dot"></span>
                            <strong>Live chat preview</strong>
                        </header>

                        <div class="demo-body" aria-live="polite">
                            <div
                                v-for="(message, index) in demoMessages"
                                :key="index"
                                class="bubble"
                                :class="message.from"
                            >
                                {{ message.text }}
                            </div>

                            <div
                                v-if="demoTyping"
                                class="bubble typing"
                                :class="demoWho"
                                aria-label="Typing"
                            >
                                <i></i><i></i><i></i>
                            </div>
                        </div>

                        <button type="button" class="btn small" @click="playDemo">
                            Replay
                        </button>
                    </aside>
                </section>

                <section
                    v-else-if="activeTab === 'authors'"
                    id="panel-authors"
                    key="authors"
                    class="panel authors"
                    role="tabpanel"
                    aria-labelledby="tab-authors"
                >
                    <article
                        v-for="(author, index) in AUTHORS"
                        :key="author.name"
                        class="card author-card"
                        :style="{ '--i': index }"
                        @mousemove="onTilt"
                        @mouseleave="resetTilt"
                    >
                        <div class="author-info">
                            <h2>{{ author.name }}</h2>

                            <p class="author-line">
                                <span class="label">Gitea</span>
                                <span class="value">{{ author.gitea }}</span>
                                <button
                                    type="button"
                                    class="copy"
                                    :class="{ done: copiedName === author.name }"
                                    @click="copyGitea(author)"
                                >
                                    {{ copiedName === author.name ? 'Copied!' : 'Copy' }}
                                </button>
                            </p>

                            <p class="author-line">
                                <span class="label">GitHub</span>
                                <a
                                    class="value link"
                                    :href="author.github"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    {{ author.github }}
                                </a>
                            </p>
                        </div>

                        <button
                            type="button"
                            class="author-photo"
                            :aria-label="`Celebrate ${author.name}`"
                            @click="burst($event, 28)"
                        >
                            <img
                                v-if="author.image"
                                :src="author.image"
                                :alt="author.name"
                            />
                            <span v-else>{{ initials(author.name) }}</span>
                        </button>
                    </article>
                </section>

                <section
                    v-else
                    id="panel-features"
                    key="features"
                    class="panel features"
                    role="tabpanel"
                    aria-labelledby="tab-features"
                >
                    <article
                        v-for="(feature, index) in FEATURES"
                        :key="feature.title"
                        class="card feature-card"
                        :style="{ '--i': index }"
                    >
                        <div class="feature-top">
                            <button
                                type="button"
                                class="feature-icon"
                                :aria-label="`Celebrate ${feature.title}`"
                                @click="burst($event, 20)"
                            >
                                <svg
                                    viewBox="0 0 24 24"
                                    width="24"
                                    height="24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    aria-hidden="true"
                                    v-html="feature.icon"
                                ></svg>
                            </button>

                            <span class="feature-number">
                                {{ String(index + 1).padStart(2, '0') }}
                            </span>
                        </div>

                        <h3>{{ feature.title }}</h3>
                        <p>{{ feature.text }}</p>
                    </article>
                </section>
            </Transition>
        </main>
    </div>
</template>

<style scoped>
.welcome-page {
    min-height: 100vh;
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    overflow-x: hidden;
    background: var(--page-background);
    color: var(--font-color);
}

.confetti-layer {
    position: fixed;
    inset: 0;
    z-index: 100;
    pointer-events: none;
    overflow: hidden;
}

.confetti {
    position: absolute;
    border-radius: 2px;
    animation: confetti-fly 1s cubic-bezier(0.2, 0.7, 0.4, 1) forwards;
}

@keyframes confetti-fly {
    0% {
        opacity: 1;
        transform: translate(0, 0) rotate(0deg);
    }

    100% {
        opacity: 0;
        transform: translate(var(--dx), calc(var(--dy) + 140px)) rotate(var(--rot));
    }
}

.welcome-top {
    position: sticky;
    top: 0;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 14px clamp(16px, 4vw, 40px);
    border-bottom: 2px solid var(--main-color);
    background: var(--bg-color);
}

.brand {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 0;
    border: none;
    background: none;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 22px;
    cursor: pointer;
}

.brand-mark {
    display: inline-block;
    color: var(--input-focus);
    transition: transform 0.4s ease;
}

.brand:hover .brand-mark {
    transform: rotate(180deg) scale(1.2);
}

.top-actions {
    display: flex;
    align-items: center;
    gap: 16px;
}

.theme-switch {
    display: flex;
    gap: 8px;
}

.swatch {
    width: 22px;
    height: 22px;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    cursor: pointer;
    transition: transform 0.15s ease;
}

.swatch:hover {
    transform: scale(1.2) rotate(15deg);
}

.swatch.active {
    transform: scale(1.2);
    box-shadow: 0 0 0 3px var(--bg-color), 0 0 0 5px var(--input-focus);
}

.btn {
    min-height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.btn.big {
    min-height: 46px;
    padding: 0 26px;
    font-size: 12px;
}

.btn.small {
    min-height: 32px;
    padding: 0 14px;
    font-size: 10px;
}

.btn.placeholder {
    min-width: 120px;
    border-style: dashed;
    box-shadow: none;
    opacity: 0.4;
    cursor: default;
}

.btn.primary {
    background: var(--input-focus);
    color: #fff;
}

.btn:not(.placeholder):hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.btn:not(.placeholder):active {
    transform: translate(3px, 3px);
    box-shadow: none;
}

.welcome-main {
    width: min(1100px, 100%);
    box-sizing: border-box;
    margin: 0 auto;
    padding: clamp(24px, 5vw, 56px) clamp(16px, 4vw, 40px) 64px;
}

.hero {
    position: relative;
    margin-bottom: 36px;
    padding: clamp(12px, 3vw, 32px) 0;
}

.shapes {
    position: absolute;
    inset: 0;
    pointer-events: none;
}

.shape {
    position: absolute;
    translate:
        calc(var(--mx, 0) * var(--d) * 1px)
        calc(var(--my, 0) * var(--d) * 1px);
    transition: translate 0.25s ease-out;
    animation: float 6s ease-in-out infinite;
}

.shape.circle {
    top: 6%;
    right: 12%;
    width: 64px;
    height: 64px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
}

.shape.square {
    top: 52%;
    right: 4%;
    width: 44px;
    height: 44px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: #ffd23f;
    box-shadow: 4px 4px var(--main-color);
    animation-delay: -2s;
}

.shape.triangle {
    top: 14%;
    right: 32%;
    width: 46px;
    height: 46px;
    background: #ff6b9d;
    clip-path: polygon(50% 0, 100% 100%, 0 100%);
    animation-delay: -4s;
}

.shape.ring {
    top: 70%;
    right: 26%;
    width: 52px;
    height: 52px;
    border: 6px solid #4ecdc4;
    border-radius: 50%;
    animation-delay: -1s;
}

.shape.plus {
    top: 2%;
    right: 46%;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 40px;
    font-weight: 700;
    line-height: 1;
    animation-delay: -3s;
}

@keyframes float {
    0%,
    100% {
        transform: translateY(0) rotate(0deg);
    }

    50% {
        transform: translateY(-14px) rotate(8deg);
    }
}

.eyebrow {
    position: relative;
    margin: 0 0 10px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 3px;
}

.hero h1 {
    position: relative;
    max-width: 760px;
    min-height: 2.4em;
    margin: 0;
    font-family: "Liter", serif;
    font-size: clamp(28px, 5vw, 52px);
    font-weight: 400;
    line-height: 1.2;
}

.word {
    display: inline-block;
    min-width: 2ch;
    padding: 0 10px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: #fff;
}

.caret {
    display: inline-block;
    width: 3px;
    height: 0.9em;
    margin-left: 3px;
    background: currentColor;
    vertical-align: -0.05em;
    animation: blink 0.9s steps(1) infinite;
}

@keyframes blink {
    50% {
        opacity: 0;
    }
}

.hero-actions {
    position: relative;
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    margin-top: 28px;
}

.stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
    margin-bottom: 36px;
}

.stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 16px 8px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    transition: transform 0.15s ease;
}

.stat:hover {
    transform: translateY(-4px) rotate(-1deg);
}

.stat strong {
    font-family: "Liter", serif;
    font-size: clamp(28px, 5vw, 44px);
    font-weight: 400;
    line-height: 1;
    color: var(--input-focus);
}

.stat span {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 28px;
    scroll-margin-top: 80px;
}

.tab {
    min-height: 40px;
    padding: 0 22px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.tab:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.tab.active {
    background: var(--input-focus);
    color: #fff;
    transform: translate(3px, 3px);
    box-shadow: none;
}

.tab:focus-visible,
.btn:focus-visible,
.link:focus-visible,
.brand:focus-visible,
.swatch:focus-visible,
.copy:focus-visible,
.author-photo:focus-visible,
.feature-icon:focus-visible {
    outline: 3px solid var(--input-focus);
    outline-offset: 3px;
}

.swap-enter-active,
.swap-leave-active {
    transition: opacity 0.18s ease, transform 0.18s ease;
}

.swap-enter-from {
    opacity: 0;
    transform: translateY(14px);
}

.swap-leave-to {
    opacity: 0;
    transform: translateY(-8px);
}

.card {
    box-sizing: border-box;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.about-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
    gap: 28px;
    align-items: start;
}

.about-card {
    padding: clamp(20px, 4vw, 40px);
}

.about-card h2 {
    margin: 0 0 18px;
    font-family: "Liter", serif;
    font-size: 28px;
    font-weight: 400;
}

.about-card p {
    margin: 0 0 16px;
    color: var(--font-color-sub);
    font-size: 14px;
    line-height: 1.7;
}

.chips {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 24px;
}

.chip {
    padding: 6px 12px;
    border: 2px solid var(--main-color);
    border-radius: 999px;
    background: var(--page-background);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    cursor: default;
    transition: transform 0.15s ease, background 0.15s ease, color 0.15s ease;
}

.chip:hover {
    background: var(--input-focus);
    color: #fff;
    transform: translateY(-3px) rotate(-3deg);
}

.demo-card {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px;
}

.demo-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 12px;
    border-bottom: 2px solid var(--page-background);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.demo-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #4ecdc4;
    animation: pulse 1.6s ease-in-out infinite;
}

@keyframes pulse {
    0%,
    100% {
        box-shadow: 0 0 0 0 rgba(78, 205, 196, 0.6);
    }

    50% {
        box-shadow: 0 0 0 7px rgba(78, 205, 196, 0);
    }
}

.demo-body {
    min-height: 190px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
    border-radius: 6px;
    background: var(--page-background);
}

.bubble {
    max-width: 82%;
    padding: 9px 13px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    font-size: 12px;
    line-height: 1.5;
    animation: bubble-in 0.3s ease-out;
}

.bubble.them {
    align-self: flex-start;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.bubble.me {
    align-self: flex-end;
    background: var(--input-focus);
    box-shadow: 3px 3px var(--main-color);
    color: #fff;
}

.bubble.typing {
    display: flex;
    gap: 4px;
    padding: 12px 14px;
}

.bubble.typing i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    animation: typing-bounce 1s infinite ease-in-out;
}

.bubble.typing i:nth-child(2) {
    animation-delay: 0.15s;
}

.bubble.typing i:nth-child(3) {
    animation-delay: 0.3s;
}

@keyframes typing-bounce {
    0%,
    60%,
    100% {
        opacity: 0.3;
        transform: translateY(0);
    }

    30% {
        opacity: 1;
        transform: translateY(-4px);
    }
}

@keyframes bubble-in {
    from {
        opacity: 0;
        transform: translateY(8px) scale(0.92);
    }

    to {
        opacity: 1;
        transform: translateY(0) scale(1);
    }
}

.authors {
    display: flex;
    flex-direction: column;
    gap: 28px;
}

.author-card {
    min-height: 220px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 32px;
    padding: clamp(20px, 4vw, 36px);
    transform: perspective(900px) rotateX(var(--rx, 0deg)) rotateY(var(--ry, 0deg));
    transition: transform 0.15s ease-out, box-shadow 0.15s ease-out;
    animation: rise 0.5s ease-out both;
    animation-delay: calc(var(--i) * 0.1s);
}

.author-card:hover {
    box-shadow: 10px 10px var(--main-color);
}

@keyframes rise {
    from {
        opacity: 0;
        translate: 0 24px;
    }

    to {
        opacity: 1;
        translate: 0 0;
    }
}

.author-info {
    min-width: 0;
    flex: 1;
}

.author-info h2 {
    margin: 0 0 20px;
    font-family: "Liter", serif;
    font-size: clamp(24px, 4vw, 34px);
    font-weight: 400;
    overflow-wrap: anywhere;
}

.author-line {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 10px;
    margin: 0 0 10px;
}

.label {
    min-width: 62px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.value {
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    overflow-wrap: anywhere;
}

.link {
    color: var(--input-focus);
    text-decoration: underline;
    text-underline-offset: 3px;
}

.copy {
    padding: 2px 9px;
    border: 2px solid var(--main-color);
    border-radius: 4px;
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
}

.copy:hover,
.copy.done {
    background: var(--input-focus);
    color: #fff;
}

.author-photo {
    flex-shrink: 0;
    width: clamp(120px, 20vw, 180px);
    aspect-ratio: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--page-background);
    box-shadow: 4px 4px var(--main-color);
    font-family: "Liter", serif;
    font-size: 48px;
    color: var(--font-color-sub);
    cursor: pointer;
    transition: transform 0.2s ease;
}

.author-photo:hover {
    transform: rotate(3deg) scale(1.05);
}

.author-photo:active {
    transform: rotate(-3deg) scale(0.96);
}

.author-photo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.features {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 24px;
}

.feature-card {
    padding: 22px;
    transition: transform 0.15s ease, box-shadow 0.15s ease;
    animation: rise 0.5s ease-out both;
    animation-delay: calc(var(--i) * 0.06s);
}

.feature-card:hover {
    transform: translate(-3px, -3px) rotate(-0.6deg);
    box-shadow: 9px 9px var(--main-color);
}

.feature-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
}

.feature-icon {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--page-background);
    box-shadow: 3px 3px var(--main-color);
    color: var(--input-focus);
    cursor: pointer;
    transition: transform 0.2s ease, background 0.15s ease, color 0.15s ease;
}

.feature-icon svg {
    display: block;
}

.feature-card:hover .feature-icon {
    background: var(--input-focus);
    color: #fff;
}

.feature-card:hover .feature-icon {
    animation: wiggle 0.6s ease-in-out;
}

.feature-icon:active {
    transform: scale(0.88);
}

@keyframes wiggle {
    0%,
    100% {
        transform: rotate(0deg);
    }

    25% {
        transform: rotate(-12deg) scale(1.1);
    }

    75% {
        transform: rotate(12deg) scale(1.1);
    }
}

.feature-number {
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 2px;
}

.feature-card h3 {
    margin: 0 0 8px;
    font-family: "Liter", serif;
    font-size: 20px;
    font-weight: 400;
}

.feature-card p {
    margin: 0;
    color: var(--font-color-sub);
    font-size: 13px;
    line-height: 1.6;
}

@media (max-width: 820px) {
    .about-grid {
        grid-template-columns: 1fr;
    }

    .shape.triangle,
    .shape.plus,
    .shape.ring {
        display: none;
    }
}

@media (max-width: 640px) {
    .welcome-top {
        flex-wrap: wrap;
    }

    .top-actions {
        gap: 12px;
    }

    .shapes {
        opacity: 0.35;
    }

    .stats {
        gap: 10px;
    }

    .author-card {
        flex-direction: column-reverse;
        align-items: flex-start;
        gap: 20px;
    }

    .author-photo {
        width: 100%;
        max-width: 200px;
    }

    .tab {
        flex: 1;
        padding: 0 12px;
    }

    .hero-actions .btn {
        flex: 1;
    }
}

@media (prefers-reduced-motion: reduce) {
    *,
    *::before,
    *::after {
        animation: none !important;
        transition: none !important;
    }

    .shape {
        translate: none;
    }
}
</style>