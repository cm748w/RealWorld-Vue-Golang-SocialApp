<template>
    <div class="post-card q-mb-md">
        <!-- show post -->
        <q-card v-if="!EditPost" class="post-view" flat bordered>
            <q-item>
                <q-item-section avatar>
                    <q-avatar class="post-avatar">
                        <img v-if="user?.imageUrl" :src="user?.imageUrl" />
                        <img v-else :src="defaultAvatar" />
                    </q-avatar>
                </q-item-section>
                <q-item-section>
                    <q-item-label class="text-weight-bold">{{ user.name }}</q-item-label>
                    <q-item-label caption>{{ getTime() }}</q-item-label>
                </q-item-section>
            </q-item>

            <q-img
                v-if="localPost.selectedFile"
                class="post-img"
                style="cursor: pointer;"
                @click="GoToDetails"
                :src="localPost.selectedFile"
            />

            <q-card-section>
                <div class="text-subtitle1 text-weight-bold q-mb-xs">{{ localPost.title }}</div>
                <div class="text-body1" style="white-space: pre-wrap;">{{ localPost.message }}</div>

                <div v-if="localPost.comments?.length" class="q-mt-sm">
                    <div class="muted text-caption q-mb-xs text-weight-medium">Comments</div>
                    <div class="text-body2 muted" v-for="(comment, index) in localPost.comments" :key="index">
                        <q-icon name="eva-message-circle-outline" size="14px" class="q-mr-xs" />
                        {{ comment }}
                    </div>
                </div>

                <q-separator class="q-my-sm" />

                <div class="row items-center justify-between q-gutter-sm">
                    <q-btn flat round :color="UserLike ? 'accent' : undefined" @click="debouncedLike">
                        <q-icon :name="UserLike ? 'eva-heart' : 'eva-heart-outline'" size="20px" />
                        <span v-if="LikesCount()" class="q-ml-xs">{{ LikesCount() }}</span>
                    </q-btn>
                </div>
            </q-card-section>

            <q-card-actions>
                <q-input
                    outlined dense
                    v-model="form.text"
                    label="Add a comment..."
                    class="col q-px-sm"
                    @keyup.enter="AddComment"
                >
                    <template v-slot:append>
                        <q-btn v-if="form.text !== ''" @click="AddComment" flat round color="accent" icon="eva-plus-square" />
                    </template>
                </q-input>
            </q-card-actions>
        </q-card>

        <!-- edit post -->
        <div v-else class="q-pa-md items-start q-gutter-md">
            <q-card class="col-12 surface">
                <q-card-section>
                    <div class="text-h6 q-mb-sm">Edit Post</div>
                    <q-input dense outlined v-model="localPost.title" autofocus placeholder="Post Title"
                        maxlength="120" counter />
                    <div class="q-mt-sm">
                        <q-input outlined v-model="localPost.message" placeholder="What's on your mind?"
                            type="textarea" autogrow maxlength="500" counter />
                    </div>
                    <div class="q-pa-md q-pl-0">
                        <q-file v-model="file" label="Pick Image" filled />
                    </div>
                    <div v-if="localPost.selectedFile" class="q-mt-sm row items-center q-gutter-sm">
                        <q-img :src="localPost.selectedFile" spinner-color="red"
                            style="height: 140px; max-width: 200px;" class="rounded-borders" />
                        <q-btn flat round dense icon="eva-trash-2-outline" color="negative"
                            @click="clearImage" aria-label="Remove image" />
                    </div>
                    <q-btn flat label="Update" class="btn-inherit" @click="FileUpdate" />
                </q-card-section>
            </q-card>
        </div>
    </div>
</template>

<script>
import moment from 'moment';
import { mapActions, mapGetters } from 'vuex'
import { debounce } from '@/utils/timing'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

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
            defaultAvatar: DEFAULT_AVATAR,
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
            this.localPost.comments.push(this.form.text)
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
        },
        clearImage() {
            this.localPost.selectedFile = null
            this.file = null
        }
    },
    computed:{
        ...mapGetters('auth', ['GetUserData']),
    },
    created() {
        // 防抖（leading 模式）：第一次点击立即生效，800ms 内的重复点击直接丢弃
        this.debouncedLike = debounce(this.Like, 800, { leading: true, trailing: false })
    },
    async mounted(){
        this.localPost = JSON.parse(JSON.stringify(this.post))

        // 拉取作者资料；后端对 user/getUser 有 30次/分/IP 的限流，
        // 万一失败就回退到帖子内嵌的作者名，避免作者名空白。
        const response = await this.GetUserById(this.localPost?.creator)
        const fetched = response?.user || {}
        this.user = fetched._id ? fetched : { name: this.localPost?.name || 'Unknown' }
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

<style lang="scss" scoped>
.post-card {
  .post-view {
    background: #fff;
  }

  .post-avatar {
    box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.55);
  }

  .post-img {
    width: 100%;
    border-radius: 12px;
    padding: 0 16px;
  }
}

// Dark mode: post card surface
body.body--dark .post-card .post-view {
  background: #101828;
}
</style>
