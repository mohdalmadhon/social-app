<script setup>
import { onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute } from 'vue-router';
import { logout } from '@/api/auth/auth';
import { addNotification } from '@/data/notifications';
import { unreadNotificationCount, refreshUnreadNotificationCount } from '@/data/notificationCount';
import { sideNavOpen, closeSideNav } from '@/data/sideNav';

const route = useRoute();
const desktopQuery = window.matchMedia('(min-width: 1025px)');

function handleKeydown(event) {
    if (event.key === 'Escape' && sideNavOpen.value) {
        closeSideNav();
    }
}

function handleBreakpointChange(event) {
    if (event.matches) {
        closeSideNav();
    }
}

watch(sideNavOpen, (open) => {
    document.body.style.overflow = open ? 'hidden' : '';
});

watch(() => route.fullPath, closeSideNav);

onMounted(() => {
    refreshUnreadNotificationCount();
    window.addEventListener('keydown', handleKeydown);
    desktopQuery.addEventListener('change', handleBreakpointChange);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeydown);
    desktopQuery.removeEventListener('change', handleBreakpointChange);
    document.body.style.overflow = '';
    closeSideNav();
});

async function logoutHandler() {
    closeSideNav();

    try {
        await logout();
    } catch (err) {
        addNotification(err, 'error');
    }
}
</script>

<template>
    <aside class="side-navigation" aria-hidden="true"></aside>

    <Teleport to="body">
        <div class="side-overlay" :class="{ open: sideNavOpen }" aria-hidden="true" @click="closeSideNav"></div>

        <div id="side-drawer" class="side-inner" :class="{ open: sideNavOpen }" aria-label="Main navigation">
            <button type="button" class="drawer-close" aria-label="Close navigation menu" @click="closeSideNav">
                <span class="drawer-close-title">Menu</span>
                <span class="drawer-close-icon">×</span>
            </button>

            <nav>
                <a href="/home" :class="{ active: route.path === '/home' }" @click="closeSideNav">
                    <span class="icon">⌂</span>
                    <span class="label">Home</span>
                </a>

                <a href="/post/new" class="drawer-only" :class="{ active: route.path === '/post/new' }"
                    @click="closeSideNav">
                    <span class="icon">+</span>
                    <span class="label">New post</span>
                </a>

                <a href="/me" :class="{ active: route.path === '/me' }" @click="closeSideNav">
                    <span class="icon">◉</span>
                    <span class="label">Profile</span>
                </a>

                <a href="/groups" :class="{ active: route.path === '/groups' }" @click="closeSideNav">
                    <span class="icon">▦</span>
                    <span class="label">Groups</span>
                </a>

                <a href="/notifications" :class="{ active: route.path === '/notifications' }" @click="closeSideNav">
                    <span class="icon">♢</span>
                    <span class="label">notifications</span>

                    <span v-if="unreadNotificationCount > 0" class="badge">
                        {{ unreadNotificationCount > 99 ? '99+' : unreadNotificationCount }}
                    </span>
                </a>

                <a href="/chats" :class="{ active: route.path === '/chats' }" @click="closeSideNav">
                    <span class="icon">✉</span>
                    <span class="label">chats</span>
                </a>
            </nav>

            <div class="side-bottom">
                <a href="/settings" :class="{ active: route.path === '/settings' }" @click="closeSideNav">
                    <span class="icon">⚙</span>
                    <span class="label">Settings</span>
                </a>

                <a href="#" @click.prevent="logoutHandler">
                    <span class="icon">↪</span>
                    <span class="label">Log out</span>
                </a>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.side-navigation {
    flex-shrink: 0;
    width: clamp(160px, 16vw, 220px);
}

.side-overlay {
    display: none;
}

.drawer-close {
    display: none;
}

.side-inner {
    position: fixed;
    z-index: 50;
    top: 64px;
    bottom: 0;
    left: 0;
    width: clamp(160px, 16vw, 220px);
    padding: clamp(16px, 2.5vw, 25px) clamp(10px, 1.5vw, 18px);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 16px;
    background: var(--bg-color);
    border-right: 2px solid var(--main-color);
    box-sizing: border-box;
    overflow-y: auto;
    overscroll-behavior: contain;
}

.side-inner nav,
.side-bottom {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.side-inner a {
    position: relative;
    min-height: 44px;
    display: flex;
    align-items: center;
    gap: 13px;
    padding: 0 13px;
    border: 2px solid transparent;
    border-radius: 5px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    overflow: hidden;
    text-decoration: none;
}

.side-inner a.drawer-only {
    display: none;
}

.badge {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    margin-left: auto;
    border: 2px solid var(--main-color);
    border-radius: 999px;
    background: #d9534f;
    color: #fff;
    font-size: 9px;
    font-weight: 700;
    line-height: 1;
    flex-shrink: 0;
}

.side-inner a:hover {
    border-color: var(--main-color);
    background: var(--page-background);
}

.side-inner a.active {
    border: 2px solid var(--main-color);
    background: var(--input-focus);
    color: white;
    box-shadow: 3px 3px var(--main-color);
}

.icon {
    flex-shrink: 0;
    width: 20px;
    text-align: center;
    font-size: 16px;
}

.label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

@media (max-width: 1024px) {
    .side-navigation {
        width: 0;
        flex-shrink: 0;
    }

    .side-overlay {
        display: block;
        position: fixed;
        z-index: 10000;
        inset: 0;
        background: rgba(0, 0, 0, 0.55);
        opacity: 0;
        visibility: hidden;
        transition: opacity 0.25s ease, visibility 0s linear 0.25s;
    }

    .side-overlay.open {
        opacity: 1;
        visibility: visible;
        transition: opacity 0.25s ease, visibility 0s;
    }

    .side-inner {
        z-index: 10001;
        top: 0;
        width: min(300px, 85vw);
        padding: 14px 16px 20px;
        padding-top: max(14px, env(safe-area-inset-top));
        padding-left: max(16px, env(safe-area-inset-left));
        padding-bottom: max(20px, env(safe-area-inset-bottom));
        box-shadow: 6px 0 0 rgba(0, 0, 0, 0.12);
        transform: translateX(-100%);
        visibility: hidden;
        transition: transform 0.25s ease, visibility 0s linear 0.25s;
    }

    .side-inner.open {
        transform: translateX(0);
        visibility: visible;
        transition: transform 0.25s ease, visibility 0s;
    }

    .drawer-close {
        flex-shrink: 0;
        width: 100%;
        min-height: 44px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 4px 0 13px;
        border: 0;
        border-bottom: 2px solid var(--main-color);
        border-radius: 0;
        background: transparent;
        color: var(--font-color);
        cursor: pointer;
    }

    .drawer-close-title {
        font-family: "JetBrains Mono", monospace;
        font-size: 12px;
        font-weight: 600;
        letter-spacing: 2px;
        text-transform: uppercase;
    }

    .drawer-close-icon {
        width: 36px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 2px solid var(--main-color);
        border-radius: 5px;
        background: var(--page-background);
        font-size: 22px;
        line-height: 1;
    }

    .side-inner nav {
        margin-top: 4px;
    }

    .side-inner a {
        font-size: 12px;
    }

    .side-inner a.drawer-only {
        display: flex;
    }
}

@media (max-width: 520px) {
    .side-inner {
        width: min(280px, 88vw);
    }
}

@media (prefers-reduced-motion: reduce) {
    .side-overlay,
    .side-inner {
        transition: none !important;
    }
}
</style>
