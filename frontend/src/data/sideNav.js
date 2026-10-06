import { ref } from 'vue';

export const sideNavOpen = ref(false);

export function openSideNav() {
    sideNavOpen.value = true;
}

export function closeSideNav() {
    sideNavOpen.value = false;
}

export function toggleSideNav() {
    sideNavOpen.value = !sideNavOpen.value;
}
