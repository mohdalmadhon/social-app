import { checkSessionResponse } from "@/helpers/auth/auth";
import { About, Users, Profile } from "@/models/users";
import { router } from "@/router/router";
import { sendWS } from "../socket/socket";

export async function getProfileData(id, count) {
    const resp = await fetch(`/api/profile?id=${id}`, {
        method: "GET",
        credentials: 'include',
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    const result = await resp.json();

    if (!resp.ok) {
        if (result.message == "cannot view your own profile this way") {
            router.push("/me")
            return;
        }
        throw new Error("network error, could not connet to server")
    }

    const data = result.data;
    const userInfo = data.UserInfo;

    const about = new About(
        data.About?.Bio,
        data.About?.Work,
        data.About?.Education,
        data.About?.Travel,
        data.About?.interests,
        data.About?.Hobbies,
        data.About?.Website,
        data.About?.Linkedin,
        data.About?.instagram,
        data.About?.Twitter
    );

    const profile = new Profile(
        data.NumOfFollowing,
        data.NumOfFollowers,
        data.NumOfPosts,
        userInfo.avatar,
        about,
        data.Following,
        data.Followers,
        data.Friends
    );

    profile.show = result.showProfile;

    const user = new Users(
        userInfo.ID,
        userInfo.firstName,
        userInfo.lastName,
        userInfo.username,
        userInfo.email,
        userInfo.DOB,
        userInfo.isPrivate,
        profile
    );

    user.show = result.showProfile;
    user.isFollowing = result.followStatus;

    if (result.showProfile) {
        try {
            const followersResult = await getFollowers(id, count, 0);
            if (!followersResult.status) {
                throw new Error("could not get user data");
            }
            user.Profile.followers = followersResult.data;
        } catch (err) {
            throw new Error("network error, could not connet to server");
        }

        try {
            const followingResult = await getFollowing(id, count, 0);
            if (!followingResult.status) {
                throw new Error("could not get user data");
            }
            user.Profile.following = followingResult.data;
        } catch (err) {
            throw new Error("network error, could not connet to server");
        }
    }

    return user;
}

export async function requestFollow(id, method) {
    const resp = await fetch(`/api/profile/follow?targetid=${id}`, {
        method: method,
        credentials: 'include'
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    if (!resp.ok) {
        throw new Error("network error, could not connet to server")
    }

    const result = await resp.json();
    return result
}

export async function getFollowers(id, count, offset = 0) {
    const resp = await fetch(`/api/profile/follow?targetid=${id}&count=${count}&offset=${offset}`, {
        method: "GET",
        credentials: 'include'
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    if (!resp.ok) {
        throw new Error('could not connect to network')
    }

    const result = await resp.json();
    return result;
}

export async function removeFollower(id) {
    const resp = await fetch(`/api/profile/followers?targetid=${id}`, {
        method: "DELETE",
        credentials: 'include'
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    const result = await resp.json();

    if (!resp.ok) {
        throw new Error(result.message || 'could not connect to network')
    }

    return result;
}

export async function getFollowing(id, count, offset = 0) {
    const resp = await fetch(`/api/profile/following?targetid=${id}&count=${count}&offset=${offset}`, {
        method: "GET",
        credentials: 'include'
    });

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return;
    }

    if (!resp.ok) {
        throw new Error('could not connect to network')
    }

    const result = await resp.json();
    return result;
}

export async function searchFollows(searchValue = "", targetId) {
    const params = new URLSearchParams({
        search: searchValue,
        targetid: targetId
    });

    const resp = await fetch(`/api/profile/follows/search?${params.toString()}`, {
        method: "GET",
        credentials: 'include'
    });

    if (!resp.ok) {
        throw new Error("could not connect to server")
    }

    const result = await resp.json();
    return result;
}

export async function searchFollowing(searchValue = "", targetId) {
    const params = new URLSearchParams({
        search: searchValue,
        targetid: targetId
    });
    
    const resp = await fetch(`/api/profile/following/search?${params.toString()}`, {
        method: "GET",
        credentials: 'include'
    });

    if (!resp.ok) {
        throw new Error("could not connect to server")
    }

    const result = await resp.json();
    return result;
}
