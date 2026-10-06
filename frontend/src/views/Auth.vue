<script setup lang="ts">
import AuthContainer from '@/components/auth/AuthContainer.vue'
import WelcomeSection from '@/components/auth/WelcomeSection.vue'
</script>

<template>
    <main class="auth-page">
        <WelcomeSection />
        <AuthContainer />
    </main>
</template>

<style scoped>
.auth-page {
    min-height: 100vh;
    min-height: 100dvh;
    width: 100%;

    display: grid;
    grid-template-columns: 1.15fr 0.85fr;

    background: var(--page-background);

    overflow: hidden;
}

@media (max-width: 900px) {
    .auth-page {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 640px) {
    .auth-page {
        display: block;
    }
}
</style>