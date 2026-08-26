<template>
    <q-page class="constrain q-pa-md">
        <div class="row justify-center">
            <div class="col-12 col-md-6 q-mx-auto">
                <div class="q-pa-md q-gutter-sm" v-if="post && !EditPost">
                    <q-btn v-if="IsSameUser" color="primary" unelevated rounded icon="eva-edit" label="Edit Post"
                        @click="EditPost = !EditPost" />
                </div>

                <Post v-if="post" :post="post" :EditPost="EditPost" @changeEdit="EditPost = !EditPost" />

                <div v-else class="q-pa-lg text-center muted">
                    <q-spinner color="" size="3em" />
                    <div class="q-mt-md">Loading post...</div>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import Post from './Post.vue'
import { mapActions } from 'vuex';
export default {
    name:'PostDetails',
    data(){
        return {
            EditPost: false,
            post: null,
            IsSameUser: false,
        }
    },
    methods:{
        ...mapActions(['GetPost']),
    },
    async mounted(){
        const response = await this.GetPost(this.$route.params.id)
        const post = response?.post || response || null
        this.post = post

        const logedInUser = JSON.parse(localStorage.getItem('profile'))
        const LogedInUserId = logedInUser?.result?._id
        if(post?.creator == LogedInUserId){
            this.IsSameUser = true
        }
    },
    components:{
        Post
    }
}
</script>
