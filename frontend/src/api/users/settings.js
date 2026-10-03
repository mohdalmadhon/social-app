import { checkSessionResponse } from "@/helpers/auth/auth";
import { router } from "@/router/router";

export async function deleteAccount() {
    const resp = await fetch("/api/user", {
        method: "DELETE",
        credentials: "include"
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || `Could not delete account: ${resp.status}`);
    }

    router.replace("/login");

    return result;
}

export async function changePreferences(type, preferences) {
    const resp = await fetch(`/api/user/preferences?type=${type}&value=${preferences}`, {
        method: "PATCH",
        credentials: 'include'
    })

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not update')
    }

    return result;
}

export async function getNotificationPreferences() {
    const resp = await fetch("/api/user/notification-preferences", {
        method: "GET",
        credentials: 'include'
    })

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not get notification preferences')
    }

    return result.data;
}

export async function changeNotificationPreference(type, mode) {
    const resp = await fetch(`/api/user/notification-preferences?type=${type}&value=${mode}`, {
        method: "PATCH",
        credentials: 'include'
    })

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not update')
    }

    return result;
}

export async function getPreferences() {
    const resp = await fetch("/api/user/preferences", {
        method: "GET",
        credentials: 'include'
    })

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not get preferences')
    }

    return result.data;
}
