<script setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const props = defineProps({
    code: {
        type: [Number, String],
        default: null
    }
});

const route = useRoute();
const router = useRouter();

const PRESETS = {
    400: {
        title: 'Bad request',
        message: 'Something about that request was not right.'
    },
    401: {
        title: 'Not signed in',
        message: 'You need to log in to see this page.'
    },
    403: {
        title: 'Access denied',
        message: 'You do not have permission to view this page.'
    },
    404: {
        title: 'Page not found',
        message: 'The page you are looking for does not exist or has been moved.'
    },
    500: {
        title: 'Server error',
        message: 'Something went wrong on our side. Please try again in a moment.'
    },
    503: {
        title: 'Service unavailable',
        message: 'The server is temporarily unavailable. Please try again soon.'
    }
};

const DEFAULT = {
    title: 'Something went wrong',
    message: 'An unexpected error occurred.'
};

const status = computed(() => {
    const value = Number(props.code ?? route.query.code ?? 404);

    return Number.isInteger(value) && value >= 400 && value <= 599
        ? value
        : 404;
});

const preset = computed(() => PRESETS[status.value] || DEFAULT);

const title = computed(() => preset.value.title);

const message = computed(() => {
    const custom = route.query.message;

    if (typeof custom === 'string' && custom.trim()) {
        return custom.trim().slice(0, 200);
    }

    return preset.value.message;
});

const digits = computed(() => String(status.value).split(''));

const showLogin = computed(() => status.value === 401);

const attemptedPath = computed(() => {
    if (status.value !== 404) {
        return '';
    }

    return route.fullPath;
});

function goBack() {
    if (window.history.length > 1) {
        router.back();
        return;
    }

    router.push('/home');
}

function reload() {
    window.location.reload();
}
</script>

<template>
    <main class="error-page">
        <section class="error-card" role="alert">
            <p class="eyebrow">ERROR</p>

            <div class="error-code" aria-hidden="true">
                <span v-for="(digit, index) in digits" :key="index" class="digit">
                    {{ digit }}
                </span>
            </div>

            <h1>{{ title }}</h1>

            <p class="error-message">{{ message }}</p>

            <code v-if="attemptedPath" class="error-path">
                {{ attemptedPath }}
            </code>

            <div class="error-actions">
                <RouterLink v-if="showLogin" to="/login" class="btn primary">
                    Go to login
                </RouterLink>

                <RouterLink v-else to="/home" class="btn primary">
                    Back to home
                </RouterLink>

                <button type="button" class="btn" @click="goBack">
                    Go back
                </button>

                <button
                    v-if="status >= 500"
                    type="button"
                    class="btn"
                    @click="reload"
                >
                    Try again
                </button>
            </div>
        </section>
    </main>
</template>

<style scoped>
.error-page {
    min-height: 100vh;
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    box-sizing: border-box;
    background: var(--page-background);
    color: var(--font-color);
}

.error-card {
    width: min(520px, 100%);
    box-sizing: border-box;
    padding: 40px 32px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    text-align: center;
}

.eyebrow {
    margin: 0 0 18px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 3px;
}

.error-code {
    display: flex;
    justify-content: center;
    gap: 10px;
    margin-bottom: 22px;
}

.digit {
    width: 64px;
    height: 80px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--page-background);
    box-shadow: 4px 4px var(--main-color);
    font-family: "Liter", serif;
    font-size: 44px;
    line-height: 1;
}

.digit:nth-child(2) {
    background: var(--input-focus);
    color: #fff;
}

h1 {
    margin: 0 0 10px;
    font-family: "Liter", serif;
    font-size: 26px;
    font-weight: 400;
}

.error-message {
    margin: 0 auto;
    max-width: 380px;
    color: var(--font-color-sub);
    font-size: 13px;
    line-height: 1.6;
}

.error-path {
    display: block;
    max-width: 100%;
    margin: 18px auto 0;
    padding: 8px 12px;
    box-sizing: border-box;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    overflow-wrap: anywhere;
    text-align: left;
}

.error-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 12px;
    margin-top: 28px;
}

.btn {
    min-height: 40px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.btn.primary {
    background: var(--input-focus);
    color: #fff;
}

.btn:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.btn:active {
    transform: translate(4px, 4px);
    box-shadow: none;
}

@media (max-width: 480px) {
    .error-card {
        padding: 28px 18px;
    }

    .digit {
        width: 52px;
        height: 66px;
        font-size: 34px;
    }

    .error-actions {
        flex-direction: column;
    }

    .btn {
        width: 100%;
    }
}
</style>