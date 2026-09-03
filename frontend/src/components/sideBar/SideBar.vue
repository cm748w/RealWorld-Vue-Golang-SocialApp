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
                    <q-item-label caption v-if="!GetUserData?.result">Sign in to get started</q-item-label>
                </q-item-section>
            </q-item>

            <q-separator />

            <q-item clickable v-ripple to="/" :class="activeClass('/')">
                <q-item-section avatar><q-icon color="" name="eva-home-outline" /></q-item-section>
                <q-item-section>Home</q-item-section>
            </q-item>

            <q-item clickable v-ripple :to="profilePath" :class="activeClass('/profile')">
                <q-item-section avatar><q-icon color="" name="eva-person-outline" /></q-item-section>
                <q-item-section>Profile</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Chat" :class="activeClass('/Chat')">
                <q-item-section avatar><q-icon color="" name="eva-message-square-outline" /></q-item-section>
                <q-item-section>Messages</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Notification" :class="activeClass('/Notification')">
                <q-item-section avatar><q-icon color="" name="eva-bell-outline" /></q-item-section>
                <q-item-section>Notifications</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Auth" v-if="!GetUserData?.result">
                <q-item-section avatar><q-icon color="" name="eva-log-in-outline" /></q-item-section>
                <q-item-section>Sign In</q-item-section>
            </q-item>
        </q-card>
    </div>
</template>

<script>
import { mapGetters } from 'vuex'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

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
            if (path === '/') return this.$route.path === '/' ? 'nav-active' : ''
            return this.$route.path.startsWith(path) ? 'nav-active' : ''
        }
    }
}
</script>
