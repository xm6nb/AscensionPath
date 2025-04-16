<template>
  <div class="page-content user">
    <div class="content">
      <div class="left-wrap">
        <div class="user-wrap box-style">
          <img class="bg" src="@imgs/user/bg.png" />
          <img class="avatar" :src="RandomJpgImg(userInfo.id || Math.random() * 10)" />
          <h2 class="name">{{ userInfo.username }}</h2>
          <p class="des">AscensionPath 是一款开源的漏洞靶场平台.</p>

          <div class="outer-info">
            <div>
              <i class="iconfont-sys">&#xe72e;</i>
              <span>{{ userInfo.email }}</span>
            </div>
          </div>

          <!-- <div class="lables">
            <h3>标签</h3>
            <div>
              <div v-for="item in lableList" :key="item">
                {{ item }}
              </div>
            </div>
          </div> -->
        </div>
      </div>
      <div class="right-wrap">
        <div class="info box-style">
          <h1 class="title">基本设置</h1>

          <el-form
            :model="form"
            class="form"
            ref="ruleFormRef"
            :rules="rules"
            label-width="86px"
            label-position="top"
          >
            <el-row>
              <el-form-item label="用户名" prop="realName">
                <el-input v-model="form.username" :disabled="!isEdit" />
              </el-form-item>
              <el-form-item label="用户身份" prop="role" class="right-input" v-if="userInfo.role === 'admin'">
                <el-select v-model="form.role" placeholder="Select" :disabled="!isEdit">
                  <el-option
                    v-for="item in options"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
            </el-row>

            <el-row>
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="form.email" :disabled="!isEdit" />
              </el-form-item>
            </el-row>

            <el-row>
              <el-form-item label="积分" prop="score" v-if="userInfo.role === 'admin'">
                <el-input-number
                  v-model="form.score"
                  class="mx-4"
                  :disabled="!isEdit"
                  :min="0"
                  :max="1000"
                  style="margin-left: 0"
                />
              </el-form-item>
            </el-row>

            <div class="el-form-item-right">
              <el-button type="primary" style="width: 90px" v-ripple @click="edit">
                {{ isEdit ? '保存' : '编辑' }}
              </el-button>
            </div>
          </el-form>
        </div>

        <div class="info box-style" style="margin-top: 20px">
          <h1 class="title">更改密码</h1>

          <el-form :model="pwdForm" class="form" label-width="86px" label-position="top">
            <el-form-item label="当前密码" prop="password">
              <el-input v-model="pwdForm.password" type="password" :disabled="!isEditPwd" />
            </el-form-item>

            <el-form-item label="新密码" prop="newPassword">
              <el-input v-model="pwdForm.newPassword" type="password" :disabled="!isEditPwd" />
            </el-form-item>

            <el-form-item label="确认新密码" prop="confirmPassword">
              <el-input v-model="pwdForm.confirmPassword" type="password" :disabled="!isEditPwd" />
            </el-form-item>

            <div class="el-form-item-right">
              <el-button type="primary" style="width: 90px" v-ripple @click="editPwd">
                {{ isEditPwd ? '保存' : '编辑' }}
              </el-button>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useUserStore } from '@/store/modules/user'
  import { FormInstance, FormRules } from 'element-plus'
  import { RandomJpgImg } from '@/utils/utils'
  import api from '@/utils/http'
  import { BaseResult } from '@/types/axios'

  const userStore = useUserStore()
  const userInfo = computed(() => userStore.getUserInfo)

  const isEdit = ref(false)
  const isEditPwd = ref(false)
  const date = ref('')

  let form = reactive({
    id: userInfo.value.id,
    username: userInfo.value.username,
    email: userInfo.value.email,
    role: userInfo.value.role,
    score: userInfo.value.score
  })

  const init = () => {
    api.get<BaseResult>({
        url: `/api/v1/users/getUserInfo`,
        params: {
          id: userInfo.value.id
        }
      })
      .then((res) => {
        const user = res.data
        form.id = user.id
        form.username = user.username
        form.email = user.email
        form.role = user.role
        form.score = user.score
      })
      .catch((err) => {})
  }

  onMounted(() => {
    init()
    getDate()
  })

  const pwdForm = reactive({
    password: '',
    newPassword: '',
    confirmPassword: ''
  })

  const ruleFormRef = ref<FormInstance>()

  const rules = reactive<FormRules>({
    username: [
      { required: true, message: '请输入昵称', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 30 个字符', trigger: 'blur' }
    ],
    email: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
    role: [{ type: 'array', required: true, message: '请选择用户身份', trigger: 'blur' }],
    score: [
      { required: true, message: '请输入积分', trigger: 'blur' },
      { min: 0, max: 10000, message: '积分范围在 0 到 10000 之间', trigger: 'blur' }
    ]
  })

  const options = [
    {
      value: 'admin',
      label: '管理员'
    },
    {
      value: 'user',
      label: '普通用户'
    },
    {
      value: 'vip',
      label: 'VIP用户'
    }
  ]

  const getDate = () => {
    const d = new Date()
    const h = d.getHours()
    let text = ''

    if (h >= 6 && h < 9) {
      text = '早上好'
    } else if (h >= 9 && h < 11) {
      text = '上午好'
    } else if (h >= 11 && h < 13) {
      text = '中午好'
    } else if (h >= 13 && h < 18) {
      text = '下午好'
    } else if (h >= 18 && h < 24) {
      text = '晚上好'
    } else if (h >= 0 && h < 6) {
      text = '很晚了，早点睡'
    }

    date.value = text
  }

  const edit = () => {
    isEdit.value = !isEdit.value
    if (!isEdit.value) {
      if (form.username === userInfo.value.username && form.email === userInfo.value.email && form.role === userInfo.value.role && form.score === userInfo.value.score) {
        ElMessage.error('没有修改任何信息')
        return
      }
      // 保存用户信息
      api
        .post<BaseResult>({
          url: `/api/v1/users/profile`,
          data: {
            code: 200,
            message: '更新用户数据',
            data: {
              id: form.id,
              email: form.email,
              role: form.role,
              username: form.username,
              score: form.score,
              status: 1
            }
          }
        })
        .then((res) => {
          ElNotification({
            title: '提示',
            message: '更新用户信息成功',
            type: 'success'
          })
          init()
        })
        .catch((error) => {
          ElMessage.error('更新失败:' + error.message)
        })
    }
  }

  const editPwd = () => {
    isEditPwd.value = !isEditPwd.value
    if (!isEditPwd.value) {
      if (pwdForm.password === '' || pwdForm.newPassword === '' || pwdForm.confirmPassword === '') {
        pwdForm.password = ''
        pwdForm.newPassword = ''
        pwdForm.confirmPassword = ''
        ElMessage.error('请填写完整的密码信息')
        return
      }
      if (pwdForm.newPassword !== pwdForm.confirmPassword) {
        pwdForm.password = ''
        pwdForm.newPassword = ''
        pwdForm.confirmPassword = ''
        ElMessage.error('两次输入的密码不一致')
        return
      }
      api
        .post<BaseResult>({
          url: `/api/v1/users/updatePassword`,
          data: {
            code: 200,
            message: '更新用户密码',
            data: {
              user_id: userInfo.value.id,
              old_password: pwdForm.password,
              new_password: pwdForm.newPassword
            }
          }
        })
        .then((res) => {
          ElNotification({
            title: '提示',
            message: '修改成功',
            type: 'success'
          })
        })
        .catch((error) => {
          ElMessage.error('更新失败:' + error.message)
        })
    }
  }
</script>

<style lang="scss">
  .user {
    .icon {
      width: 1.4em;
      height: 1.4em;
      overflow: hidden;
      vertical-align: -0.15em;
      fill: currentcolor;
    }
  }
</style>

<style lang="scss" scoped>
  .page-content {
    width: 100%;
    height: 100%;
    padding: 0 !important;
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;

    $box-radius: calc(var(--custom-radius) + 4px);

    .box-style {
      border: 1px solid var(--art-border-color);
    }

    .content {
      position: relative;
      display: flex;
      justify-content: space-between;
      margin-top: 10px;

      .left-wrap {
        width: 450px;
        margin-right: 25px;

        .user-wrap {
          position: relative;
          height: 600px;
          padding: 35px 40px;
          overflow: hidden;
          text-align: center;
          background: var(--art-main-bg-color);
          border-radius: $box-radius;

          .bg {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 200px;
            object-fit: cover;
          }

          .avatar {
            position: relative;
            z-index: 10;
            width: 80px;
            height: 80px;
            margin-top: 120px;
            object-fit: cover;
            border: 2px solid #fff;
            border-radius: 50%;
          }

          .name {
            margin-top: 20px;
            font-size: 22px;
            font-weight: 400;
          }

          .des {
            margin-top: 20px;
            font-size: 14px;
          }

          .outer-info {
            width: 300px;
            margin: auto;
            margin-top: 30px;
            text-align: left;

            > div {
              margin-top: 10px;

              span {
                margin-left: 8px;
                font-size: 14px;
              }
            }
          }

          .lables {
            margin-top: 40px;

            h3 {
              font-size: 15px;
              font-weight: 500;
            }

            > div {
              display: flex;
              flex-wrap: wrap;
              justify-content: center;
              margin-top: 15px;

              > div {
                padding: 3px 6px;
                margin: 0 10px 10px 0;
                font-size: 12px;
                background: var(--art-main-bg-color);
                border: 1px solid var(--art-border-color);
                border-radius: 2px;
              }
            }
          }
        }

        .gallery {
          margin-top: 25px;
          border-radius: 10px;

          .item {
            img {
              width: 100%;
              height: 100%;
              object-fit: cover;
            }
          }
        }
      }

      .right-wrap {
        flex: 1;
        overflow: hidden;
        border-radius: $box-radius;

        .info {
          background: var(--art-main-bg-color);
          border-radius: $box-radius;

          .title {
            padding: 15px 25px;
            font-size: 20px;
            font-weight: 400;
            color: var(--art-text-gray-800);
            border-bottom: 1px solid var(--art-border-color);
          }

          .form {
            box-sizing: border-box;
            padding: 30px 25px;

            > .el-row {
              .el-form-item {
                width: calc(50% - 10px);
              }

              .el-input,
              .el-select {
                width: 100%;
              }
            }

            .right-input {
              margin-left: 20px;
            }

            .el-form-item-right {
              display: flex;
              align-items: center;
              justify-content: end;

              .el-button {
                width: 110px !important;
              }
            }
          }
        }
      }
    }
  }

  @media only screen and (max-width: $device-ipad-vertical) {
    .page-content {
      .content {
        display: block;
        margin-top: 5px;

        .left-wrap {
          width: 100%;
        }

        .right-wrap {
          width: 100%;
          margin-top: 15px;
        }
      }
    }
  }
</style>
