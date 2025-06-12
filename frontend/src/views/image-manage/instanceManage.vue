<template>
  <div class="image-list">
    <table-bar
      :showTop="false"
      :showSetting="showSetting"
      @search="search"
      @reset="resetForm(searchFormRef)"
      @changeColumn="changeColumn"
      :columns="columns"
    >
      <template #top>
        <el-form :model="searchForm" ref="searchFormRef" label-width="82px">
          <el-row :gutter="20">
            <form-input label="用户名" prop="name" v-model="searchForm.username" />
            <form-input label="漏洞类型" prop="vulType" v-model="searchForm.vulType" />
            <form-input label="开发语言" prop="devLanguage" v-model="searchForm.devLanguage" />
          </el-row>
          <el-row :gutter="20">
            <form-input label="场景名称" prop="name" v-model="searchForm.envName" />
            <form-input label="数据库" prop="devDatabase" v-model="searchForm.devDatabase" />
            <form-input label="开发框架" prop="devMiddleware" v-model="searchForm.devMiddleware" />
          </el-row>
        </el-form>
      </template>
    </table-bar>

    <el-table
      :data="tableData"
      selection
      v-loading="loading"
      element-loading-text="AscensionPath正在处理请求中......"
      :element-loading-spinner="svg"
      element-loading-svg-view-box="-10, -10, 50, 50"
      element-loading-background="rgba(122, 122, 122, 0.8)"
    >
      <el-table-column fixed label="实例信息" #default="scope" width="250px" v-if="columns[0].show">
        <el-tooltip placement="top" effect="light">
          <template #content>
            <div style="max-width: 480px"
              ><span style="font-size: small">实例信息</span>
              <hr />
              <p style="font-size: 15px">
                开启用户:{{ scope.row.username }}<br />
                场景名称:{{ scope.row.env_name }}
              </p>
            </div>
          </template>
          <div class="info" style="display: flex; align-items: center">
            <div>
              <p class="image-vul-name">{{ scope.row.username }}</p>
              <p class="email">{{ scope.row.env_name }}</p>
            </div>
          </div>
        </el-tooltip>
      </el-table-column>
      <el-table-column label="评分" prop="rank" width="75px" v-if="columns[1].show">
      </el-table-column>
      <el-table-column
        label="镜像描述"
        prop="image_desc"
        #default="scope"
        width="120px"
        v-if="columns[2].show"
      >
        <el-tooltip placement="top" effect="light">
          <template #content>
            <div style="max-width: 480px"
              ><span style="font-size: small">镜像描述</span>
              <hr />
              <p style="font-size: 15px">{{ scope.row.env_desc }}</p>
            </div>
          </template>
          {{
            scope.row.env_desc?.length > 7
              ? scope.row.env_desc.substring(0, 7) + '...'
              : scope.row.env_desc
          }}
        </el-tooltip>
      </el-table-column>
      <el-table-column
        label="开启时间"
        prop="start_time"
        v-if="columns[3].show"
        #default="scope"
        width="230px"
      >
        {{ formatDate(scope.row.start_time) }}
      </el-table-column>
      <el-table-column
        label="过期时间"
        prop="expire_time"
        v-if="columns[4].show"
        #default="scope"
        width="230px"
      >
        {{ formatDate(scope.row.expire_time) }}
      </el-table-column>
      <el-table-column label="场景开销" prop="cost" v-if="columns[5].show" width="100px">
      </el-table-column>
      <el-table-column #default="scope" label="端口映射" v-if="columns[6].show" width="100px">
        <el-tooltip placement="top" effect="light">
          <template #content>
            <div style="max-width: 480px"
              ><span style="font-size: small">映射地址</span>
              <hr />
              <el-link
                type="primary"
                v-for="link in convertPortToLink(scope.row.ports)"
                :href="link"
                target="_blank"
                style="margin-right: 10px"
                >{{ link }}</el-link
              >
            </div>
          </template>
          {{ convertPortToLink(scope.row.ports)[0].substring(0, 7) + '...' }}
        </el-tooltip>
      </el-table-column>
      <el-table-column label="漏洞类型" v-if="columns[7].show" width="110px">
        <template #default="scope">
          <el-tag
            v-for="(type, index) in scope.row.degree.HoleType"
            :key="index"
            type="info"
            style="margin-bottom: 5px"
          >
            {{ type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开发语言" v-if="columns[8].show" width="100px">
        <template #default="scope">
          <el-tag
            v-for="(type, index) in scope.row.degree.devLanguage"
            :key="index"
            type="info"
            style="margin-bottom: 5px"
          >
            {{ type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开发数据库" v-if="columns[9].show" width="110px">
        <template #default="scope">
          <el-tag
            v-for="(type, index) in scope.row.degree.devDatabase"
            :key="index"
            type="info"
            style="margin-bottom: 5px"
          >
            {{ type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开发框架" v-if="columns[10].show" width="100px">
        <template #default="scope">
          <el-tag
            v-for="(type, index) in scope.row.degree.devClassify"
            :key="index"
            type="info"
            style="margin-bottom: 5px"
          >
            {{ type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="来源" prop="from" sortable v-if="columns[11].show" width="100px" />
      <el-table-column fixed="right" label="操作" width="150px">
        <template #default="scope">
          <el-tooltip class="box-item" effect="dark" content="延长半小时" placement="bottom">
            <button-table type="edit" @click="extendTime(scope.row.id)" />
          </el-tooltip>
          <el-tooltip class="box-item" effect="dark" content="移除实例" placement="bottom">
            <button-table
              type="delete"
              @click="removeInstance(scope.row.user_id, scope.row.vul_env_id)"
            />
          </el-tooltip>
        </template>
      </el-table-column>
    </el-table>
    <div style="display: flex; justify-content: center"
      ><el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { FormInstance } from 'element-plus'
  import api from '@/utils/http'
  import { BaseResult } from '@/types/axios'
  import { formatDate } from '@/utils/utils'

  const loading = ref(false)
  const svg = `
        <path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>`
  const showSetting = ref(false)

  const columns = reactive([
    { name: '实例信息', show: true },
    { name: '评分', show: true },
    { name: '镜像描述', show: true },
    { name: '开启时间', show: true },
    { name: '过期时间', show: true },
    { name: '场景开销', show: true },
    { name: '端口映射', show: true },
    { name: '漏洞类型', show: true },
    { name: '开发语言', show: true },
    { name: '开发数据库', show: true },
    { name: '开发框架', show: true },
    { name: '来源', show: true }
  ])

  const searchFormRef = ref<FormInstance>()
  interface SearchForm {
    username: string
    envName: string
    vulType: string
    devLanguage: string
    devDatabase: string
    devMiddleware: string
  }
  const searchForm = reactive<SearchForm>({
    username: '',
    envName: '',
    vulType: '',
    devLanguage: '',
    devDatabase: '',
    devMiddleware: ''
  })

  const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return
    formEl.resetFields()
  }
  let conditionalArrays: any = ref([]) // 存储条件镜像数据
  let imageData: any = ref([]) // 存储所有镜像数据
  let tableData: any = ref([]) // 存储表格数据

  // 添加分页相关变量
  const currentPage = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  // 处理页码变化
  const handlePageChange = (page: number) => {
    currentPage.value = page
    tableData.value = conditionalArrays.value.slice(
      (currentPage.value - 1) * pageSize.value,
      currentPage.value * pageSize.value
    )
  }

  // 处理每页条数变化
  const handleSizeChange = (size: number) => {
    pageSize.value = size
    tableData.value = conditionalArrays.value.slice(
      (currentPage.value - 1) * pageSize.value,
      currentPage.value * pageSize.value
    )
  }

  // 更新表格数据
  const updateTableData = () => {
    loading.value = true
    if (
      searchForm.username ||
      searchForm.vulType ||
      searchForm.devLanguage ||
      searchForm.devDatabase ||
      searchForm.devMiddleware
    ) {
      search()
    } else {
      conditionalArrays.value = imageData.value
      total.value = imageData.value.length
      handlePageChange(1)
    }
    loading.value = false
  }

  onMounted(() => {
    init()
  })

  // 初始化时获取用户列表
  const init = () => {
    getInstances().then((data) => {
      console.log(data)
      imageData.value = data
      updateTableData()
    })
  }

  const getInstances = () => {
    // 模拟获取用户列表的 API 请求
    return new Promise((resolve) => {
      api
        .get<BaseResult>({
          url: `/api/v1/vul/getAllInstance`
        })
        .then((res) => {
          resolve(res.data)
        })
        .catch((error) => {})
    })
  }

  const search = () => {
    conditionalArrays.value = []
    tableData.value = []
    for (const index in imageData.value) {
      const shouldInclude =
        (searchForm.envName === '' ||
          imageData.value[index]?.env_name?.includes(searchForm.envName)) &&
        (searchForm.username === '' ||
          imageData.value[index]?.username?.includes(searchForm.username)) &&
        (searchForm.vulType === '' ||
          (imageData.value[index]?.degree?.HoleType || []).some((item: string) =>
            item.includes(searchForm.vulType)
          )) &&
        (searchForm.devLanguage === '' ||
          (imageData.value[index]?.degree?.devLanguage || []).some((item: string) =>
            item.includes(searchForm.devLanguage)
          )) &&
        (searchForm.devDatabase === '' ||
          (imageData.value[index]?.degree?.devDatabase || []).some((item: string) =>
            item.includes(searchForm.devDatabase)
          )) &&
        (searchForm.devMiddleware === '' ||
          (imageData.value[index]?.degree?.devClassify || []).some((item: string) =>
            item.includes(searchForm.devMiddleware)
          ))
      if (shouldInclude) {
        conditionalArrays.value.push(imageData.value[index])
      }
    }
    total.value = conditionalArrays.value.length
    handlePageChange(1)
  }

  const changeColumn = (list: any) => {
    columns.values = list
  }

  const extendTime = async (id: number) => {
    await api
      .get<BaseResult>({
        url: `/api/v1/vul/extendExpireTime`,
        params: {
          id: id
        }
      })
      .then((res) => {
        ElNotification({
          title: '提示',
          message: '该实例已成功延长30分钟',
          type: 'success'
        })
        init()
      })
      .catch((error) => {})
  }

  const removeInstance = async (userID: number, vulEnvID: number) => {
    loading.value = true
    await api
      .post<BaseResult>({
        url: '/api/v1/vul/removeInstance',
        data: {
          code: 200,
          message: '移除场景实例',
          data: {
            user_id: userID,
            vul_env_id: vulEnvID
          }
        }
      })
      .then((res) => {
        ElNotification({
          title: '提示',
          message: '成功移除实例场景',
          type: 'success'
        })
        init()
      })
      .catch((err) => {})
  }

  const convertPortToLink = (ports: string[]) => {
    let out = []
    if (ports && ports.length > 0) {
      const host = window.location.hostname
      try {
        // 声明类型
        interface PortMap {
          [containerPort: string]: string
        }

        // 新方案：智能合并所有JSON片段
        const portMap = ports.reduce((acc: PortMap, str: string) => {
          // 1. 移除字符串首尾的冗余字符
          const cleanedStr = str.replace(/^[\s"{}\\]*|[\s"{}\\]*$/g, '')

          // 2. 提取键值对
          const pairs = cleanedStr.split(',').filter(Boolean)

          // 3. 合并到结果对象
          pairs.forEach((pair: string) => {
            const [key, value] = pair.split(':').map((s: string) => s.replace(/^"|"$/g, '').trim())
            if (key && value) {
              acc[key] = value
            }
          })
          return acc
        }, {} as PortMap) // 初始化空对象并断言类型

        // 生成链接
        for (const [containerPort, hostPort] of Object.entries(portMap)) {
          out.push(`http://${host}:${hostPort}# 来自容器内端口:${containerPort}`)
          out.push(`https://${host}:${hostPort}# 来自容器内端口:${containerPort}`)
        }
      } catch (e) {
        console.error('解析端口失败:', e)
      }
    }
    if (out.length === 0) {
      out.push('未检测到该镜像的映射端口，故暂无访问地址')
    }
    return out
  }
</script>

<style lang="scss" scoped>
  .image-list {
    width: 100%;
    height: 100%;

    .info {
      > div {
        margin-left: 10px;

        .image-vul-name {
          font-weight: 500;
          color: var(--art-text-gray-800);
        }
      }
    }
  }

  #setting-upload {
    font-family: inherit;
    font-size: 1rem;
    background: royalblue;
    color: white;
    padding: 0.7em 1em;
    padding-left: 0.9em;
    display: flex;
    align-items: center;
    border: none;
    border-radius: 16px;
    overflow: hidden;
    transition: all 0.2s;
    cursor: pointer;
  }

  #setting-upload span {
    display: block;
    margin-left: 0.3em;
    transition: all 0.3s ease-in-out;
  }

  #setting-upload svg {
    display: block;
    transform-origin: center center;
    transition: transform 0.3s ease-in-out;
  }

  #setting-upload:hover .svg-wrapper {
    animation: fly-1 0.6s ease-in-out infinite alternate;
  }

  #setting-upload:hover svg {
    transform: translateX(2.7em) rotate(45deg) scale(1.1);
  }

  #setting-upload:hover span {
    transform: translateX(7em);
  }

  #setting-upload:active {
    transform: scale(0.95);
  }

  @keyframes fly-1 {
    from {
      transform: translateY(0.1em);
    }

    to {
      transform: translateY(-0.1em);
    }
  }

  #pull-image {
    border: none;
    display: flex;
    padding: 0.75rem 1.5rem;
    background-color: #488aec;
    color: #ffffff;
    font-size: 0.75rem;
    line-height: 1rem;
    font-weight: 700;
    text-align: center;
    cursor: pointer;
    text-transform: uppercase;
    vertical-align: middle;
    align-items: center;
    border-radius: 0.5rem;
    user-select: none;
    gap: 0.75rem;
    white-space: nowrap;
    box-shadow:
      0 4px 6px -1px #488aec31,
      0 2px 4px -1px #488aec17;
    transition: all 0.6s ease;
  }

  #pull-image:hover {
    box-shadow:
      0 10px 15px -3px #488aec4f,
      0 4px 6px -2px #488aec17;
  }

  #pull-image:focus,
  #pull-image:active {
    opacity: 0.85;
    box-shadow: none;
  }

  #pull-image svg {
    width: 1.25rem;
    height: 1.25rem;
  }
</style>
