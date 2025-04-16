<template>
  <div class="login register">
    <div class="left-wrap">
      <left-view></left-view>
    </div>
    <div class="right-wrap">
      <div class="header">
        <svg
          class="icon"
          aria-hidden="true"
          viewBox="0 0 16 16"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
        >
          <path
            fill="#000000"
            fill-rule="evenodd"
            d="M13 0a3 3 0 00-1.65 5.506 7.338 7.338 0 01-.78 1.493c-.22.32-.472.635-.8 1.025a1.509 1.509 0 00-.832.085 12.722 12.722 0 00-1.773-1.124c-.66-.34-1.366-.616-2.215-.871a1.5 1.5 0 10-2.708 1.204c-.9 1.935-1.236 3.607-1.409 5.838a1.5 1.5 0 101.497.095c.162-2.07.464-3.55 1.25-5.253.381-.02.725-.183.979-.435.763.23 1.367.471 1.919.756a11.13 11.13 0 011.536.973 1.5 1.5 0 102.899-.296c.348-.415.64-.779.894-1.148.375-.548.665-1.103.964-1.857A3 3 0 1013 0zm-1.5 3a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0z"
            clip-rule="evenodd"
          />
        </svg>

        <h1>{{ systemName }}</h1>
      </div>
      <div class="login-wrap">
        <div class="form">
          <h3 class="title">{{ $t('register.title') }}</h3>
          <p class="sub-title">{{ $t('register.subTitle') }}</p>
          <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
            <el-form-item prop="username">
              <el-input
                v-model.trim="formData.username"
                :placeholder="$t('register.placeholder[0]')"
                size="large"
              />
            </el-form-item>

            <el-form-item prop="email">
              <el-input
                v-model.trim="formData.email"
                :placeholder="$t('register.placeholder[1]')"
                size="large"
              />
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model.trim="formData.password"
                :placeholder="$t('register.placeholder[2]')"
                size="large"
                type="password"
                autocomplete="off"
              />
            </el-form-item>

            <el-form-item prop="confirmPassword">
              <el-input
                v-model.trim="formData.confirmPassword"
                :placeholder="$t('register.placeholder[3]')"
                size="large"
                type="password"
                autocomplete="off"
                @keyup.enter="register"
              />
            </el-form-item>

            <el-form-item prop="agreement">
              <el-checkbox v-model="formData.agreement">
                {{ $t('register.agreeText') }}
                <router-link
                  style="color: var(--main-color); text-decoration: none"
                  to="/privacy-policy"
                  >{{ $t('register.privacyPolicy') }}</router-link
                >
              </el-checkbox>
            </el-form-item>

            <div style="margin-top: 15px">
              <el-button
                class="register-btn"
                size="large"
                type="primary"
                @click="register"
                :loading="loading"
                v-ripple
              >
                {{ $t('register.submitBtnText') }}
              </el-button>
            </div>

            <div class="footer">
              <p>
                {{ $t('register.hasAccount') }}
                <router-link to="/login">{{ $t('register.toLogin') }}</router-link>
              </p>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import LeftView from '@/components/Pages/Login/LeftView.vue'
  import AppConfig from '@/config'
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import api from '@/utils/http'
  import { BaseResult } from '@/types/axios'
  import type { Action } from 'element-plus'

  const { t } = useI18n()

  const router = useRouter()
  const formRef = ref<FormInstance>()

  const systemName = AppConfig.systemInfo.name
  const loading = ref(false)

  const formData = reactive({
    username: '',
    password: '',
    confirmPassword: '',
    email: '',
    agreement: false
  })

  const validatePass = (rule: any, value: string, callback: any) => {
    if (value === '') {
      callback(new Error(t('register.placeholder[1]')))
    } else {
      if (formData.confirmPassword !== '') {
        formRef.value?.validateField('confirmPassword')
      }
      callback()
    }
  }

  const validatePass2 = (rule: any, value: string, callback: any) => {
    if (value === '') {
      callback(new Error(t('register.rule[0]')))
    } else if (value !== formData.password) {
      callback(new Error(t('register.rule[1]')))
    } else {
      callback()
    }
  }

  const rules = reactive<FormRules>({
    username: [
      { required: true, message: t('register.placeholder[0]'), trigger: 'blur' },
      { min: 3, max: 20, message: t('register.rule[2]'), trigger: 'blur' }
    ],
    password: [
      { required: true, validator: validatePass, trigger: 'blur' },
      { min: 6, message: t('register.rule[3]'), trigger: 'blur' }
    ],
    confirmPassword: [{ required: true, validator: validatePass2, trigger: 'blur' }],
    agreement: [
      {
        validator: (rule: any, value: boolean, callback: any) => {
          if (!value) {
            callback(new Error(t('register.rule[4]')))
          } else {
            callback()
          }
        },
        trigger: 'change'
      }
    ],
    email: [
      { required: true, message: '请输入邮箱地址', trigger: 'blur' },
      {
        type: 'email',
        message: '请输入正确的邮箱格式',
        trigger: ['blur', 'change']
      }
    ]
  })

  const register = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      loading.value = true

      // 模拟注册请求
      await api
        .post<BaseResult>({
          url: '/api/v1/users/register',
          data: {
            code: 200,
            message: '更新用户数据',
            data: {
              email: formData.email,
              username: formData.username,
              password: formData.password
            }
          }
        })
        .then((res) => {
          ElMessageBox.alert('注册请求发送成功,请等待管理员通过......', '提示', {
            confirmButtonText: 'OK',
            callback: (action: Action) => {
              ElMessage({
                type: 'success',
                message: `感谢您对AscensionPath的支持🙏`
              })
            }
          })
          toLogin()
        })
        .catch((err) => {
          ElMessage.error('注册失败')
        })
      loading.value = false
    } catch (error) {
      console.log('验证失败', error)
    }
  }

  const toLogin = () => {
    setTimeout(() => {
      router.push('/login')
    }, 1000)
  }
</script>

<style lang="scss" scoped>
  @use '../login/index' as login;
  @use './index' as register;
</style>
