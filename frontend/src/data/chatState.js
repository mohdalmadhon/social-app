import { ref } from 'vue';

export let activePage = ref(null);
export let messageSent = ref(true);
export const openGroupPage = ref(null);
export const chatsSidebarOpen = ref(false);

export function openChatsSidebar() {
    chatsSidebarOpen.value = true;
}

export function closeChatsSidebar() {
    chatsSidebarOpen.value = false;
}

export function toggleChatsSidebar() {
    chatsSidebarOpen.value = !chatsSidebarOpen.value;
}
