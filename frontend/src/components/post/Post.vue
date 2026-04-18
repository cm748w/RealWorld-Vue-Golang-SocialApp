<template>
    <div>
        <!-- show post -->
         <q-card v-if="!EditPost" class="card-post q-mb-md" flat bordered>
            <q-item>
                <q-item-section avatar>
                    <q-avatar>
                        <img v-if="user?.imageUrl" :src="user?.imageUrl" />
                        <img v-else src="https://cdn-icons-png.flaticon.com/512/1077/1077063.png" />
                    </q-avatar>
                </q-item-section>
                <q-item-section>
                    <q-item-label class="text-bold">{{ user.name }}</q-item-label>
                    <q-item-label caption>
                        {{ getTime() }}
                    </q-item-label>
                </q-item-section>
            </q-item>
            <q-separator />
            <q-img
                style="cursor: pointer;"
                @click="GoToDetails"
                :src="localPost.selectedFile || 'https://ts3.tc.mm.bing.net/th/id/OIP-C.1Ahs_HVJQDBFPAPeYn2ARwHaEl?rs=1&pid=ImgDetMain&o=7&rm=3'"
            />
            <q-card-section>
                <div class="text-h6">{{ localPost.title }}</div>
                <div class="text-subtitle1">{{ localPost.message }}</div>
                <q-separator />
                <div class="text-subtitle4"
                    v-for="(comment, index) in localPost.comments"
                    :key="index">
                    {{ comment }}
                </div>

                <q-btn v-if="!UserLike" @click="Like" flat round color="red" icon="eva-heart-outline">
                    {{ LikesCount() }}
                </q-btn>
                <q-btn v-else @click="Like" flat round color="red" icon="eva-heart">
                    {{ LikesCount() }}
                </q-btn>
            </q-card-section>

            <q-input outlined v-model="form.text" label="add count..">
                <q-btn v-if="form.text !== ''" @click="AddComment" flat round color="primary"
                icon="eva-plus-square" />
            </q-input>

         </q-card>
         <!-- edit post -->
          <div v-else class="q-pa-md items-start q-gutter-md">
            <q-card class="my-card col-12">
                <q-card-section>
                    <div class="text-h6">Edit Post</div>
                    <q-input dense v-model="localPost.title" autofocus placeholder="Post Title" />
                    <div>
                        <q-input v-model="localPost.message"
                            placeholder="What's on your mind?"
                            type="textarea"
                        />
                    </div>
                    <div class="q-pa-md">
                        <q-file
                            v-model="file"
                            label="Pick Image"
                            filled
                        />
                    </div>

                    <div>
                        <q-img 
                            :src="localPost.selectedFile"
                            spinner-color="red"
                            style="height: 140px; max-width: 150px;"
                        />
                    </div>

                    <q-btn flat label="Update" v-close-popup @click="FileUpdate" />
                </q-card-section>
            </q-card>
          </div>
    </div>



</template>

<script>
import moment from 'moment';
import { mapActions,mapGetters } from 'vuex'

export default {
    name: 'PostComponent',
    props: ['post', 'EditPost'],
    data() {
        return {
            user: {},
            form: { text: '' },
            file: null,
            UserLike: false,
            localPost: {},
        }
    },
    watch: {
        file() {
            this.ConvertToBase64()
        }
    },
    methods: {
        ...mapActions('users', ['GetUserById']),
        ...mapActions(['LikePostByUser', 'commentPost', 'updatePost']),

        GoToDetails(){
            this.$router.push({path:`/PostDetails/${this.localPost?._id}`})
        },
        async FileUpdate(){
            const PostData = {
                id: this.localPost._id,
                title: this.localPost.title,
                selectedFile: this.localPost.selectedFile,
                message: this.localPost.message,
            }

            const res = await this.updatePost(PostData)
            if(res){
                this.$emit('changeEdit')
            }

        },
        getTime(){
            return moment(this.localPost?.createdAt).fromNow()
        },
        Like(){
            this.LikePostByUser(this.localPost._id)
            const uid = this.GetUserData?.result?._id
            const likes = Array.isArray(this.localPost.likes) ? this.localPost.likes : []
            if(this.UserLike){
                this.localPost.likes = likes.filter(id => id != uid)
            } else {
                this.localPost.likes = [...likes, uid]
            }
            this.UserLike = !this.UserLike
        },
        LikesCount(){
            if(this.localPost.likes?.length > 0){
                return String(this.localPost.likes?.length)
            }
        },
        AddComment(){
            // console.log("comment ", this.form.text)
            this.localPost.comments.push(this.form.text)
            // store
            this.commentPost({ value: this.form.text, id:this.localPost._id})
            this.form.text = ''
        },

        ConvertToBase64() {
            var reader = []
            reader = new FileReader()
            reader.readAsDataURL(this.file)
            new Promise(()=> {
                reader.onload = () => {
                    this.localPost.selectedFile = reader.result
                }
            })
        }
    },
    computed:{
        ...mapGetters('auth', ['GetUserData']),
    },
    async mounted(){
        // Create local copy of post prop
        this.localPost = JSON.parse(JSON.stringify(this.post))

        const response = await this.GetUserById(this.localPost?.creator)
        const { user } = response || {}
        this.user = user || {}
        // get if user liked the post or not
        const uid = this.GetUserData?.result?._id
        const likes = Array.isArray(this.localPost.likes) ? this.localPost.likes : []
        var isLike = likes.find((like)=> like == uid)
        if(isLike){
            this.UserLike = true
        } else {
            this.UserLike = false
        }
    }
}
</script>