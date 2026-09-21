import { createRouter, createWebHistory } from 'vue-router'
import Auth from '@/views/Auth.vue'
import PersonalProfile from '@/views/profiles/PersonalProfile.vue'
import EditProfile from '@/views/profiles/EditProfile.vue'
import Profile from '@/views/profiles/Profile.vue'
import AddPostPage from '@/views/posts/AddPostPage.vue'
import HomePage from '@/views/home/HomePage.vue'
import Notifications from '@/views/Notifications.vue'
import Chats from '@/views/chats/Chats.vue'
import GroupsPages from '@/views/groups/GroupsPages.vue'
import GroupChatPage from '@/views/groups/GroupChatPage.vue'
import SearchPage from '@/views/search/SearchPage.vue'
import Settings from '@/views/Settings.vue'

const routes = [
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
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes
})
