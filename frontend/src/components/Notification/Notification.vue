<template>
    <q-page class="constrain q-pa-md">
        <div class="row justify-center">
            <div class="col-12 col-md-8">
                <div class="q-pa-md q-pb-xs">
                    <div class="text-h5 text-weight-bold">Notifications</div>
                </div>

                <q-list :bordered="NotifyList.length > 0" class="surface" style="overflow:hidden;">

                    <div v-for="notify in NotifyList" :key="notify._id">
                        <q-item clickable @click="MoveToThePath(notify)"
                            :class="{ 'bg-grey-3': !notify.isRead }">
                            <q-item-section top avatar>
                                <q-avatar>
                                    <img :src="notify?.user?.avatar || defaultAvatar">
                                </q-avatar>
                            </q-item-section>

                            <q-item-section>
                                <q-item-label>{{ notify?.details }}</q-item-label>
                                <q-item-label caption class="muted">{{ notify?.user?.name }}</q-item-label>
                            </q-item-section>

                            <q-item-section side v-if="!notify.isRead">
                                <q-badge color="accent" label="New" />
                            </q-item-section>
                        </q-item>
                        <q-separator spaced />
                    </div>

                    <q-item v-if="NotifyList.length === 0" class="empty-notice-item">
                        <q-item-section class="text-center muted q-pa-lg empty-notice-text">
                            <q-icon name="eva-bell-outline" size="48px" class="q-mb-md" />
                            这里目前没有关于您的通知
                        </q-item-section>
                    </q-item>
                </q-list>
            </div>
        </div>
    </q-page>
</template>

<script>
import { mapGetters, mapActions, mapState } from 'vuex'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    name:'Notification-Component',
    data(){
        return {
            NotifyList:[],
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    watch: {
        "RealTimeNotify.notifyidData": async function (notify) {
            console.log("noty", notify)
            this.NotifyList.unshift(notify)
        }
    },
    async mounted(){
        const userId = this.GetUserData?.result?._id

        if (!userId) {
            this.NotifyList = []
            return
        }
        this.NotifyList = await this.GetUnReadedNotifyNum(userId) || []
        console.log("notifilist", this.NotifyList)
        setTimeout(async () => {
            // 接口只接收 userId，之前的写法对每条未读都调一次，等于同一请求发 N 遍。
            // 这里只调一次 mark-as-read，再把本地项标记为已读。
            if (this.NotifyList.some(el => !el.isRead)) {
                await this.MarkNotifyAsReaded(userId)
                this.NotifyList.forEach(el => {
                    if (!el.isRead) el.isRead = true
                })
            }
        }, 500)
    },
    computed:{
        ...mapGetters('auth', ['GetUserData']),
        ...mapState(['RealTimeNotify']),
    },
    methods:{
        ...mapActions(['GetUnReadedNotifyNum', 'MarkNotifyAsReaded']),

        MoveToThePath(notify){
            console.log(notify.details)
            if(String(notify?.details || '').includes("post")){
                this.$router.push(`/PostDetails/${notify.targetId}`)
            } else {
                this.$router.push(`/profile/${notify.targetId}`)
            }
        }
    }
}
</script>

<style scoped>
.empty-notice-text {
    font-size: 16px;
    line-height: 1.6;
}

.empty-notice-item {
    background: transparent !important;
    border: 0 !important;
    box-shadow: none !important;
}
</style>
