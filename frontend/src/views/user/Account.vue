<template>
  <div class="page-content">
    <table-bar
      :showTop="false"
      @search="search"
      @reset="resetForm(searchFormRef)"
      @changeColumn="changeColumn"
      :columns="columns"
    >
      <template #top>
        <el-form :model="searchForm" ref="searchFormRef" label-width="82px">
          <el-row :gutter="20">
            <form-input label="用户名" prop="name" v-model="searchForm.name" />
            <form-input label="邮箱" prop="email" v-model="searchForm.email" />
          </el-row>
          <el-row :gutter="20">
            <form-select
              label="状态"
              prop="status"
              v-model="searchForm.status"
              :options="statusOptions"
            />
            <form-select
              label="用户身份"
              prop="role"
              v-model="searchForm.role"
              :options="roleOptions"
            />
          </el-row>
        </el-form>
      </template>
      <template #bottom>
        <el-button @click="showDialog('add')" v-ripple>添加用户</el-button>
      </template>
    </table-bar>

    <art-table
      :data="tableData"
      selection
      :currentPage="currentPage"
      :pageSize="pageSize"
      :total="total"
      @current-change="handlePageChange"
      @size-change="handleSizeChange"
    >
      <template #default>
        <el-table-column
          label="用户名"
          prop="avatar"
          #default="scope"
          width="250px"
          v-if="columns[0].show"
        >
          <div class="user" style="display: flex; align-items: center">
            <img class="avatar" :src="RandomJpgImg(scope.row.id || Math.random())" />
            <div>
              <p class="user-name">{{ scope.row.username }}</p>
              <p class="email">{{ scope.row.email }}</p>
            </div>
          </div>
        </el-table-column>
        <el-table-column label="ID" prop="id" width="50px" v-if="columns[1].show">
        </el-table-column>
        <el-table-column
          label="用户身份"
          prop="role"
          #default="scope"
          width="120px"
          sortable
          v-if="columns[2].show"
        >
          {{
            scope.row.role === 'admin'
              ? '管理员'
              : scope.row.role === 'user'
                ? '普通用户'
                : scope.row.role === 'vip'
                  ? 'VIP用户'
                  : '未知身份'
          }}
        </el-table-column>
        <el-table-column label="额度" prop="score" v-if="columns[3].show" />
        <el-table-column
          label="状态"
          prop="status"
          v-if="columns[4].show"
        >
          <template #default="scope">
            <el-tag :type="getTagType(scope.row.status)">
              {{ buildTagText(scope.row.status) }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column label="最后登录时间" prop="last_login" sortable v-if="columns[5].show" width="230px">
          <template #default="scope">
            {{ formatDate(scope.row.last_login) }}
          </template>
        </el-table-column>
        <el-table-column label="创建日期" prop="created_at" sortable v-if="columns[6].show" width="230px">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="近期活动时间" prop="updated_at" sortable v-if="columns[7].show" width="230px">
          <template #default="scope">
            {{ formatDate(scope.row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="150px">
          <template #default="scope">
            <button-table type="edit" @click="showDialog('edit', scope.row)" />
            <button-table type="delete" @click="deleteUser(scope.row.id)" />
          </template>
        </el-table-column>
      </template>
    </art-table>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加用户' : '编辑用户'"
      width="30%"
    >
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item label="用户身份" prop="role">
          <el-select v-model="formData.role">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
            <el-option label="VIP用户" value="vip" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="formData.status">
            <el-option label="可用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="formData.password" show-password />
        </el-form-item>
        <el-form-item label="额度" prop="score">
          <el-input-number v-model="formData.score" :min="0" :max="100000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">提交</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { FormInstance } from 'element-plus'
  import { ElMessageBox, ElMessage } from 'element-plus'
  import type { FormRules } from 'element-plus'
  import api from '@/utils/http'
  import { BaseResult } from '@/types/axios'
  import { formatDate } from '@/utils/utils'
  import { RandomJpgImg } from '@/utils/utils'

  const dialogType = ref('add')
  const dialogVisible = ref(false)

  const formData = reactive({
    id: -1,
    username: '',
    email: '',
    status: 0,
    role: '',
    score: 0,
    password: ''
  })
  const statusOptions = [
    {
      value: 1,
      label: '可用'
    },
    {
      value: 0,
      label: '禁用'
    },
    {
      value: -1,
      label: '所有'
    }
  ]

  const roleOptions = [
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

  const columns = reactive([
    { name: '用户名', show: true },
    { name: 'ID', show: true },
    { name: '用户身份', show: true },
    { name: '额度', show: true },
    { name: '状态', show: true },
    { name: '最后登录时间', show: true },
    { name: '创建日期', show: true },
    { name: '近期活动时间', show: true }
  ])

  const searchFormRef = ref<FormInstance>()
  interface SearchForm {
    name: string
    email: string
    status: number
    role: string
  }
  const searchForm = reactive<SearchForm>({
    name: '',
    email: '',
    status: -1,
    role: ''
  })

  const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return
    formEl.resetFields()
  }
  let tableData: any = ref([])

  // 添加分页相关变量
  const currentPage = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  // 处理页码变化
  const handlePageChange = (page: number) => {
    currentPage.value = page
    updateTableData()
  }

  // 处理每页条数变化
  const handleSizeChange = (size: number) => {
    pageSize.value = size
    updateTableData()
  }
  // 初始化表格数据
  async function updateTableData() {
    if (searchForm.name || searchForm.email || searchForm.status !== -1 || searchForm.role) {
      search()
    } else {
      await getUsers(currentPage.value, pageSize.value).then((data) => {
        tableData.value = (data as { users: any[]; count: number }).users
        total.value = (data as { users: any[]; count: number }).count
      })
    }
  }
  updateTableData()
  function getUsers(page: number, pagesize: number) {
    // 模拟获取用户列表的 API 请求
    return new Promise((resolve) => {
      api
        .get<BaseResult>({
          url: `/api/v1/users/getAllUsers`,
          params: {
            page: page,
            pagesize: pagesize
          }
        })
        .then((res) => {
          resolve(res.data)
        })
    })
  }

  const showDialog = (type: string, row?: any) => {
    dialogVisible.value = true
    dialogType.value = type

    if (type === 'edit' && row) {
      formData.id = row.id
      formData.username = row.username
      formData.email = row.email
      formData.role = row.role
      formData.score = row.score
      formData.status = row.status
      formData.password = ''
    } else {
    }
  }

  const deleteUser = (id: number) => {
    ElMessageBox.confirm('确定要注销该用户吗？', '注销用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    }).then(() => {
      api
        .post<BaseResult>({
          url: `/api/v1/users/deleteUser`,
          data: {
            code: 200,
            message: '删除用户',
            data: {
              id: id
            }
          }
        })
        .then((res) => {
          if (res.code === 200) {
            ElMessage.success('注销成功')
            updateTableData()
          }
        })
        .catch((error) => {
          ElMessage.error('注销失败:' + error.message)
        })
    })
  }

  const search = () => {
    api
      .post<BaseResult>({
        url: '/api/v1/users/searchUsers',
        data: {
          code: 200,
          message: '搜索用户',
          data: {
            username: searchForm.name,
            email: searchForm.email,
            status: searchForm.status,
            role: searchForm.role,
            page: currentPage.value,
            pagesize: pageSize.value
          }
        }
      })
      .then((res) => {
        if (res.code === 200) {
          tableData.value = res.data.users
          total.value = res.data.count
        }
      })
      .catch((error) => {
        ElMessage.error('搜索失败:' + error.message)
      })
  }

  const changeColumn = (list: any) => {
    columns.values = list
  }

  const getTagType = (status: string) => {
    switch (status) {
      case '1':
        return 'success'
      case '2':
        return 'info'
      case '3':
        return 'warning'
      case '4':
        return 'danger'
      default:
        return 'info'
    }
  }

  const buildTagText = (status: number) => {
    let text = ''
    if (status === 0) {
      text = '禁用'
    } else if (status === 1) {
      text = '可用'
    }
    return text
  }

  const rules = reactive<FormRules>({
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    role: [{ required: true, message: '请选择用户身份', trigger: 'change' }],
    email: [
      { required: true, message: '请输入邮箱地址', trigger: 'blur' },
      {
        type: 'email',
        message: '请输入正确的邮箱格式',
        trigger: ['blur', 'change']
      }
    ]
  })

  const formRef = ref<FormInstance>()

  const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate((valid) => {
      if (valid) {
        // 提交表单数据
        if (dialogType.value === 'edit') {
          api
            .post<BaseResult>({
              url: `/api/v1/users/profile`,
              data: {
                code: 200,
                message: '更新用户数据',
                data: {
                  id: formData.id,
                  email: formData.email,
                  status: formData.status,
                  role: formData.role,
                  username: formData.username,
                  password: formData.password,
                  score: formData.score
                }
              }
            })
            .then((res) => {
              if (res.code === 200) {
                updateTableData()
                ElMessage.success('更新成功')
                dialogVisible.value = false
              }
            })
            .catch((error) => {
              ElMessage.error('更新失败:' + error.message)
            })
        } else {
          api
            .post<BaseResult>({
              url: `/api/v1/users/addUser`,
              data: {
                code: 200,
                message: '添加用户数据',
                data: {
                  email: formData.email,
                  status: formData.status,
                  role: formData.role,
                  username: formData.username,
                  password: formData.password,
                  score: formData.score
                }
              }
            })
            .then((res) => {
              if (res.code === 200) {
                updateTableData()
                ElMessage.success('添加成功')
                dialogVisible.value = false
              }
            })
            .catch((error) => {
              ElMessage.error('添加失败:', error.message)
            })
        }
      }
    })
  }
</script>

<style lang="scss" scoped>
  .page-content {
    width: 100%;
    height: 100%;

    .user {
      .avatar {
        width: 40px;
        height: 40px;
        border-radius: 6px;
      }

      > div {
        margin-left: 10px;

        .user-name {
          font-weight: 500;
          color: var(--art-text-gray-800);
        }
      }
    }
  }
</style>
