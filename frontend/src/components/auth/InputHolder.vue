<script setup>
defineProps(['name', 'maxLength', 'minLength', 'type', 'id', 'placeHolder', 'required']);
</script>

<template>
    <input :name="name" :maxlength="maxLength || 75" :minlength="minLength" :type="type || 'text'" :id="id"
        :placeholder="placeHolder" :required="required" class="flip-card__input">
</template>

<style>
.flip-card__input {
    width: 100%;
    height: 46px;

    padding: 5px 12px;

    border: 2px solid var(--main-color);

    border-radius: 5px;

    outline: none;

    background: var(--bg-color);

    box-shadow: 4px 4px var(--main-color);

    color: var(--font-color);

    font-family: "JetBrains Mono", monospace;

    font-size: 13px;
    font-weight: 600;

    transition: 0.2s;
}

.flip-card__input::placeholder {
    color: var(--font-color-sub);

    opacity: 0.7;
}

.flip-card__input:focus {
    border-color: var(--input-focus);

    box-shadow: 4px 4px var(--input-focus);
}
</style>