<template>
  <div class="login">
    <div class="left-wrap">
      <left-view></left-view>
    </div>
    <div class="right-wrap">
      <div class="top-right-wrap">
        <div class="btn theme-btn" @click="toggleTheme">
          <i class="iconfont-sys">
            {{ isDark ? '&#xe6b5;' : '&#xe725;' }}
          </i>
        </div>
        <el-dropdown @command="changeLanguage" popper-class="langDropDownStyle">
          <div class="btn language-btn">
            <i class="iconfont-sys icon-language">&#xe611;</i>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <div v-for="lang in languageOptions" :key="lang.value" class="lang-btn-item">
                <el-dropdown-item
                  :command="lang.value"
                  :class="{ 'is-selected': locale === lang.value }"
                >
                  <span class="menu-txt">{{ lang.label }}</span>
                  <i v-if="locale === lang.value" class="iconfont-sys icon-check">&#xe621;</i>
                </el-dropdown-item>
              </div>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="header">
        <svg class="icon" aria-hidden="true" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg" fill="none"><path fill="#000000" fill-rule="evenodd" d="M13 0a3 3 0 00-1.65 5.506 7.338 7.338 0 01-.78 1.493c-.22.32-.472.635-.8 1.025a1.509 1.509 0 00-.832.085 12.722 12.722 0 00-1.773-1.124c-.66-.34-1.366-.616-2.215-.871a1.5 1.5 0 10-2.708 1.204c-.9 1.935-1.236 3.607-1.409 5.838a1.5 1.5 0 101.497.095c.162-2.07.464-3.55 1.25-5.253.381-.02.725-.183.979-.435.763.23 1.367.471 1.919.756a11.13 11.13 0 011.536.973 1.5 1.5 0 102.899-.296c.348-.415.64-.779.894-1.148.375-.548.665-1.103.964-1.857A3 3 0 1013 0zm-1.5 3a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0z" clip-rule="evenodd"/></svg>

        <h1>{{ systemName }}</h1>
      </div>
      <div class="login-wrap">
        <div class="form">
          <h3 class="title">{{ $t('login.title') }}</h3>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>
          <el-form
            ref="formRef"
            :model="formData"
            :rules="rules"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <el-form-item prop="username">
              <el-input
                :placeholder="$t('login.placeholder[0]')"
                size="large"
                v-model.trim="formData.username"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                :placeholder="$t('login.placeholder[1]')"
                size="large"
                v-model.trim="formData.password"
                type="password"
                radius="8px"
                autocomplete="off"
                show-password
              />
            </el-form-item>
            <div class="drag-verify">
              <div class="drag-verify-content" :class="{ error: !isPassing && isClickPass }">
                <!-- :background="isDark ? '#181818' : '#eee'" -->
                <DragVerify
                  ref="dragVerify"
                  v-model:value="isPassing"
                  :width="width < 500 ? 328 : 438"
                  :text="$t('login.sliderText')"
                  textColor="var(--art-gray-800)"
                  :successText="$t('login.sliderSuccessText')"
                  :progressBarBg="getCssVariable('--el-color-primary')"
                  background="var(--art-gray-200)"
                  handlerBg="var(--art-main-bg-color)"
                  @pass="onPass"
                />
              </div>
              <p class="error-text" :class="{ 'show-error-text': !isPassing && isClickPass }">{{
                $t('login.placeholder[2]')
              }}</p>
            </div>

            <div class="forget-password">
              <el-checkbox v-model="formData.rememberPassword">{{
                $t('login.rememberPwd')
              }}</el-checkbox>
              <!-- <router-link to="/forget-password">{{ $t('login.forgetPwd') }}</router-link> -->
            </div>

            <div style="margin-top: 30px">
              <el-button
                class="login-btn"
                size="large"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </el-button>
            </div>

            <div class="footer">
              <p>
                {{ $t('login.noAccount') }}
                <router-link to="/register">{{ $t('login.register') }}</router-link>
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
  import { ElMessage, ElNotification } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { HOME_PAGE } from '@/router'
  import { ApiStatus } from '@/utils/http/status'
  import { getCssVariable } from '@/utils/colors'
  import { languageOptions } from '@/language'
  import { LanguageEnum, SystemThemeEnum } from '@/enums/appEnum'
  import { useI18n } from 'vue-i18n'
  import  Avatar from '@imgs/user/avatar.png'

  const { t } = useI18n()
  import { useSettingStore } from '@/store/modules/setting'
  import type { FormInstance, FormRules } from 'element-plus'

  const userStore = useUserStore()
  const router = useRouter()
  const isPassing = ref(false)
  const isClickPass = ref(false)

  const systemName = AppConfig.systemInfo.name
  const formRef = ref<FormInstance>()
  const formData = reactive({
    username: AppConfig.systemInfo.login.username,
    password: AppConfig.systemInfo.login.password,
    rememberPassword: true
  })

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder[0]'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder[1]'), trigger: 'blur' }]
  }))

  const loading = ref(false)
  const { width } = useWindowSize()

  const store = useSettingStore()
  const isDark = computed(() => store.isDark)

  const onPass = () => {}

  const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate(async (valid) => {
      if (valid) {
        if (!isPassing.value) {
          isClickPass.value = true
          return
        }

        loading.value = true

        // 延时辅助函数
        const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

        try {
          const res = await UserService.login({
            body: {
              code: 200,
              message: '请求登录',
              data: {
                username: formData.username,
                password: formData.password
              }
            }
          })

          if (res.code === ApiStatus.success && res.data) {
            // 设置 token
            userStore.setToken(res.data.token)

            // 获取用户信息
            userStore.setUserInfo({
              id: res.data.id,
              username: res.data.username,
              avatar: Avatar,
              email: res.data.email,
              role: res.data.role,
              score: res.data.score,
            })
            // 设置登录状态
            userStore.setLoginStatus(true)
            // 延时辅助函数
            await delay(1000)
            // 登录成功提示
            showLoginSuccessNotice()
            // 跳转首页
            if(res.data.role === 'admin'){
              router.push(HOME_PAGE)
            }else{
              router.push(RoutesAlias.CreateInstance)
            }
          } else {
            ElMessage.error(res.message)
          }
        } finally {
          await delay(1000)
          loading.value = false
        }
      }
    })
  }

  // 登录成功提示
  const showLoginSuccessNotice = () => {
    setTimeout(() => {
      ElNotification({
        title: t('login.success.title'),
        type: 'success',
        showClose: false,
        duration: 2500,
        zIndex: 10000,
        message: `${t('login.success.message')}, ${systemName}!`
      })
    }, 300)
  }

  // 切换语言
  const { locale } = useI18n()

  const changeLanguage = (lang: LanguageEnum) => {
    if (locale.value === lang) return
    locale.value = lang
    userStore.setLanguage(lang)
  }

  // 切换主题
  import { useTheme } from '@/composables/useTheme'
  import { UserService } from '@/api/usersApi'
import { RoutesAlias } from '@/router/modules/routesAlias'

  const toggleTheme = () => {
    let { LIGHT, DARK } = SystemThemeEnum
    useTheme().switchThemeStyles(useSettingStore().systemThemeType === LIGHT ? DARK : LIGHT)
  }
</script>

<style lang="scss" scoped>
  @use './index';
</style>
