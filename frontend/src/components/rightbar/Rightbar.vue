<template>
<div class="q-pa-md">
    <q-card class="surface q-pb-xs">
        <div class="section-title q-pa-md q-pb-sm">Who to follow</div>
        <q-separator />

        <template v-if="UsersData.length">
            <q-item
                v-for="user in UsersData"
                :key="user._id"
                :to="`/profile/${user?._id}`"
                clickable v-ripple>

                <q-item-section avatar>
                    <q-avatar>
                        <img v-if="user.imageUrl" :src="user.imageUrl" >
                        <img v-else :src="defaultAvatar" >
                    </q-avatar>
                </q-item-section>
                <q-item-section>
                    <q-item-label class="text-weight-bold">{{ user?.name }}</q-item-label>
                    <q-item-label caption class="ellipsis">{{ user?.bio }}</q-item-label>
                </q-item-section>
            </q-item>
        </template>
        <q-item v-else class="text-center muted">
            <q-item-section class="q-pa-lg">No suggestions yet</q-item-section>
        </q-item>
    </q-card>
</div>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    name:'RightBar',
    data(){
        return {
            UsersData:[],
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    computed:{
        ...mapGetters('auth', ['GetUserData']),
        currentUserId() {
            return this.GetUserData?.result?._id || null
        },
    },
    methods:{
        ...mapActions('users', ['GetRecommendUsers']),
        async loadRecommendations(userId) {
            if (!userId) {
                this.UsersData = []
                return
            }
            const response = await this.GetRecommendUsers(userId)
            const users = Array.isArray(response?.users) ? response.users : []
            this.UsersData = users
        }
    },
    watch: {
        currentUserId: {
            immediate: true,
            handler(userId) {
                this.loadRecommendations(userId)
            },
        },
    },
    async mounted(){
        // 数据由 watcher 统一驱动，避免挂载时机早于 auth state 初始化
    }
}
</script>
