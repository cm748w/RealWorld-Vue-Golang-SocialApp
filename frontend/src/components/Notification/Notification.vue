<template>
    <q-page class="constrain q-pa-md">
        <div class="row q-col-gutter-lg">
            <div class="col-12">

                <q-list :bordered="NotifyList.length > 0" padding>

                    <div v-for="notify in NotifyList" :key="notify._id">
                        <q-item clickable @click="MoveToThePath(notify)" :class="{'text-red': !notify.isRead}" >
                            <q-item-section top avatar>
                                <q-avatar>
                                    <img :src="notify?.user?.avatar || 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center'">
                                </q-avatar>
                            </q-item-section>

                            <q-item-section>
                                <q-item-label>{{ notify?.details }}</q-item-label>
                                <q-item-label>{{ notify?.user?.name }}</q-item-label>
                            </q-item-section>
                        </q-item>
                        <q-separator spaced />
                    </div>

                    <q-item v-if="NotifyList.length === 0" class="empty-notice-item">
                        <q-item-section class="text-center text-grey-7 empty-notice-text">
                            尊敬的彦祖先生，您好！
                            <br/>这里目前没有关于您通知。
                        </q-item-section>
                    </q-item>


                </q-list>
            </div>
        </div>
    </q-page>


</template>

<script>
import { mapGetters, mapActions } from 'vuex'
// import { watch } from 'vue'

export default {
    name:'Notification-Component',
    data(){
        return {
            NotifyList:[]
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
        // mark notification as readed
        setTimeout(() => {
            this.NotifyList.forEach(async el =>{
                if(!el.isRead){
                    await this.MarkNotifyAsReaded(userId)
                    el.isRead = true
                }
            })
        }, 500)
    },
    computed:{
        ...mapGetters('auth', ['GetUserData'])
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
    font-family: "STKaiti", "KaiTi", "Kaiti SC", "楷体", serif;
    font-size: 24px;
    line-height: 1.6;
}

.empty-notice-item {
    background: transparent !important;
    border: 0 !important;
    box-shadow: none !important;
}
</style>