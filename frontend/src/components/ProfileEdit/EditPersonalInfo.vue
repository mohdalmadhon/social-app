<script setup>
import { reactive } from 'vue';
import { LIMITS } from '@/helpers/limits';
import {
    validateName,
    validateUsername,
    validateEmail,
    validatePassword,
    validateAbout
} from '@/helpers/validators/registration';

import FormField from '@/components/ProfileEdit/FormField.vue';
import AvatarUploader from '@/components/ProfileEdit/AvatarUploader.vue';
import { updateUserInfo } from '@/api/users/editProfile';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    firstName: String,
    lastName: String,
    username: String,
    email: String,
    bio: String,
    avatar_path: String,
    isPrivate: Boolean
});

const form = reactive({
    FirstName: props.firstName || '',
    LastName: props.lastName || '',
    Username: props.username || '',
    Email: props.email || '',
    About: props.bio || '',
    Password: '',
    IsPrivate: props.isPrivate || false
});

const errors = reactive({
    FirstName: '',
    LastName: '',
    Username: '',
    Email: '',
    Password: '',
    About: ''
});

function validateForm() {
    errors.FirstName = validateName(form.FirstName.trim());
    errors.LastName = validateName(form.LastName.trim());
    errors.Username = validateUsername(form.Username.trim());
    errors.Email = validateEmail(form.Email.trim());
    errors.Password = form.Password ? validatePassword(form.Password) : '';
    errors.About = validateAbout(form.About);

    return !Object.values(errors).some(Boolean);
}

function togglePrivacy() {
    form.IsPrivate = !form.IsPrivate;
}

async function updateInfo() {
    if (!validateForm()) {
        addNotification('Please fix the highlighted fields', 'error')
        return;
    }

    if (form.IsPrivate) {
        form.IsPrivate = 1
    } else {
        form.IsPrivate = 0
    }
    try {
        const result = await updateUserInfo(form);
        if (!result.status) {
            addNotification(result.message, 'error')
            return;
        }
        addNotification(result.message, 'success')
    } catch (err) {
        addNotification(err.message, 'error')
    }

}
</script>

<template>
    <section class="edit-section">
        <div class="section-heading">
            <p class="eyebrow">PROFILE</p>
            <h2>Personal Info</h2>
        </div>

        <div class="edit-card">
            <AvatarUploader :src="props.avatar_path" @change="onAvatarChange" />

            <div class="privacy-setting">
                <div>
                    <p class="privacy-title">
                        {{ form.IsPrivate ? 'Private account' : 'Public account' }}
                    </p>

                    <p class="privacy-description">
                        {{
                            form.IsPrivate
                                ? 'Only approved followers can see your posts.'
                                : 'Anyone can see your posts and profile.'
                        }}
                    </p>
                </div>

                <button type="button" class="privacy-button" :class="{ private: form.IsPrivate }"
                    @click="togglePrivacy">
                    {{ form.IsPrivate ? 'Make public' : 'Make private' }}
                </button>
            </div>

            <form @submit.prevent="updateInfo">
                <div class="field-grid">
                    <FormField id="firstName" label="First name" v-model="form.FirstName" placeholder="First name"
                        :maxlength="LIMITS.name" :error="errors.FirstName" />

                    <FormField id="lastName" label="Last name" v-model="form.LastName" placeholder="Last name"
                        :maxlength="LIMITS.name" :error="errors.LastName" />

                    <FormField id="username" label="Username" v-model="form.Username" placeholder="Username"
                        :maxlength="LIMITS.username" :error="errors.Username" />

                    <FormField id="email" label="Email" type="email" v-model="form.Email" placeholder="Email address"
                        :maxlength="LIMITS.email" :error="errors.Email" />

                    <FormField id="password" label="Password" type="password" v-model="form.Password"
                        placeholder="****************" :maxlength="LIMITS.password" :error="errors.Password" />
                </div>

                <FormField id="bio" label="Bio" type="textarea" v-model="form.About"
                    placeholder="Tell people about yourself" :maxlength="LIMITS.about" :error="errors.About" />

                <button type="submit" class="confirm-button">
                    Confirm changes
                </button>
            </form>
        </div>
    </section>
</template>

<style scoped>
.edit-section {
    scroll-margin-top: 100px;
    width: 100%;
    max-width: 100%;
}

.section-heading {
    margin-bottom: clamp(14px, 2.5vw, 20px);
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
    letter-spacing: 2px;
}

h2 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: clamp(20px, 4vw, 29px);
}

.edit-card {
    display: flex;
    flex-direction: column;
    gap: clamp(16px, 2.5vw, 22px);
    padding: clamp(14px, 3vw, 25px);
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);
    max-width: 100%;
    box-sizing: border-box;
}

.privacy-setting {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: clamp(12px, 2.5vw, 20px);
    padding: clamp(12px, 2.5vw, 18px);
    border: 2px solid var(--main-color);
    border-radius: 5px;
}

.privacy-title {
    margin: 0 0 5px;
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(11px, 1.6vw, 12px);
    font-weight: 700;
}

.privacy-description {
    margin: 0;
    font-size: clamp(11px, 1.6vw, 12px);
    opacity: 0.7;
}

.privacy-button {
    flex-shrink: 0;
    padding: clamp(9px, 1.8vw, 11px) clamp(14px, 2.5vw, 18px);
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 3px 3px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(9px, 1.4vw, 10px);
    font-weight: 600;
    cursor: pointer;
}

.privacy-button:hover {
    transform: translate(-1px, -1px);
}

.privacy-button.private {
    background: var(--main-color);
}

.field-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
    gap: clamp(14px, 2.5vw, 20px);
}

.confirm-button {
    align-self: flex-start;
    padding: clamp(11px, 2vw, 13px) clamp(16px, 3vw, 22px);
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(9px, 1.4vw, 10px);
    font-weight: 600;
}

.confirm-button:hover {
    transform: translate(-1px, -1px);
}
</style>