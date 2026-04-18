<template>
<div>
    <q-item class="RightBar"
        v-for="user in UsersData"
        :key="user._id"
        :to="`/profile/${user?._id}`" >

        <q-item-section avatar>
            <q-avatar>
                <img v-if="user.imageUrl" :src="user.imageUrl" >
                <img v-else src="https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center" >

            </q-avatar>
        </q-item-section>
        <q-item-section>
            <q-item-label class="text-bold">{{ user?.name }}</q-item-label>
            <q-item-label caption>
                {{ user?.bio }} 
            </q-item-label>
        </q-item-section>
    </q-item>
</div>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'
export default {
    name:'RightBar',
    data(){
        return {
            UsersData:[]
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
            console.log("Rightbar", users)
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

<style lang="sass" scoped>
.RightBar
    cursor: pointer
</style>



