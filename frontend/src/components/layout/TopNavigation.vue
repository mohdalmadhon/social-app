<script setup>
import { getUserData } from '@/api/users/personalProfile';
import { addNotification } from '@/data/notifications';
import { useSearchBox } from '@/helpers/search/useSearchBox';
import { onMounted, ref } from 'vue';

const avatar = ref('');
const { searchQuery, submitSearch } = useSearchBox();

async function getData() {
    try {
        const result = await getUserData()
        avatar.value = result.Profile.avatar;
    } catch (err) {
        addNotification(err.message || 'error', 'error')
    }

}
onMounted(getData)
</script>

<template>
    <header class="top-navigation">
        <div class="nav-left">

            <form class="search" role="search" @submit.prevent="submitSearch">
                <button class="search-submit" type="submit" aria-label="Search">⌕</button>
                <input
                    v-model="searchQuery"
                    type="text"
                    placeholder="Search"
                    aria-label="Search groups, users and posts"
                    enterkeyhint="search"
                    autocomplete="off"
                />
            </form>
        </div>
        
        <nav class="nav-links">
            <a href="/home">Home</a>
        </nav>

        <div class="nav-right">
            <a href="/post/new"><button class="nav-button">+</button></a>

            <a href="/me" class="nav-avatar">
                <img :src="`/uploads/${avatar}`" alt="">
            </a>
        </div>
    </header>
</template>

<style scoped>
.top-navigation {
    position: fixed;
    z-index: 100;
    top: 0;
    left: 0;
    width: 100%;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 clamp(12px, 3vw, 28px);
    background: var(--bg-color);
    border-bottom: 2px solid var(--main-color);
    box-sizing: border-box;
    gap: 12px;
}

.nav-left,
.nav-right,
.nav-links {
    display: flex;
    align-items: center;
    min-width: 0;
}

.nav-left {
    gap: clamp(10px, 2vw, 18px);
    flex: 1;
    min-width: 0;
}

.logo {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--input-focus);
    border: 2px solid var(--main-color);
    border-radius: 6px;
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 19px;
    font-weight: 600;
}

.search {
    width: clamp(120px, 20vw, 220px);
    height: 38px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    box-sizing: border-box;
    min-width: 0;
}

.search-submit {
    flex-shrink: 0;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--main-color);
    font-size: 20px;
    line-height: 1;
}

.search input {
    width: 100%;
    min-width: 0;
    border: 0;
    outline: 0;
    background: transparent;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.nav-links {
    gap: clamp(14px, 2.5vw, 35px);
}

.nav-links a {
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    white-space: nowrap;
}

.nav-links a:hover {
    color: var(--input-focus);
}

.nav-right {
    flex-shrink: 0;
    gap: clamp(6px, 1.5vw, 10px);
}

.nav-button {
    flex-shrink: 0;
    width: 35px;
    height: 35px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--main-color);
    font-size: 16px;
    box-shadow: 3px 3px var(--main-color);
}

.nav-button:active {
    transform: translate(3px, 3px);
    box-shadow: none;
}

.nav-avatar {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 5px;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    overflow: hidden;
    box-sizing: border-box;
    box-shadow: 3px 3px var(--main-color);
    transition: transform 0.15s ease;
}

.nav-avatar img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
    border-radius: 50%;
}

.nav-avatar:hover {
    transform: translateY(-2px);
}

.nav-avatar:active {
    transform: translate(3px, 3px);
    box-shadow: none;
}

@media (max-width: 900px) {
    .nav-links {
        display: none;
    }
}

@media (max-width: 640px) {
    .search {
        width: 100%;
    }
}

@media (max-width: 480px) {
    .nav-right > .nav-button {
        display: none;
    }
}
</style>