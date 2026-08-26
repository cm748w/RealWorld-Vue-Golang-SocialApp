<template>
    <q-page-sticky position="bottom-right" :offset="fabOffset" v-show="GetUserData?.result">
        <q-btn
            :label="isMobileNav ? '' : 'Create Post'"
            icon="eva-plus-circle-outline"
            color="primary"
            rounded
            unelevated
            class="fab-btn q-mr-md"
            @click="persistent = true"
        />

        <!-- popup -->
        <q-dialog v-model="persistent" persistent transition-show="scale" transition-hide="scale" :maximized="isMobileNav">
            <q-card class="surface" :style="cardStyle">
                <q-card-section class="row items-center justify-between q-pb-none">
                    <div class="text-h6 text-weight-bold">Create Post</div>
                    <q-btn flat round dense icon="eva-close-outline" color="grey-7" @click="persistent = false" />
                </q-card-section>

                <q-card-section class="q-pt-md">
                    <q-input dense outlined v-model="post.title" autofocus placeholder="Post Title"
                        maxlength="120" counter />
                    <div class="q-mt-sm">
                        <q-input outlined v-model="post.message" placeholder="What's on your mind?"
                            type="textarea" autogrow maxlength="500" counter />
                    </div>

                    <div class="row q-gutter-xs q-mt-sm no-scrollbar">
                        <q-btn v-for="emoji in emojis" :key="emoji" flat round dense size="sm"
                            class="emoji-btn" @click="addEmoji(emoji)">
                            <span style="font-size:1.1rem;">{{ emoji }}</span>
                        </q-btn>
                    </div>

                    <div class="q-mt-sm q-gutter-sm">
                        <q-file v-model="file" label="Pick Image" filled style="max-width: 400px;" />
                    </div>

                    <div v-if="post.selectedFile" class="q-mt-sm row items-center q-gutter-sm">
                        <q-img
                            :src="post.selectedFile"
                            spinner-color="red"
                            style="height: 120px; max-width: 140px;"
                            class="rounded-borders"
                        />
                        <q-btn flat round dense icon="eva-trash-2-outline" color="negative"
                            @click="clearImage" aria-label="Remove image" />
                    </div>
                </q-card-section>

                <q-card-actions align="right" class="q-pa-md">
                    <q-btn flat label="Cancel" color="grey-7" @click="persistent = false" />
                    <q-btn label="Create" color="primary" @click="CreatePost" />
                </q-card-actions>
            </q-card>
        </q-dialog>
    </q-page-sticky>
</template>

<script>

import { mapActions, mapGetters } from 'vuex'

export default {
    name: 'AddComponent',
    data() {
        return {
            persistent: false,
            post: { title: '', message: '', name: '', selectedFile: null },
            file: null,
            emojis: ['😀', '😂', '😍', '😎', '🥳', '👍', '❤️', '🔥', '🎉', '🙏', '💯', '🤔']
        }
    },
    computed: {
        ...mapGetters('auth', ['GetUserData']),
        isMobileNav() {
            return this.$q.screen.lt.md
        },
        fabOffset() {
            return this.isMobileNav ? [16, 84] : [16, 16]
        },
        cardStyle() {
            return `width: min(90vw, 500px);`
        }
    },
    watch: {
        file() {
            if (this.file && !this.file.type.match('image.*')) {
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'Only image files are allowed'
                })
                this.file = null
                return
            }
            if (this.file && this.file.size > 5 * 1024 * 1024) {
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'File size must be less than 5MB'
                })
                this.file = null
                return
            }
            this.ConvertToBase64()
        }
    },
    methods: {
        ...mapActions(['createPost']),
        sanitizeInput(input) {
            return input.replace(/[&<>'"]/g, function(match) {
                const sanitizeMap = {
                    '&': '&amp;',
                    '<': '&lt;',
                    '>': '&gt;',
                    "'": '&#39;',
                    '"': '&quot;'
                }
                return sanitizeMap[match]
            })
        },
        async CreatePost() {
            var name = this.GetUserData?.result?.name || JSON.parse(localStorage.getItem('profile'))?.result?.name
            this.post.name = name
            const title = (this.post.title || '').trim()
            const message = (this.post.message || '').trim()
            var isValidate = true
            if (!title) {
                isValidate = false
                this.$q.notify({ icon: 'eva-alert-circle-outline', type: 'negative', message: 'Title is required' })
            }
            if (!message) {
                isValidate = false
                this.$q.notify({ icon: 'eva-alert-circle-outline', type: 'negative', message: 'Message is required' })
            }
            if(isValidate){
                try {
                    this.post.title = this.sanitizeInput(title)
                    this.post.message = this.sanitizeInput(message)
                    const data = await this.createPost(this.post)
                    if (data) {
                        this.$emit('created')
                        this.post = { title: '', message: '', name: '', selectedFile: null }
                        this.file = null
                        this.persistent = false
                        this.$q.notify({ icon: 'eva-check-circle-outline', type: 'positive', message: 'Post created successfully' })
                    } else {
                        this.$q.notify({ icon: 'eva-alert-circle-outline', type: 'negative', message: 'Failed to create post' })
                    }
                } catch (error) {
                    const responseData = error?.response?.data
                    const errorMessage = responseData?.message || responseData || 'Failed to create post'
                    this.$q.notify({ icon: 'eva-alert-circle-outline', type: 'negative', message: errorMessage })
                }
            }
        },
        ConvertToBase64(){
            if (!this.file) return
            var reader = new FileReader()
            reader.readAsDataURL(this.file)
            reader.onload = ()=> {
                this.post.selectedFile = reader.result
            }
            reader.onerror = ()=> {
                this.$q.notify({ icon: 'eva-alert-circle-outline', type: 'negative', message: 'Failed to read file' })
            }
        },
        addEmoji(emoji) {
            this.post.message = (this.post.message || '') + emoji
        },
        clearImage() {
            this.post.selectedFile = null
            this.file = null
        },
    },
}</script>

<style lang="scss" scoped>
.fab-btn {
  box-shadow: 0 6px 20px rgba(79, 70, 229, 0.35);
}

.emoji-btn {
  font-size: 1.2rem;
  transition: transform 0.1s ease, background 0.1s ease;

  &:hover {
    transform: scale(1.2);
  }
}
</style>
