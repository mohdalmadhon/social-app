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
