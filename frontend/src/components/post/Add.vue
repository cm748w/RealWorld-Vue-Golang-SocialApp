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
                    <q-input dense outlined v-model="post.title" autofocus placeholder="Post Title" />
                    <div class="q-mt-sm">
                        <q-input outlined v-model="post.message" placeholder="What's on your mind?" type="textarea" autogrow />
                    </div>
                    <div class="q-mt-sm q-gutter-sm">
                        <q-file v-model="file" label="Pick Image" filled style="max-width: 400px;" />
                    </div>
                    <div class="q-mt-sm q-gutter-sm row items-start">
                        <q-img
                            :src="post.selectedFile"
                            spinner-color="red"
                            style="height: 140px; max-width: 150px;"
                            class="rounded-borders"
                        />
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
            file: null
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
    },
}</script>

<style lang="scss" scoped>
.fab-btn {
  box-shadow: 0 6px 20px rgba(79, 70, 229, 0.35);
}
</style>
