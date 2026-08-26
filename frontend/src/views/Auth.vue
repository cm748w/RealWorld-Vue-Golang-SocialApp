<template>
    <div class="auth-page">
        <div class="constrain q-pa-md">
            <div class="text-center q-pa-lg q-pt-xl">
                <q-icon name="eva-camera-outline" size="56px" color="grey-9" />
                <div class="brand-font text-grey-10 text-h3 text-weight-bold q-mt-sm">Home</div>
                <div class="muted text-subtitle1 q-mt-xs">Connect, share, and discover.</div>
            </div>

            <div class="row q-col-gutter-lg items-start justify-center">
                <div class="col-12 col-md-6">
                    <q-card class="my-card surface">
                        <q-card-section>
                            <div class="section-title text-h6 q-mb-sm">Sign In</div>
                            <form @submit.prevent.stop="Login" class="q-gutter-md">
                                <q-input filled v-model="Sin_data.email" label="邮箱 *" hint="请输入邮箱" :rules="emailRules"
                                    lazy-rules />
                                <q-input filled v-model="Sin_data.password" type="password" label="密码 *" hint="请输入密码"
                                    :rules="passwordRules" lazy-rules />
                                <q-btn label="Sign In" type="submit" color="primary" unelevated class="full-width q-py-sm q-mt-sm" />
                            </form>
                        </q-card-section>
                    </q-card>
                </div>

                <div class="col-12 col-md-6">
                    <q-card class="my-card surface">
                        <q-card-section>
                            <div class="section-title text-h6 q-mb-sm">Create Account</div>
                            <form @submit.prevent.stop="Register" class="q-gutter-md">
                                <q-input filled v-model="Sup_data.firstName" label="名字 *" hint="请输入名字" :rules="nameRules"
                                    lazy-rules />
                                <q-input filled v-model="Sup_data.lastName" label="姓氏 *" hint="请输入姓氏" :rules="nameRules"
                                    lazy-rules />
                                <q-input filled v-model="Sup_data.email" label="邮箱 *" hint="请输入邮箱" :rules="emailRules"
                                    lazy-rules />
                                <q-input filled v-model="Sup_data.password" type="password" label="密码 *" hint="密码（至少6个字符）"
                                    :rules="passwordRules" lazy-rules />
                                <q-btn label="Sign Up" type="submit" color="positive" unelevated class="full-width q-py-sm q-mt-sm" />
                            </form>
                        </q-card-section>
                    </q-card>
                </div>
            </div>
        </div>
    </div>
</template>

<script>

import { mapActions } from 'vuex'

export default {
    name: 'AuthView',
    data() {
        return {
            Sin_data: {
                email: '',
                password: '',
            },
            Sup_data: {
                email: '',
                password: '',
                firstName: '',
                lastName: '',
            },
            nameRules: [
                val => !!val || '请输入姓名'
            ],
            emailRules: [
                val => !!val || '请输入邮箱',
                val => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val) || '请输入有效的邮箱地址'
            ],
            passwordRules: [
                val => !!val || '请输入密码',
                val => val.length >= 6 || '密码至少需要6个字符'
            ]
        }
    },
    methods: {
        ...mapActions('auth', ['signin', 'signup']),
        ...mapActions("RealTimeNotify", ["connectToNotify"]),
        ...mapActions(['createChatConnection']),
        async Login() {
            console.log("login in data", this.Sin_data)

            let validate = true;

            for (const rule of this.emailRules) {
                const error = rule(this.Sin_data.email);
                if (error !== true) {
                    this.$q.notify({
                        icon: 'eva-alert-circle-outline',
                        type: 'negative',
                        message: error
                    });
                    validate = false;
                    break;
                }
            }

            if (validate) {
                for (const rule of this.passwordRules) {
                    const error = rule(this.Sin_data.password);
                    if (error !== true) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: error
                        });
                        validate = false;
                        break;
                    }
                }
            }

            if (validate) {
                var formdata = { email: this.Sin_data.email, password: this.Sin_data.password }
                const data = await this.signin(formdata);
                console.log("data response", data)

                if (data?.response?.data?.message || data?.response?.data) {
                    let errorMessage = "登录失败";
                    if (data.response.data.message) {
                        errorMessage = data.response.data.message;
                    } else if (data.response.data) {
                        errorMessage = data.response.data;
                    }

                    if (errorMessage.includes('invalid email or password')) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: `邮箱或密码错误，请检查您的凭据。`
                        })
                    } else {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: `错误：${errorMessage}`
                        })
                    }
                } else {
                    this.$q.notify({
                        icon: 'eva-alert-circle-outline',
                        type: 'positive',
                        message: `登录成功，欢迎回来！`
                    })
                    this.connectToNotify()
                    this.createChatConnection()
                    this.$router.push('/')
                }
            }
        },
        async Register() {
            console.log("Register in data", this.Sup_data)

            let isValidate = true;

            for (const rule of this.nameRules) {
                const error = rule(this.Sup_data.firstName);
                if (error !== true) {
                    this.$q.notify({
                        icon: 'eva-alert-circle-outline',
                        type: 'negative',
                        message: error
                    });
                    isValidate = false;
                    break;
                }
            }

            if (isValidate) {
                for (const rule of this.nameRules) {
                    const error = rule(this.Sup_data.lastName);
                    if (error !== true) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }

            if (isValidate) {
                for (const rule of this.emailRules) {
                    const error = rule(this.Sup_data.email);
                    if (error !== true) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }

            if (isValidate) {
                for (const rule of this.passwordRules) {
                    const error = rule(this.Sup_data.password);
                    if (error !== true) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }

            if (isValidate) {
                const data = await this.signup(this.Sup_data)
                console.log("data on register", data)

                if (data?.response?.data?.message) {
                    let errorMessage = data.response.data.message;

                    if (errorMessage.includes('already exists')) {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: `该邮箱已被注册，请使用其他邮箱或直接登录。`
                        })
                    } else {
                        this.$q.notify({
                            icon: 'eva-alert-circle-outline',
                            type: 'negative',
                            message: `错误：${errorMessage}`
                        })
                    }
                } else {
                    this.$q.notify({
                        icon: 'eva-alert-circle-outline',
                        type: 'positive',
                        message: `注册成功，欢迎加入我们的社区！`
                    })
                    this.connectToNotify()
                    this.createChatConnection()
                }
            }
        },
    }
}
</script>

<style lang="scss" scoped>
.auth-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #eef2f7 0%, #ffffff 100%);
  display: flex;
  align-items: flex-start;
  justify-content: center;
}
</style>
