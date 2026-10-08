import { Users, Profile, About } from "@/models/users";
import { checkSessionResponse } from "@/helpers/auth/auth";
import { router } from "@/router/router";

export async function getUserData() {
    const resp = await fetch("/api/user", {
        method: "GET",
        credentials: "include"
    });

    const result = await resp.json();

    if (!checkSessionResponse(resp)) {
        router.replace("/login");
        return null;
    }

    if (!resp.ok) {
        throw new Error(result.message || `Failed to get user data: ${resp.status}`);
    }

    const data = result.data;
    const user = data.UserInfo;

    const about = new About(
        data.About.Bio,
        data.About.Work,
        data.About.Education,
        data.About.Travel,
        data.About.interests,
        data.About.Hobbies,
        data.About.Website,
        data.About.Linkedin,
        data.About.instagram,
        data.About.Twitter
    );

    const profile = new Profile(
        data.NumOfFollowing,
        data.NumOfFollowers,
        data.NumOfPosts,
        user.avatar,
        about,
        data.Following,
        data.Followers,
        data.Friends
    );

    const userData = new Users(
        user.ID,
        user.firstName,
        user.lastName,
        user.username,
        user.email,
        user.DOB,
        user.isPrivate,
        profile
    );
    console.log(userData)
    return userData;

}