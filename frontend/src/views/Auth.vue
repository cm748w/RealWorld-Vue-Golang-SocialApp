<template>
    <div class="row q-col-gutter-lg">
        <div class="col-5">
            <q-card class="my-card" style="width: 100%; padding: 10px;">
                <h1 class="text-h6 text-center">登录</h1>
                <q-card-section>
                    <form @submit.prevent.stop="Login" class="q-gutter-md">
                        <q-input 
                            filled 
                            v-model="Sin_data.email" 
                            label="邮箱 *" 
                            hint="请输入邮箱" 
                            :rules="emailRules"
                            lazy-rules 
                        />
                        <q-input 
                            filled 
                            v-model="Sin_data.password" 
                            type="password" 
                            label="密码 *"
                            hint="请输入密码" 
                            :rules="passwordRules"
                            lazy-rules 
                        />
                        <div>
                            <q-btn label="登录" type="submit" color="primary" />
                        </div>
                    </form>
                </q-card-section>

            </q-card>
        </div>
        <div class="col-7">
            <q-card class="my-card" style="width: 100%; padding: 10px;">
                <h1 class="text-h6 text-center">注册</h1>
                <q-card-section>
                    <form @submit.prevent.stop="Register" class="q-gutter-md">
                        <q-input 
                            filled 
                            v-model="Sup_data.firstName" 
                            label="名字 *" 
                            hint="请输入名字"
                            :rules="nameRules"
                            lazy-rules 
                        />
                        <q-input 
                            filled 
                            v-model="Sup_data.lastName" 
                            label="姓氏 *" 
                            hint="请输入姓氏"
                            :rules="nameRules"
                            lazy-rules 
                        />
                        <q-input 
                            filled 
                            v-model="Sup_data.email" 
                            label="邮箱 *" 
                            hint="请输入邮箱"
                            :rules="emailRules"
                            lazy-rules 
                        />
                        <q-input 
                            filled 
                            v-model="Sup_data.password" 
                            type="password" 
                            label="密码 *"
                            hint="密码（至少6个字符）"
                            :rules="passwordRules"
                            lazy-rules 
                        />
                        <div>
                            <q-btn label="注册" type="submit" color="positive" />
                        </div>
                    </form>
                </q-card-section>

            </q-card>
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
            // 表单验证规则
            nameRules: [
                val => !!val || '请输入姓名',
                val => val.length >= 2 || '姓名至少需要2个字符'
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

        async Login() {
            console.log("login in data", this.Sin_data)
            
            // 手动验证所有字段
            let validate = true;
            
            // 验证邮箱
            for (const rule of this.emailRules) {
                const error = rule(this.Sin_data.email);
                if (error !== true) {
                    this.$q.notify({
                        icon:'eva-alert-circle-outline',
                        type:'negative',
                        message: error
                    });
                    validate = false;
                    break;
                }
            }
            
            if (validate) {
                // 验证密码
                for (const rule of this.passwordRules) {
                    const error = rule(this.Sin_data.password);
                    if (error !== true) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message: error
                        });
                        validate = false;
                        break;
                    }
                }
            }
            
            // success and to do next
            if(validate){
                var formdata = {email:this.Sin_data.email, password:this.Sin_data.password}
                const data = await this.signin(formdata);
                console.log("data response", data)

                if(data?.response?.data?.message || data?.response?.data){
                    let errorMessage = "登录失败";
                    if (data.response.data.message) {
                        errorMessage = data.response.data.message;
                    } else if (data.response.data) {
                        errorMessage = data.response.data;
                    }
                    
                    // 更详细的错误提示
                    if (errorMessage.includes('invalid email or password')) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message:`邮箱或密码错误，请检查您的凭据。`
                        })
                    } else {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message:`错误：${errorMessage}`
                        })
                    }
                } else {
                    this.$q.notify({
                        icon:'eva-alert-circle-outline',
                        type:'positive',
                        message:`登录成功，欢迎回来！`
                    })
                }
            }
        },
        // -----------------------------------------------------------------------------
        async Register() {
            console.log("Register in data", this.Sup_data)
            
            // 手动验证所有字段
            let isValidate = true;
            
            // 验证名字
            for (const rule of this.nameRules) {
                const error = rule(this.Sup_data.firstName);
                if (error !== true) {
                    this.$q.notify({
                        icon:'eva-alert-circle-outline',
                        type:'negative',
                        message: error
                    });
                    isValidate = false;
                    break;
                }
            }
            
            if (isValidate) {
                // 验证姓氏
                for (const rule of this.nameRules) {
                    const error = rule(this.Sup_data.lastName);
                    if (error !== true) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }
            
            if (isValidate) {
                // 验证邮箱
                for (const rule of this.emailRules) {
                    const error = rule(this.Sup_data.email);
                    if (error !== true) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }
            
            if (isValidate) {
                // 验证密码
                for (const rule of this.passwordRules) {
                    const error = rule(this.Sup_data.password);
                    if (error !== true) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message: error
                        });
                        isValidate = false;
                        break;
                    }
                }
            }
            
            if(isValidate){
                const data = await this.signup(this.Sup_data)
                console.log("data on register", data)
                
                if(data?.response?.data?.message){
                    let errorMessage = data.response.data.message;
                    
                    // 更详细的错误提示
                    if (errorMessage.includes('already exists')) {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message:`该邮箱已被注册，请使用其他邮箱或直接登录。`
                        })
                    } else {
                        this.$q.notify({
                            icon:'eva-alert-circle-outline',
                            type:'negative',
                            message:`错误：${errorMessage}`
                        })
                    }
                } else {
                    // 意味着成功了
                    this.$q.notify({
                        icon:'eva-alert-circle-outline',
                        type:'positive',
                        message:`注册成功，欢迎加入我们的社区！`
                    })
                    // this.$router.push('/')
                }
            }
        },
    }
}
</script>