<template>
    <RouterLink to="/home" class="back-home">
        <span class="back-arrow" aria-hidden="true">←</span>
        <span>Back to home</span>
    </RouterLink>
</template>

<style scoped>
.back-home {
    align-self: flex-start;
    width: fit-content;
    min-height: 36px;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 18px;
    padding: 0 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.back-home:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.back-home:active {
    transform: translate(3px, 3px);
    box-shadow: none;
}

.back-arrow {
    font-size: 14px;
    line-height: 1;
}
</style>
