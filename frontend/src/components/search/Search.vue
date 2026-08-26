<template>
    <q-page class="constrain q-pa-md">
        <div class="row justify-center">
            <div class="col-12 col-md-8">
                <div class="q-pa-md text-center">
                    <div class="text-h5 text-weight-bold q-mb-md">Search Results</div>
                    <q-btn-toggle v-model="model"
                        toggle-color="primary"
                        unelevated
                        no-caps
                        class="q-toggle"
                        :options="[
                            {label: 'Posts', value: 'Posts'},
                            {label: 'Users', value: 'Users'},
                        ]"
                    />
                </div>

                <div class="q-pa-md q-pt-sm">
                    <q-list separator bordered class="surface" style="overflow:hidden">

                        <template v-if="model == 'Users'">
                            <q-item v-for="data in Users" :key="data._id" clickable v-ripple @click="GoUser(data._id)">
                                <q-item-section avatar>
                                    <q-avatar>
                                        <img v-if="data?.imageUrl" :src="data?.imageUrl" />
                                        <img v-else :src="defaultAvatar" />
                                    </q-avatar>
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label class="text-weight-bold">{{ data?.name }}</q-item-label>
                                    <q-item-label caption class="ellipsis">{{ data?.bio }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item v-if="!Users.length">
                                <q-item-section class="text-center muted q-pa-lg">No users found</q-item-section>
                            </q-item>
                        </template>

                        <template v-if="model == 'Posts'">
                            <q-item v-for="post in Posts" :key="post._id" clickable v-ripple @click="GoPost(post._id)">
                                <q-item-section thumbnail>
                                    <img :src="post.selectedFile" class="rounded-borders" style="width:48px;height:48px;object-fit:cover;">
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label class="text-weight-bold">{{ post?.title }}</q-item-label>
                                    <q-item-label caption class="ellipsis">{{ post?.message }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item v-if="!Posts.length">
                                <q-item-section class="text-center muted q-pa-lg">No posts found</q-item-section>
                            </q-item>
                        </template>

                    </q-list>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { mapActions } from 'vuex'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    name:'SearchComponent',
    data(){
        return {
            value:true,
            model:'Posts',
            Users:[],
            Posts:[],
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    watch:{
        $route(){
            this.GetData()
        }
    },
    methods:{
        ...mapActions(['getPostsUsersBySearch']),
        async GetData(){
            const AllData = await this.getPostsUsersBySearch(String(this.$route.query.search))
            console.log("s", AllData)
            this.Users = AllData?.user || []
            this.Posts = AllData?.posts || []
        },
        GoUser(id){
            this.$router.push({path: `/profile/${id}`})
        },
        GoPost(id){
            this.$router.push({path: `/PostDetails/${id}`})
        },
    },
    mounted(){
        this.GetData()
    }
}
</script>

<style lang="scss" scoped>
.q-toggle {
  .q-btn {
    border-radius: 999px;
  }
}
</style>
