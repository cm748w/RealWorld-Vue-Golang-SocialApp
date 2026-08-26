<template>
    <div class="q-pa-md">
        <q-card class="surface overflow-hidden">
            <div class="profile-cover" />

            <div class="q-pa-md">
                <div class="row items-center q-col-gutter-md">
                    <div class="col-auto">
                        <q-avatar size="84px" class="profile-avatar">
                            <img v-if="Luserdata?.imageUrl" :src="Luserdata?.imageUrl">
                            <img v-else :src="defaultAvatar">
                        </q-avatar>
                    </div>
                    <div class="col">
                        <div class="text-h5 text-weight-bold ellipsis">{{ Luserdata?.name }}</div>
                        <div class="muted q-mt-xs" v-if="Luserdata?.bio">{{ Luserdata?.bio }}</div>
                    </div>
                </div>

                <q-separator class="q-my-lg" />

                <div class="row q-gutter-lg items-center">
                    <span class="muted"><b class="text-dark">{{ userPosts?.length || 0 }}</b> Posts</span>
                    <span class="muted"><b class="text-dark">{{ Luserdata?.followers?.length || 0 }}</b> Followers</span>
                    <span class="muted"><b class="text-dark">{{ Luserdata?.following?.length || 0 }}</b> Following</span>
                </div>
            </div>

            <q-separator />

            <div class="q-pa-md">
                <q-btn v-if="isSameUser" label="Edit Profile" color="primary" unelevated rounded class="full-width" @click="Edit" />
                <q-btn v-else-if="!isUserFollowing" label="Follow" color="accent" unelevated rounded class="full-width" @click="FollowOrUnFollow" />
                <q-btn v-else label="Following" color="grey-6" unelevated rounded class="full-width" @click="FollowOrUnFollow" />
            </div>
        </q-card>
    </div>
</template>

<script>
import { mapActions } from 'vuex'
import { Notify } from 'quasar'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    props: ['userData', 'userPosts', 'isSameUser'],
    data() {
        return {
            isUserFollowing: false,
            Luserdata: {},
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    watch: {
        userData: {
            handler(newVal) {
                if (newVal) {
                    this.Luserdata = { ...newVal }
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
                const updatedUser = await this.FollowUser(userId)

                this.Luserdata = { ...(updatedUser?.user || updatedUser || this.Luserdata) }
                this.syncFollowingState()

                Notify.create({
                    message: wasFollowing ? '取消关注成功' : '关注成功',
                    type: 'positive',
                    timeout: 3000
                })
            } catch (error) {
                console.error('Error following/unfollowing user:', error)
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
        if (!this.isSameUser) {
            this.checkUserFollowing()
        }
    }
}
</script>

<style lang="scss" scoped>
.profile-cover {
  height: 128px;
  background: linear-gradient(120deg, #0EA5E9 0%, #38BDF8 55%, #F43F5E 100%);
}

.profile-avatar {
  box-shadow: 0 0 0 4px #fff, 0 4px 16px rgba(15, 23, 42, 0.15);
}
</style>
