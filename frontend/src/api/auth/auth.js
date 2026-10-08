import { checkSessionResponse } from "@/helpers/auth/auth";
import { router } from "@/router/router";

async function parseResponse(resp, fallback) {
    let result = {}

    try {
        result = await resp.json()
    } catch {
        result = {}
    }

    if (!resp.ok) {
        const error = new Error(result.message || `${fallback}: ${resp.status}`)

        error.status = resp.status
        error.retryAfter = result.retryAfter
        error.attemptsLeft = result.attemptsLeft

        throw error
    }

    return result
}

export async function registerUser(userData) {
    const resp = await fetch("/api/user", {
        method: "POST",
        body: userData
    })

    return parseResponse(resp, "Registration failed")
}

export async function checkRegistration(type, value) {
    const params = new URLSearchParams({ type, value })

    const resp = await fetch(`/api/registration/check?${params}`)

    return parseResponse(resp, "Check failed")
}

export async function sendEmailCode(email) {
    const resp = await fetch("/api/email/code", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email })
    })

    return parseResponse(resp, "Could not send code")
}

export async function verifyEmailCode(email, code) {
    const resp = await fetch("/api/email/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, code })
    })

    return parseResponse(resp, "Verification failed")
}

export async function loggingSession(userLogger) {
    const resp = await fetch("/api/session", {
        method: "POST",
        credentials: 'include',
        body: JSON.stringify(userLogger)
    });

    const result = await resp.json()

    if (!resp.ok) {
        throw new Error(result.message || `Logging failed: ${resp.status}`)
    }

    return result
}

export async function logout() {
    const resp = await fetch("/api/session", {
        method: "DELETE",
        credentials: 'include'
    });

    if(!checkSessionResponse(resp)) {
        router.push("/login");
    }

    if (!resp.ok && resp.status != 401) {
        throw new Error("could not logout")

    }

    router.push("/login")
}

export async function authorizeSession() {
    const resp = await fetch("/api/session", {
        method: "GET",
        credentials: "include"
    });

    const result = await resp.json();

    return result;
}