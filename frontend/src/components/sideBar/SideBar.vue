<template>
    <div class="q-pa-md">
        <q-card class="surface q-pb-xs">
            <q-item :to="profilePath" class="q-pa-md">
                <q-item-section avatar>
                    <q-avatar size="52px">
                        <img v-if="GetUserData?.result?.imageUrl" :src="GetUserData?.result?.imageUrl" />
                        <img v-else :src="defaultAvatar" />
                    </q-avatar>
                </q-item-section>
                <q-item-section>
                    <q-item-label class="text-subtitle1 text-weight-bold">{{ GetUserData?.result?.name || 'Welcome' }}</q-item-label>
                    <q-item-label caption>{{ GetUserData?.result?.email || 'Sign in to get started' }}</q-item-label>
                </q-item-section>
            </q-item>

            <q-separator />

            <q-item clickable v-ripple to="/" :class="activeClass('/')">
                <q-item-section avatar><q-icon color="primary" name="eva-home-outline" /></q-item-section>
                <q-item-section>Home</q-item-section>
            </q-item>

            <q-item clickable v-ripple :to="profilePath" :class="activeClass('/profile')">
                <q-item-section avatar><q-icon color="primary" name="eva-person-outline" /></q-item-section>
                <q-item-section>Profile</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Chat" :class="activeClass('/Chat')">
                <q-item-section avatar><q-icon color="primary" name="eva-message-square-outline" /></q-item-section>
                <q-item-section>Messages</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Notification" :class="activeClass('/Notification')">
                <q-item-section avatar><q-icon color="primary" name="eva-bell-outline" /></q-item-section>
                <q-item-section>Notifications</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Auth" v-if="!GetUserData?.result">
                <q-item-section avatar><q-icon color="primary" name="eva-log-in-outline" /></q-item-section>
                <q-item-section>Sign In</q-item-section>
            </q-item>
        </q-card>
    </div>
</template>

<script>
import { mapGetters } from 'vuex'

const DEFAULT_AVATAR = 'https://cdn-icons-png.flaticon.com/512/3237/3237472.png';

export default {
    name:'SideBar',
    data() {
        return {
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    computed:{
        ...mapGetters('auth', ['GetUserData']),
        profilePath() {
            const id = this.GetUserData?.result?._id
            return id ? `/profile/${id}` : '/Auth'
        }
    },
    methods: {
        activeClass(path) {
            if (path === '/') return this.$route.path === '/' ? 'bg-primary-1' : ''
            return this.$route.path.startsWith(path) ? 'bg-primary-1' : ''
        }
    }
}
</script>
