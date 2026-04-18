<template>
    <q-page-sticky position="bottom-left" v-show="GetUserData?.result">
        <div class="q-pa-md q-gutter-sm">
            <q-btn label="Create Post" style="cursor: pointer;" icon="eva-plus-circle-outline" 
                color="primary" @click="persistent = true" />
                
            <!-- popup -->
             <q-dialog v-model="persistent" persistent transition-show="scale" transition-hide="scale">
                <q-card style="min-width: 350px;" >
                    <q-card-section>
                        <div class="text-h6">Create Post</div>
                    </q-card-section>

                    <q-card-section class="q-pt-none">
                        <q-input dense v-model="post.title" autofocus placeholder="Post Title" />
                        <div class="q-pa-md" style="max-width: 300px;">
                            <q-input
                                v-model="post.message"
                                placeholder="What's on your mind?"
                                type="textarea"
                            />
                        </div>
                        <div class="q-pa-md">
                            <q-file
                                v-model="file"
                                label="Pick Image"
                                filled
                                style="max-width: 400px;"
                            />
                        </div>
                        <div class="q-gutter-sm row items-start">
                            <q-img
                                :src="post.selectedFile"
                                spinner-color="red"
                                style="height: 140px; max-width: 150px;"
                            />
                        </div>
                    </q-card-section>

                    <q-card-actions align="right" class="text-primary">
                        <q-btn flat label="Create" v-close-popup @click="CreatePost" />
                        <q-btn flat label="Cancel" v-close-popup />
                        
                        
                    </q-card-actions>
                </q-card>
             </q-dialog>
<!-- -------------------------------------------------------------------------------------- -->
        </div>

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
    watch: {
        file() {
            // 文件类型验证
            if (this.file && !this.file.type.match('image.*')) {
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'Only image files are allowed'
                })
                this.file = null
                return
            }
            // 文件大小验证（限制为5MB）
            if (this.file && this.file.size > 5 * 1024 * 1024) {
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'File size must be less than 5MB'
                })
                this.file = null
                return
            }
            // convert fun
            this.ConvertToBase64()
        }
    },
    computed: {
        ...mapGetters('auth', ['GetUserData'])
    },
    methods: {
        ...mapActions(['createPost']),
        // XSS防护函数
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
            // 优先使用GetUserData，如果没有再从localStorage获取
            var name = this.GetUserData?.result?.name || JSON.parse(localStorage.getItem('profile'))?.result?.name
            this.post.name = name
            const title = (this.post.title || '').trim()
            const message = (this.post.message || '').trim()
            // validation
            var isValidate = true
            // 只验证必要字段
            if (!title) {
                isValidate = false
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'Title is required'
                })
            }
            if (!message) {
                isValidate = false
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'Message is required'
                })
            }
            // after validate
            if(isValidate){
                try {
                    // XSS防护
                    this.post.title = this.sanitizeInput(title)
                    this.post.message = this.sanitizeInput(message)
                    const data = await this.createPost(this.post)
                    if (data) {
                        this.$emit('created')
                        // 清空表单
                        this.post = { title: '', message: '', name: '', selectedFile: null }
                        this.file = null
                        this.persistent = false  // 关闭对话框
                        this.$q.notify({
                            icon: 'eva-check-circle-outline',
                            type: 'positive',
                            message: 'Post created successfully'
                        })
                    } else {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: 'Failed to create post'
                        })
                    }
                } catch (error) {
                    const responseData = error?.response?.data
                    const errorMessage = responseData?.message || responseData || 'Failed to create post'
                    this.$q.notify({
                        icon: 'eva-alert-circle-outline',
                        type: 'negative',
                        message: errorMessage
                    })
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
                this.$q.notify({
                    icon: 'eva-alert-circle-outline',
                    type: 'negative',
                    message: 'Failed to read file'
                })
            }
        },
    },
}</script>