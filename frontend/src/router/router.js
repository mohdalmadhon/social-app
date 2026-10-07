import { createRouter, createWebHistory } from 'vue-router'
import Auth from '@/views/auth/Auth.vue'
import PersonalProfile from '@/views/profiles/PersonalProfile.vue'
import EditProfile from '@/views/profiles/EditProfile.vue'
import Profile from '@/views/profiles/Profile.vue'
import AddPostPage from '@/views/posts/AddPostPage.vue'
import HomePage from '@/views/home/HomePage.vue'
import Notifications from '@/views/notifications/Notifications.vue'
import Chats from '@/views/chats/Chats.vue'
import GroupsPages from '@/views/groups/GroupsPages.vue'
import GroupChatPage from '@/views/groups/GroupChatPage.vue'
import SearchPage from '@/views/search/SearchPage.vue'
import Settings from '@/views/settings/Settings.vue'
import ErrorPage from '@/views/errors/error_page.vue'
import Welcome from '@/views/welcome/Welcome.vue'
import { setSessionUserId } from '@/data/currentUser'

const routes = [
    {
        path: '/',
        component: Welcome
    },
    {
        path: '/login',
        component: Auth
    },
    {
        path: '/me',
        component: PersonalProfile
    },
    {
        path: '/me/edit',
        component: EditProfile
    },
    {
        path: '/user',
        component: Profile
    }, 
    {
        path: '/post/new',
        component: AddPostPage
    },
    {
        path: '/home',
        component: HomePage
    },
    {
        path: `/notifications`,
        component: Notifications
    },
    {
        path: '/chats',
        component: Chats
    },
    {
        path: '/groups',
        component: GroupsPages
    },
    {
        path: "/groups/:id",
        component: GroupChatPage
    },
    {
        path: '/search',
        component: SearchPage
    },
    {
        path: '/settings',
        component: Settings
    },
    {
        path: '/error',
        component: ErrorPage
    },
    {
        path: '/:pathMatch(.*)*',
        component: ErrorPage,
        props: { code: 404 }
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes
})

const PUBLIC_PATHS = ['/', '/login', '/error']

async function hasValidSession() {
    try {
        const resp = await fetch('/api/session', {
            method: 'GET',
            credentials: 'include'
        })

        if (!resp.ok) {
            setSessionUserId(null)
            return false
        }

        const result = await resp.json()

        setSessionUserId(result.userID)

        return true
    } catch {
        setSessionUserId(null)
        return false
    }
}

router.beforeEach(async (to) => {
    if (PUBLIC_PATHS.includes(to.path)) {
        return true
    }

    if (await hasValidSession()) {
        return true
    }

    return { path: '/login', replace: true }
})
