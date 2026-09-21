<script setup>
import { onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { logout } from '@/api/auth/auth';
import { addNotification } from '@/data/notifications';
import { unreadNotificationCount, refreshUnreadNotificationCount } from '@/data/notificationCount';

const route = useRoute();

onMounted(() => {
    refreshUnreadNotificationCount();
});

async function logoutHandler() {
    try {
        await logout();
    } catch (err) {
        addNotification(err, 'error');
    }
}
</script>

<template>
    <aside class="side-navigation">
        <div class="side-inner">
            <nav>
                <a href="/home" :class="{ active: route.path === '/home' }">
                    <span class="icon">⌂</span>
                    <span class="label">Home</span>
                </a>

                <a href="/me" :class="{ active: route.path === '/me' }">
                    <span class="icon">◉</span>
                    <span class="label">Profile</span>
                </a>


                <a href="/groups" :class="{ active: route.path === '/groups' }">
                    <span class="icon">▦</span>
                    <span class="label">Groups</span>
                </a>

                <a href="/notifications" :class="{ active: route.path === '/notifications' }">
                    <span class="icon">♢</span>
                    <span class="label">notifications</span>

                    <span v-if="unreadNotificationCount > 0" class="badge">
                        {{ unreadNotificationCount > 99 ? '99+' : unreadNotificationCount }}
                    </span>
                </a>

                <a href="/chats" :class="{ active: route.path === '/chats' }">
                    <span class="icon">✉</span>
                    <span class="label">chats</span>
                </a>


            </nav>

            <div class="side-bottom">
                <a href="/settings" :class="{ active: route.path === '/settings' }">
                    <span class="icon">⚙</span>
                    <span class="label">Settings</span>
                </a>

                <a href="#" @click.prevent="logoutHandler">
                    <span class="icon">↪</span>
                    <span class="label">Log out</span>
                </a>
            </div>
        </div>
    </aside>
</template>

<style scoped>
/* the aside only reserves the sidebar's width in the page layout,
   the sidebar itself (.side-inner) stays fixed while the page scrolls */
.side-navigation {
    flex-shrink: 0;
    width: clamp(72px, 16vw, 220px);
}

.side-inner {
    position: fixed;
    z-index: 50;
    top: 64px;
    bottom: 0;
    left: 0;
    width: clamp(72px, 16vw, 220px);
    padding: clamp(16px, 2.5vw, 25px) clamp(10px, 1.5vw, 18px);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: var(--bg-color);
    border-right: 2px solid var(--main-color);
    box-sizing: border-box;
    overflow-y: auto;
}

.side-navigation nav,
.side-bottom {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.side-navigation a {
    position: relative;
    min-height: 44px;
    display: flex;
    align-items: center;
    gap: 13px;
    padding: 0 clamp(8px, 1.2vw, 13px);
    border: 2px solid transparent;
    border-radius: 5px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(10px, 1.4vw, 11px);
    font-weight: 600;
    overflow: hidden;
    text-decoration: none;
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

@media (max-width: 1000px) {
    .badge {
        position: absolute;
        top: 2px;
        right: 2px;
        margin-left: 0;
    }
}

.side-navigation a:hover {
    border-color: var(--main-color);
    background: var(--page-background);
}

.side-navigation a.active {
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

@media (max-width: 1000px) {
    .side-navigation .label {
        display: none;
    }

    .side-navigation a {
        justify-content: center;
        padding: 0;
    }
}

@media (max-width: 800px) {
    .side-navigation {
        width: 100%;
        flex-shrink: initial;
    }

    .side-inner {
        position: static;
        width: 100%;
        height: auto;
        padding: 8px 15px;
        border-right: 0;
        border-bottom: 2px solid var(--main-color);
        flex-direction: row;
        overflow-x: auto;
        overflow-y: hidden;
        -webkit-overflow-scrolling: touch;
    }

    .side-navigation nav,
    .side-bottom {
        flex-direction: row;
    }

    .side-navigation a {
        white-space: nowrap;
        flex-shrink: 0;
    }

    .side-navigation .label {
        display: inline;
    }

    .side-bottom {
        display: none;
    }
}
</style>
