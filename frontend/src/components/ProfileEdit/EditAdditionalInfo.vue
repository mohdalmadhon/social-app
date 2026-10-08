<script setup>
import { updateAbout } from '@/api/users/editProfile';
import { addNotification } from '@/data/notifications';
import { validateAboutField } from '@/helpers/validators/editProfile.js';
import { reactive } from 'vue';
import { LIMITS } from '@/helpers/limits';
import { isLinkField, normalizeLink } from '@/helpers/links';

const props = defineProps({
    about: {
        type: Object,
        default: () => ({})
    }
});

const form = reactive({
    work: props.about?.work || '',
    education: props.about?.education || '',
    hobbies: props.about?.hobbies || '',
    interests: props.about?.interests || '',
    travel: props.about?.travel || '',
    website: props.about?.website || '',
    linkedin: props.about?.linkedin || '',
    twitter: props.about?.twitter || '',
    instagram: props.about?.instagram || ''
});

const errors = reactive({
    work: '',
    education: '',
    hobbies: '',
    interests: '',
    travel: '',
    website: '',
    linkedin: '',
    twitter: '',
    instagram: ''
});

function validateField(field) {
    errors[field] = validateAboutField(form[field], field);
}

async function confirmChanges() {
    Object.keys(form).forEach(validateField);

    const hasErrors = Object.values(errors).some(error => error !== '');

    if (hasErrors) {
        addNotification('Please fix the highlighted fields', 'error');
        return;
    }

    Object.keys(form).forEach(field => {
        if (isLinkField(field)) {
            form[field] = normalizeLink(field, form[field]);
        }
    });

    try {
        const result = await updateAbout(form);
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
            <h2>Additional Info</h2>
        </div>

        <div class="edit-card">
            <div class="field-grid">
                <div class="form-field">
                    <label for="work">Work</label>
                    <textarea
                        id="work"
                        v-model="form.work"
                        :maxlength="LIMITS.aboutField"
                        placeholder="Where do you work"
                        @input="validateField('work')"
                    ></textarea>

                    <p v-if="errors.work" class="error">
                        {{ errors.work }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="education">Education</label>
                    <textarea
                        id="education"
                        v-model="form.education"
                        :maxlength="LIMITS.aboutField"
                        placeholder="Where did you study"
                        @input="validateField('education')"
                    ></textarea>

                    <p v-if="errors.education" class="error">
                        {{ errors.education }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="hobbies">Hobbies</label>
                    <textarea
                        id="hobbies"
                        v-model="form.hobbies"
                        :maxlength="LIMITS.aboutField"
                        placeholder="Your hobbies"
                        @input="validateField('hobbies')"
                    ></textarea>

                    <p v-if="errors.hobbies" class="error">
                        {{ errors.hobbies }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="interests">Interests</label>
                    <textarea
                        id="interests"
                        v-model="form.interests"
                        :maxlength="LIMITS.aboutField"
                        placeholder="Your interests"
                        @input="validateField('interests')"
                    ></textarea>

                    <p v-if="errors.interests" class="error">
                        {{ errors.interests }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="travel">Travel</label>
                    <textarea
                        id="travel"
                        v-model="form.travel"
                        :maxlength="LIMITS.aboutField"
                        placeholder="Places you've been"
                        @input="validateField('travel')"
                    ></textarea>

                    <p v-if="errors.travel" class="error">
                        {{ errors.travel }}
                    </p>
                </div>
            </div>

            <div class="section-heading links-heading">
                <p class="eyebrow">SOCIAL</p>
                <h2>Links</h2>
            </div>

            <div class="field-grid">
                <div class="form-field">
                    <label for="website">Website</label>
                    <textarea
                        id="website"
                        v-model="form.website"
                        rows="1"
                        inputmode="url"
                        autocomplete="url"
                        spellcheck="false"
                        @keydown.enter.prevent
                        :maxlength="LIMITS.aboutField"
                        placeholder="https://"
                        @input="validateField('website')"
                    ></textarea>

                    <p v-if="errors.website" class="error">
                        {{ errors.website }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="linkedin">LinkedIn</label>
                    <textarea
                        id="linkedin"
                        v-model="form.linkedin"
                        rows="1"
                        inputmode="url"
                        autocomplete="url"
                        spellcheck="false"
                        @keydown.enter.prevent
                        :maxlength="LIMITS.aboutField"
                        placeholder="https://linkedin.com/in/"
                        @input="validateField('linkedin')"
                    ></textarea>

                    <p v-if="errors.linkedin" class="error">
                        {{ errors.linkedin }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="twitter">Twitter / X</label>
                    <textarea
                        id="twitter"
                        v-model="form.twitter"
                        rows="1"
                        inputmode="url"
                        autocomplete="url"
                        spellcheck="false"
                        @keydown.enter.prevent
                        :maxlength="LIMITS.aboutField"
                        placeholder="https://x.com/"
                        @input="validateField('twitter')"
                    ></textarea>

                    <p v-if="errors.twitter" class="error">
                        {{ errors.twitter }}
                    </p>
                </div>

                <div class="form-field">
                    <label for="instagram">Instagram</label>
                    <textarea
                        id="instagram"
                        v-model="form.instagram"
                        rows="1"
                        inputmode="url"
                        autocomplete="url"
                        spellcheck="false"
                        @keydown.enter.prevent
                        :maxlength="LIMITS.aboutField"
                        placeholder="https://instagram.com/"
                        @input="validateField('instagram')"
                    ></textarea>

                    <p v-if="errors.instagram" class="error">
                        {{ errors.instagram }}
                    </p>
                </div>
            </div>

            <button
                type="button"
                class="confirm-button"
                @click="confirmChanges"
            >
                Confirm changes
            </button>
        </div>
    </section>
</template>

<style scoped>
.edit-section {
    scroll-margin-top: 100px;
    width: 100%;
    max-width: 100%;
}
.error {
    margin: 0;
    color: #d9534f;
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
}
.section-heading {
    margin-bottom: clamp(14px, 2.5vw, 20px);
}

.links-heading {
    margin-top: 5px;
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

.field-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
    gap: clamp(14px, 2.5vw, 20px);
}

.form-field {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
}

.form-field label {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: clamp(8px, 1.2vw, 9px);
    font-weight: 600;
    letter-spacing: 1px;
}

.form-field textarea {
    width: 100%;
    min-height: 90px;
    padding: clamp(10px, 1.8vw, 12px) clamp(10px, 2vw, 14px);
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: clamp(12px, 1.8vw, 13px);
    resize: vertical;
    box-sizing: border-box;
}

.form-field textarea:focus {
    outline: none;
    border-color: var(--input-focus);
    box-shadow: 3px 3px var(--input-focus);
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