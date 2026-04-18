<template>
    <div class="row col-12 constrain">
        <div class="col-4 text-center">
            <q-avatar size="150px">
                <img v-if="Luserdata?.imageUrl" :src="Luserdata?.imageUrl">
                <img v-else src="https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center">
            </q-avatar>
        </div>
        <!-- ------------------------------------------------------------------------------- -->
        <div class="col-8 text-left">
            <div class="text-h6 q-pa-lg" style="margin: auto;">
                {{ Luserdata?.name }}
                <q-btn v-if="isSameUser" @click="Edit" flat label="Edit" />

                <q-btn v-if="!isSameUser && !isUserFollowing" 
                    @click="FollowOrUnFollow" flat style="color: #FF0080;" label="Follow" />

                <q-btn v-if="!isSameUser && isUserFollowing"
                    @click="FollowOrUnFollow" flat class="primary" label="UN Follow" />
            </div>
            <q-separator inset />
            <div class="text-subtitle1 q-pa-lg" style="margin: auto;">
                {{ Luserdata?.bio }}
                <div>
                    <i>{{ userPosts?.length || 0 }} posts</i>
                    <i> 
                        <i v-if="Luserdata?.followers?.length > 0">
                            {{ Luserdata?.followers?.length }}</i>
                            followers
                    </i>
                    <i>
                        <i v-if="Luserdata?.following?.length > 0">
                            {{ Luserdata?.following?.length }}</i>
                            following
                    </i>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { mapActions } from 'vuex'
import { Notify } from 'quasar'

export default {
    props: ['userData', 'userPosts', 'isSameUser'],
    data() {
        return {
            isUserFollowing: false,
            Luserdata: {}
        }
    },
    watch: {
        userData: {
            handler(newVal) {
                if (newVal) {
                    this.Luserdata = { ...newVal }
                    // 当 userData 变化时，重新检查关注状态
                    this.checkUserFollowing()
                }
            },
            immediate: true
        }
    },
    methods: {
        ...mapActions('users', ['FollowUser']),
        syncFollowingState() {
            if (this.isSameUser) {
                this.isUserFollowing = false
                return
            }

            const profile = localStorage.getItem('profile')
            const loggedUserId = profile ? JSON.parse(profile)?.result?._id : null
            const followers = this.Luserdata?.followers || []

            this.isUserFollowing = Boolean(
                loggedUserId && followers.find((id) => String(id) === String(loggedUserId))
            )
        },
        async checkUserFollowing() {
            this.syncFollowingState()
        },
        async FollowOrUnFollow() {
            try {
                const userId = this.userData?._id
                if (!userId) {
                    console.log('No userId available')
                    return
                }

                const wasFollowing = this.isUserFollowing
                
                // 调用关注 API（该 API 会切换关注状态）
                const updatedUser = await this.FollowUser(userId)

                // 直接使用本次返回的数据更新本地展示，避免重复请求
                this.Luserdata = { ...(updatedUser?.user || updatedUser || this.Luserdata) }
                this.syncFollowingState()
                
                // 显示操作成功提示
                Notify.create({
                    message: wasFollowing ? '取消关注成功' : '关注成功',
                    type: 'positive',
                    timeout: 3000
                })
            } catch (error) {
                console.error('Error following/unfollowing user:', error)
                // 添加用户友好的错误提示
                Notify.create({
                    message: '操作失败，请重试',
                    type: 'negative',
                    timeout: 3000
                })
            }
        },
        Edit() {
            this.$emit('EditProfile')
        }
    },
    mounted() {
        // 只在非当前用户的资料页检查关注状态
        if (!this.isSameUser) {
            this.checkUserFollowing()
        }
    }
}
</script>