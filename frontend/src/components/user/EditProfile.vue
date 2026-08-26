<template>
    <div class="q-pa-md">
        <q-card class="surface overflow-hidden">
            <div class="profile-cover" />

            <div class="q-pa-md">
                <div class="row items-center justify-between q-mb-md">
                    <div class="text-h6 text-weight-bold">Edit Profile</div>
                    <q-btn v-if="isSameUser" label="Save" color="primary" unelevated rounded @click="Save" />
                </div>

                <div class="row items-center q-col-gutter-md">
                    <div class="col-auto">
                        <q-avatar size="88px" class="profile-avatar">
                            <img :src="Luserdata?.imageUrl || defaultAvatar" />
                        </q-avatar>
                    </div>
                    <div class="col">
                        <q-file
                            v-model="file"
                            label="Change avatar"
                            filled
                            style="max-width: 320px;"
                        />
                    </div>
                </div>

                <q-input dense outlined v-model="Luserdata.name" label="Name" class="q-mt-lg" />
                <q-input
                    outlined
                    v-model="Luserdata.bio"
                    label="Bio"
                    type="textarea"
                    autogrow
                    class="q-mt-sm"
                />
            </div>
        </q-card>
    </div>
</template>

<script>
import { mapActions } from 'vuex'
import { Notify } from 'quasar'

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    props: ['userData', 'isSameUser'],
    data() {
        return {
            file: null,
            imageReadyPromise: null,
            defaultAvatar: DEFAULT_AVATAR,
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

                const response = await this.UpdateUserData(this.Luserdata)
                const updatedUser = response?.user || response?.result || response
                this.$emit('update-user', updatedUser)
                this.$emit('EditProfile')

                Notify.create({
                    message: '保存成功',
                    type: 'positive',
                    timeout: 3000
                })
            } catch (error) {
                console.error('Error saving profile:', error)
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

<style lang="scss" scoped>
.profile-cover {
  height: 128px;
  background: linear-gradient(120deg, #4F46E5 0%, #0EA5E9 55%, #F43F5E 100%);
}

.profile-avatar {
  border: 4px solid #fff;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.15);
}
</style>
