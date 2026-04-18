<template>
    <div class="row col-12 constrain">
        <div class="col-4 text-center">
            <q-avatar size="150px">
                <img :src="Luserdata?.imageUrl" />
            </q-avatar>
        </div>
        <div class="col-8 text-left">
            <i>Edit Profile</i>
            <div class="text-h6 q-pa-lg" style="margin: auto;">
                <q-btn v-if="isSameUser" @click="Save" flat label="Save" />
            </div>

            <q-input dense v-model="Luserdata.name" autofocus placeholder="User Data Title" />

            <div>
                <q-input 
                    v-model="Luserdata.bio" 
                    placeholder="User Data Bio" 
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
        </div>
    </div>
</template>


<script>
import { mapActions } from 'vuex'
import { Notify } from 'quasar'

export default {
    props: ['userData', 'isSameUser'],
    data() {
        return {
            file: null,
            imageReadyPromise: null,
            Luserdata: {
                ...this.userData
            }
        }
    },
    watch: {
        file(newFile){
            this.imageReadyPromise = this.ConvertToBase64(newFile)
        },
        userData: {
            handler(newVal) {
                if (newVal) {
                    this.Luserdata = { ...newVal }
                }
            },
            immediate: true
        }
    },
    methods: {
        ...mapActions('users', ['FollowUser', 'GetUserById', 'UpdateUserData']),
        ConvertToBase64(file){
            if (!file) {
                this.Luserdata.imageUrl = this.userData?.imageUrl
                return Promise.resolve(this.Luserdata.imageUrl)
            }

            return new Promise((resolve, reject) => {
                const reader = new FileReader()

                reader.onload = () => {
                    this.Luserdata.imageUrl = reader.result
                    resolve(reader.result)
                }

                reader.onerror = () => {
                    reject(new Error('图片读取失败'))
                }

                reader.readAsDataURL(file)
            })
        },
        async Save() {
            try {
                if (this.imageReadyPromise) {
                    await this.imageReadyPromise
                }

                // 调用更新用户数据的 API
                const response = await this.UpdateUserData(this.Luserdata)
                
                // 发送事件通知父组件更新，只发送用户数据部分
                const updatedUser = response?.user || response?.result || response
                this.$emit('update-user', updatedUser)
                this.$emit('EditProfile')
                
                // 显示保存成功提示
                Notify.create({
                    message: '保存成功',
                    type: 'positive',
                    timeout: 3000
                })
            } catch (error) {
                console.error('Error saving profile:', error)
                // 显示保存失败提示
                Notify.create({
                    message: '保存失败，请重试',
                    type: 'negative',
                    timeout: 3000
                })
            }
        }
    },
}
</script>