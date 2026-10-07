<script setup>
import { onBeforeUnmount, pushScopeId, reactive, ref } from 'vue'

import {
    loggingSession,
    registerUser,
    checkRegistration,
    sendEmailCode,
    verifyEmailCode,
} from '@/api/auth/auth.js'

import {
    validateName,
    validateUsername,
    validateEmail,
    validateDOB,
    validatePassword,
    validateAbout,
    handleEmailInput,
    handleUsernameInput,
    handleNameInput,
    validateAvatar,
} from '@/helpers/validators/registration.js'

import InputHolder from './InputHolder.vue'
import { router } from '@/router/router.js'
import { addNotification } from '@/data/notifications.js'

const toggleInput = ref(null)
const signupForm = ref(null)

const loginErrors = reactive({
    identifier: '',
    password: '',
    form: '',
})

const errors = reactive({
    firstName: '',
    lastName: '',
    username: '',
    dob: '',
    email: '',
    about: '',
    password: '',
    avatar: '',
})

const touched = reactive({
    firstName: false,
    lastName: false,
    username: false,
    dob: false,
    email: false,
    about: false,
    password: false,
    avatar: false,
})

const availability = reactive({
    username: null,
    email: null,
})

const CODE_LENGTH = 6
const MAX_ATTEMPTS = 3

const step = ref('form')
const pendingEmail = ref('')
const code = ref('')
const codeError = ref('')
const attemptsLeft = ref(MAX_ATTEMPTS)
const resendIn = ref(0)
const sendingCode = ref(false)
const verifying = ref(false)
const registering = ref(false)

let pendingForm = null
let verifiedEmail = ''
let verifyToken = ''
let resendTimer = null

function stopResendTimer() {
    if (resendTimer) {
        clearInterval(resendTimer)
        resendTimer = null
    }
}

function startResendTimer(seconds) {
    stopResendTimer()

    resendIn.value = seconds

    if (seconds <= 0) return

    resendTimer = setInterval(() => {
        resendIn.value -= 1

        if (resendIn.value <= 0) {
            resendIn.value = 0
            stopResendTimer()
        }
    }, 1000)
}

onBeforeUnmount(stopResendTimer)

function resetVerification() {
    stopResendTimer()

    step.value = 'form'
    pendingEmail.value = ''
    code.value = ''
    codeError.value = ''
    attemptsLeft.value = MAX_ATTEMPTS
    resendIn.value = 0

    pendingForm = null
    verifiedEmail = ''
    verifyToken = ''
}

function resetSignupForm() {
    signupForm.value?.reset()

    Object.keys(errors).forEach(key => {
        errors[key] = ''
        touched[key] = false
    })

    availability.username = null
    availability.email = null
}

function validateField(field, value) {
    touched[field] = true

    const validators = {
        avatar: validateAvatar,
        firstName: validateName,
        lastName: validateName,
        username: validateUsername,
        email: validateEmail,
        dob: validateDOB,
        password: validatePassword,
        about: validateAbout,
    }

    errors[field] = validators[field](value)

    if (field === 'username' && errors.username) {
        availability.username = null
    }

    if (field === 'email' && errors.email) {
        availability.email = null
    }
}

function handleAvatar(event) {
    validateField('avatar', event.target.files?.[0] || null)
}

function handleName(field, event) {
    const value = handleNameInput(event.target.value)

    event.target.value = value
    validateField(field, value)
}

function handleUsername(event) {
    const value = handleUsernameInput(event.target.value)

    event.target.value = value
    validateField('username', value)

    if (!value) {
        availability.username = null
    }
}

function handleEmail(event) {
    const value = handleEmailInput(event.target.value)

    event.target.value = value
    validateField('email', value)

    if (!value) {
        availability.email = null
    }
}

async function checkAvailability(field, value) {
    if (!value || errors[field]) {
        availability[field] = null
        return
    }

    try {
        const type = field === 'username' ? 'name' : 'email'

        const result = await checkRegistration(type, value)

        availability[field] = result.avilable
    } catch {
        availability[field] = null
    }
}

async function checkUsernameAvailability(event) {
    const value = event.target.value.trim()

    validateField('username', value)

    if (errors.username) {
        availability.username = null
        return
    }

    await checkAvailability('username', value)
}

async function checkEmailAvailability(event) {
    const value = event.target.value.trim()

    validateField('email', value)

    if (errors.email) {
        availability.email = null
        return
    }

    await checkAvailability('email', value)
}

function inputClass(field) {
    if (!touched[field]) return ''

    if (errors[field]) return 'input-error'

    if (availability[field] === false) return 'input-error'

    return 'input-valid'
}

function validateForm(form) {
    const data = new FormData(form)

    validateField('firstName', data.get('FirstName') || '')
    validateField('lastName', data.get('LastName') || '')
    validateField('username', data.get('UserName') || '')
    validateField('dob', data.get('dob') || '')
    validateField('email', data.get('Email') || '')
    validateField('about', data.get('About') || '')
    validateField('password', data.get('Password') || '')
    validateField('avatar', data.get('Avatar'))

    return (
        Object.values(errors).every(error => !error) &&
        availability.email !== false &&
        (
            !data.get('UserName') ||
            availability.username !== false
        )
    )
}

async function requestCode(email) {
    sendingCode.value = true
    codeError.value = ''

    try {
        const result = await sendEmailCode(email)

        code.value = ''
        attemptsLeft.value = result.attempts ?? MAX_ATTEMPTS
        startResendTimer(result.cooldown ?? 60)
        step.value = 'code'
    } catch (error) {
        if (error.status === 409) {
            touched.email = true
            availability.email = false
            step.value = 'form'
            addNotification(error.message, 'error')
            return
        }

        if (error.status === 429) {
            startResendTimer(error.retryAfter ?? 60)
            step.value = 'code'
            codeError.value = error.message
            return
        }

        addNotification(`Could not send code: ${error.message}`, 'error')
    } finally {
        sendingCode.value = false
    }
}

async function finishRegistration() {
    if (!pendingForm) return

    registering.value = true
    pendingForm.set('VerifyToken', verifyToken)

    try {
        const result = await registerUser(pendingForm)

        if (!result.status) {
            addNotification(`Failed to register: ${result.message}`, 'error')
            step.value = 'form'
            return
        }

        addNotification('Account created successfully.', 'success');
        
        window.location.replace("/");
        return;

    } catch (error) {
        if (error.status === 403) {
            verifiedEmail = ''
            verifyToken = ''
            code.value = ''
            attemptsLeft.value = 0
            codeError.value = 'Your email verification expired. Request a new code.'
            step.value = 'code'
        } else {
            step.value = 'form'
        }

        addNotification(`Failed to register: ${error.message}`, 'error')
    } finally {
        registering.value = false
    }
}

async function sendData(event) {
    event.preventDefault()

    if (sendingCode.value || registering.value) return

    const form = event.target

    if (!validateForm(form)) return

    pendingForm = new FormData(form)

    const email = String(pendingForm.get('Email') || '').trim().toLowerCase()

    pendingEmail.value = email

    if (verifyToken && verifiedEmail === email) {
        await finishRegistration()
        return
    }

    await requestCode(email)
}

function handleCode(event) {
    const value = event.target.value.replace(/\D/g, '').slice(0, CODE_LENGTH)

    event.target.value = value
    code.value = value
    codeError.value = ''
}

async function verifyCode() {
    if (verifying.value || registering.value || attemptsLeft.value <= 0) return

    if (code.value.length !== CODE_LENGTH) {
        codeError.value = `Enter the ${CODE_LENGTH}-digit code`
        return
    }

    verifying.value = true
    codeError.value = ''

    try {
        const result = await verifyEmailCode(pendingEmail.value, code.value)

        verifiedEmail = pendingEmail.value
        verifyToken = result.token

        await finishRegistration()
    } catch (error) {
        if (typeof error.attemptsLeft === 'number') {
            attemptsLeft.value = error.attemptsLeft
        }

        code.value = ''
        codeError.value = error.message
    } finally {
        verifying.value = false
    }
}

async function resendCode() {
    if (sendingCode.value || resendIn.value > 0) return

    await requestCode(pendingEmail.value)
}

function backToForm() {
    step.value = 'form'
    code.value = ''
    codeError.value = ''
}

async function loggUser(event) {
    event.preventDefault()

    loginErrors.identifier = ''
    loginErrors.password = ''
    loginErrors.form = ''

    const formData = new FormData(event.target)

    const identifier = formData.get('Identifier')
    const password = formData.get('Pass')

    if (!identifier) {
        loginErrors.identifier = 'Email or username is required'
    }

    if (!password) {
        loginErrors.password = 'Password is required'
    }

    if (loginErrors.identifier || loginErrors.password) return

    try {
        const result = await loggingSession({
            Identifier: identifier,
            Pass: password,
        })

        if (result.status) router.replace('/home')
    } catch (error) {
        loginErrors.form =
            error.message || 'Invalid email, username, or password'
    }
}
</script>

<template>
    <section class="auth-section">
        <div class="wrapper">
            <div class="card-switch">
                <label class="switch">
                    <input ref="toggleInput" type="checkbox" class="toggle">

                    <span class="slider"></span>

                    <span class="card-side"></span>

                    <div class="flip-card__inner">
                        <div class="flip-card__front">
                            <div class="title">
                                Log in
                            </div>

                            <p class="form-subtitle">
                                Welcome back.
                            </p>

                            <form @submit.prevent="loggUser" class="flip-card__form">
                                <div class="input-group">
                                    <label for="login-email">
                                        Email or username
                                    </label>

                                    <InputHolder id="login-email" type="text" name="Identifier" :minLength="3"
                                        :maxLength="75" placeHolder="email@example.com" autocomplete="username"
                                        required />

                                    <span v-if="loginErrors.identifier" class="input-error-message">
                                        {{ loginErrors.identifier }}
                                    </span>
                                </div>

                                <div class="input-group">
                                    <label for="login-password">
                                        Password
                                    </label>

                                    <InputHolder id="login-password" type="password" name="Pass" :minLength="8"
                                        :maxLength="100" placeHolder="Password" autocomplete="current-password"
                                        required />

                                    <span v-if="loginErrors.password" class="input-error-message">
                                        {{ loginErrors.password }}
                                    </span>
                                </div>

                                <span v-if="loginErrors.form" class="input-error-message login-error" role="alert">
                                    {{ loginErrors.form }}
                                </span>

                                <button class="flip-card__btn" type="submit">
                                    Let's go!
                                </button>
                            </form>

                            <p class="form-footer">
                                Don't have an account?
                                <span>Sign up</span>
                            </p>
                        </div>

                        <div class="flip-card__back">
                            <div class="title">
                                Sign up
                            </div>

                            <p class="form-subtitle">
                                Create your account.
                            </p>

                            <form ref="signupForm" class="flip-card__form" @submit.prevent="sendData">
                                <div class="input-row">
                                    <div class="input-group">
                                        <label for="first-name">
                                            First name
                                        </label>

                                        <InputHolder id="first-name" type="text" name="FirstName" :minLength="2"
                                            :maxLength="15" placeHolder="First name" :class="inputClass('firstName')"
                                            autocomplete="given-name" @input="handleName('firstName', $event)"
                                            required />

                                        <span v-if="touched.firstName && errors.firstName" class="input-error-message">
                                            {{ errors.firstName }}
                                        </span>
                                    </div>

                                    <div class="input-group">
                                        <label for="last-name">
                                            Last name
                                        </label>

                                        <InputHolder id="last-name" type="text" name="LastName" :minLength="2"
                                            :maxLength="15" placeHolder="Last name" :class="inputClass('lastName')"
                                            autocomplete="family-name" @input="handleName('lastName', $event)"
                                            required />

                                        <span v-if="touched.lastName && errors.lastName" class="input-error-message">
                                            {{ errors.lastName }}
                                        </span>
                                    </div>
                                </div>

                                <div class="input-row">
                                    <div class="input-group">
                                        <label for="username">
                                            Username
                                        </label>

                                        <InputHolder id="username" type="text" name="UserName" :minLength="3"
                                            :maxLength="12" placeHolder="Username" :class="inputClass('username')"
                                            autocomplete="nickname" @input="handleUsername"
                                            @blur="checkUsernameAvailability" />

                                        <span v-if="touched.username && errors.username" class="input-error-message">
                                            {{ errors.username }}
                                        </span>

                                        <span v-else-if="availability.username === false" class="input-error-message">
                                            Username is already taken
                                        </span>
                                    </div>

                                    <div class="input-group">
                                        <label for="dob">
                                            Date of birth
                                        </label>

                                        <InputHolder id="dob" type="date" name="dob" :class="inputClass('dob')"
                                            @input="validateField('dob', $event.target.value)" required />

                                        <span v-if="touched.dob && errors.dob" class="input-error-message">
                                            {{ errors.dob }}
                                        </span>
                                    </div>
                                </div>

                                <div class="input-group">
                                    <label for="signup-email">
                                        Email
                                    </label>

                                    <InputHolder id="signup-email" type="email" name="Email" :minLength="5"
                                        :maxLength="75" placeHolder="email@example.com" :class="inputClass('email')"
                                        autocomplete="email" @input="handleEmail" @blur="checkEmailAvailability"
                                        required />

                                    <span v-if="touched.email && errors.email" class="input-error-message">
                                        {{ errors.email }}
                                    </span>

                                    <span v-else-if="availability.email === false" class="input-error-message">
                                        Email is already in use
                                    </span>
                                </div>

                                <div class="input-group">
                                    <label for="avatar">
                                        Avatar
                                    </label>

                                    <InputHolder id="avatar" type="file" name="Avatar"
                                        accept="image/jpeg,image/png,image/gif" :class="inputClass('avatar')"
                                        @change="handleAvatar" />

                                    <span v-if="touched.avatar && errors.avatar" class="input-error-message">
                                        {{ errors.avatar }}
                                    </span>
                                </div>

                                <div class="input-group">
                                    <label for="about">
                                        About
                                    </label>

                                    <textarea id="about" name="About" maxlength="1000"
                                        placeholder="Tell us a little about yourself..." :class="inputClass('about')"
                                        @input="validateField('about', $event.target.value)"></textarea>

                                    <span v-if="touched.about && errors.about" class="input-error-message">
                                        {{ errors.about }}
                                    </span>
                                </div>

                                <div class="input-group">
                                    <label for="signup-password">
                                        Password
                                    </label>

                                    <InputHolder id="signup-password" type="password" name="Password" :minLength="8"
                                        :maxLength="75" placeHolder="Password" :class="inputClass('password')"
                                        autocomplete="new-password"
                                        @input="validateField('password', $event.target.value)" required />

                                    <span v-if="touched.password && errors.password" class="input-error-message">
                                        {{ errors.password }}
                                    </span>
                                </div>

                                <button class="flip-card__btn" type="submit" :disabled="sendingCode || registering">
                                    {{ sendingCode ? 'Sending...' : 'Confirm!' }}
                                </button>
                            </form>

                            <p class="form-footer">
                                Already have an account?
                                <span>Log in</span>
                            </p>
                        </div>
                    </div>
                </label>
            </div>
        </div>

        <Teleport to="body">
            <div v-if="step === 'code'" class="confirm-overlay" role="dialog" aria-modal="true"
                aria-labelledby="confirm-title">
                <form class="confirm-dialog" @submit.prevent="verifyCode">
                    <h2 id="confirm-title" class="confirm-title">
                        Verify your email
                    </h2>

                    <p class="confirm-text">
                        We sent a 6-digit code to
                        <strong>{{ pendingEmail }}</strong>.
                        It expires in 10 minutes.
                    </p>

                    <div class="input-group">
                        <label for="verify-code">
                            Verification code
                        </label>

                        <InputHolder id="verify-code" type="text" name="Code" :maxLength="6" placeHolder="123456"
                            class="code-input" inputmode="numeric" autocomplete="one-time-code" :value="code"
                            :disabled="attemptsLeft <= 0 || verifying || registering" @input="handleCode" />

                        <span v-if="codeError" class="input-error-message" role="alert">
                            {{ codeError }}
                        </span>

                        <span v-else-if="attemptsLeft > 0" class="confirm-hint">
                            {{ attemptsLeft }} {{ attemptsLeft === 1 ? 'try' : 'tries' }} left
                        </span>
                    </div>

                    <button class="flip-card__btn confirm-btn" type="submit"
                        :disabled="attemptsLeft <= 0 || verifying || registering || code.length !== CODE_LENGTH">
                        {{ verifying || registering ? 'Verifying...' : 'Verify' }}
                    </button>

                    <div class="confirm-actions">
                        <button class="link-button" type="button" :disabled="sendingCode || resendIn > 0"
                            @click="resendCode">
                            {{
                                sendingCode
                                    ? 'Sending...'
                                    : resendIn > 0
                                        ? `Resend code in ${resendIn}s`
                                        : 'Resend code'
                            }}
                        </button>

                        <button class="link-button" type="button" :disabled="verifying || registering"
                            @click="backToForm">
                            Change email
                        </button>
                    </div>
                </form>
            </div>
        </Teleport>
    </section>
</template>

<style scoped>
.login-error {
    text-align: center;
    margin-top: -5px;
}

.auth-section {
    min-height: 100vh;
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 60px;
    background: var(--page-background);
}

.wrapper {
    width: 100%;
    max-width: 520px;
}

.card-switch {
    width: 100%;
}

.switch {
    position: relative;
    width: 100%;
    min-height: 700px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
}

.toggle {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
}

.card-side {
    position: absolute;
    top: 0;
    width: 60px;
    height: 22px;
}

.card-side::before {
    position: absolute;
    content: "Log in";
    left: -90px;
    top: 0;
    width: 100px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
    text-decoration: underline;
}

.card-side::after {
    position: absolute;
    content: "Sign up";
    left: 75px;
    top: 0;
    width: 100px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
    text-decoration: none;
}

.toggle:checked~.card-side::before {
    text-decoration: none;
}

.toggle:checked~.card-side::after {
    text-decoration: underline;
}

.slider {
    position: absolute;
    top: 0;
    left: 50%;
    width: 50px;
    height: 22px;
    transform: translateX(-50%);
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    cursor: pointer;
    transition: 0.3s;
}

.slider::before {
    position: absolute;
    content: "";
    width: 20px;
    height: 20px;
    left: -2px;
    bottom: 2px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 0 3px 0 var(--main-color);
    transition: 0.3s;
}

.toggle:checked+.slider {
    background: var(--input-focus);
}

.toggle:checked+.slider::before {
    transform: translateX(30px);
}

.flip-card__inner {
    position: relative;
    width: 480px;
    height: 650px;
    margin-top: 45px;
    perspective: 1000px;
    transition: transform 0.8s;
    transform-style: preserve-3d;
}

.toggle:checked~.flip-card__inner {
    transform: rotateY(180deg);
}

.flip-card__front,
.flip-card__back {
    position: absolute;
    width: 100%;
    height: 100%;
    padding: 35px 40px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background: var(--bg-color);
    border: 2px solid var(--main-color);
    border-radius: 8px;
    box-shadow: 7px 7px var(--main-color);
    backface-visibility: hidden;
    -webkit-backface-visibility: hidden;
}

.flip-card__back {
    transform: rotateY(180deg);
    overflow-y: auto;
}

.title {
    margin: 0 0 2px;
    color: var(--main-color);
    font-family: "Liter", serif;
    font-size: 32px;
    font-weight: 900;
}

.form-subtitle {
    margin: 0 0 20px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.flip-card__form {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.input-row {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 15px;
}

.input-group {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 7px;
}

.input-group label {
    color: var(--font-color, #323232);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
}

.input-group :deep(.form-input) {
    width: 100%;
}

.input-group textarea {
    width: 100%;
    min-height: 75px;
    resize: vertical;
    padding: 10px 12px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    outline: none;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 500;
}

.input-group textarea::placeholder {
    color: var(--font-color-sub);
    opacity: 0.7;
}

.input-group textarea:focus {
    border-color: var(--input-focus);
    box-shadow: 4px 4px var(--input-focus);
}

.input-group :deep(.input-error),
.input-group textarea.input-error {
    border-color: #e74c3c !important;
    box-shadow: 4px 4px #e74c3c !important;
}

.input-group :deep(.input-valid),
.input-group textarea.input-valid {
    border-color: #2ecc71 !important;
    box-shadow: 4px 4px #2ecc71 !important;
}

.input-error-message {
    color: #e74c3c;
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    line-height: 1.3;
}

.flip-card__btn {
    align-self: center;
    width: 140px;
    height: 45px;
    margin-top: 8px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: 0.15s;
}

.flip-card__btn:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 6px 6px var(--main-color);
}

.flip-card__btn:active:not(:disabled) {
    transform: translate(4px, 4px);
    box-shadow: 0 0 var(--main-color);
}

.flip-card__btn:disabled {
    cursor: not-allowed;
    opacity: 0.55;
}

.form-footer {
    margin: 18px 0 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.form-footer span {
    color: var(--input-focus);
    font-weight: 600;
}

.confirm-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgb(0 0 0 / 55%);
}

.confirm-dialog {
    width: 100%;
    max-width: 400px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 30px 32px;
    border: 2px solid var(--main-color, #323232);
    border-radius: 8px;
    background: var(--bg-color, #fff);
    box-shadow: 7px 7px var(--main-color, #323232);
}

.confirm-title {
    margin: 0;
    color: var(--main-color, #323232);
    font-family: "Liter", serif;
    font-size: 24px;
    font-weight: 900;
}

.confirm-text {
    margin: 0;
    color: var(--font-color-sub, #666);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    line-height: 1.5;
}

.confirm-text strong {
    color: var(--font-color, #323232);
    overflow-wrap: anywhere;
}

.confirm-hint {
    color: var(--font-color-sub, #666);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.code-input {
    font-family: "JetBrains Mono", monospace;
    font-size: 20px;
    letter-spacing: 0.5em;
    text-align: center;
}

.confirm-btn {
    width: 100%;
}

.confirm-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}

.link-button {
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--input-focus, #2d8cf0);
    cursor: pointer;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    text-decoration: underline;
}

.link-button:disabled {
    color: var(--font-color-sub, #666);
    cursor: not-allowed;
    text-decoration: none;
}

@media (max-width: 900px) {
    .auth-section {
        min-height: 600px;
        padding: 60px 30px;
    }

    .flip-card__inner {
        width: 440px;
    }
}

@media (max-width: 550px) {
    .auth-section {
        padding: 50px 20px;
    }

    .flip-card__inner {
        width: min(400px, 90vw);
    }

    .flip-card__front,
    .flip-card__back {
        padding: 30px 25px;
    }

    .input-row {
        grid-template-columns: 1fr;
        gap: 15px;
    }

    .confirm-dialog {
        padding: 24px 20px;
    }
}
</style>