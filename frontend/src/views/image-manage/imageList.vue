<template>
  <div class="image-list">
    <table-bar :showTop="false" :showSetting="showSetting" @search="search" @reset="resetForm(searchFormRef)"
      @changeColumn="changeColumn" :columns="columns">
      <template #top>
        <el-form :model="searchForm" ref="searchFormRef" label-width="82px">
          <el-row :gutter="20">
            <form-input label="镜像名称" prop="name" v-model="searchForm.name" />
            <form-input label="漏洞类型" prop="vulType" v-model="searchForm.vulType" />
            <form-input label="开发语言" prop="devLanguage" v-model="searchForm.devLanguage" />
          </el-row>
          <el-row :gutter="20">
            <form-input label="数据库" prop="devDatabase" v-model="searchForm.devDatabase" />
            <form-input label="开发框架" prop="devMiddleware" v-model="searchForm.devMiddleware" />
          </el-row>
        </el-form>
      </template>
      <template #setting>
        <div style="display: flex; flex-direction: column">
          <el-descriptions title="镜像配置" :column="3" border>
            <el-descriptions-item label="Docker状态" label-align="right" align="center">
              <el-tag size="small">{{ settingConfig.docker_health ? '健康' : '异常' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="本地镜像加载路径" label-align="right" align="center">
              <el-popover class="box-item" title="提示" content="系统会加载该路径下所有的json文件作为镜像列表，请确保其格式正确。" placement="bottom"
                width="400">
                <template #reference>
                  {{ settingConfig.path }}
                </template>
              </el-popover>
            </el-descriptions-item>
          </el-descriptions>
          <div style="display: flex;justify-content: space-between">
            <el-input v-model="imageInput" style="width: 240px;margin-top: 5px;margin-bottom: 5px;margin-right: 10px;"
              placeholder="输入单个镜像地址" clearable />
            <div style="display: flex; justify-content: flex-end; margin-top: 5px; margin-bottom: 5px">
              <button class="SelectImage" @click="pullImage(imageInput)">
                <span><svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path d="M0 0h24v24H0z" fill="none"></path>
                    <path
                      d="M24 12l-5.657 5.657-1.414-1.414L21.172 12l-4.243-4.243 1.414-1.414L24 12zM2.828 12l4.243 4.243-1.414 1.414L0 12l5.657-5.657L7.07 7.757 2.828 12zm6.96 9H7.66l6.552-18h2.128L9.788 21z"
                      fill="currentColor"></path>
                  </svg>
                  拉取指定镜像</span>
              </button>
              <button id="setting-upload" @click="submitUpload">
                <div class="svg-wrapper-1">
                  <div class="svg-wrapper">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
                      <path fill="none" d="M0 0h24v24H0z"></path>
                      <path fill="currentColor"
                        d="M1.946 9.315c-.522-.174-.527-.455.01-.634l19.087-6.362c.529-.176.832.12.684.638l-5.454 19.086c-.15.529-.455.547-.679.045L12 14l6-8-8 6-8.054-2.685z">
                      </path>
                    </svg>
                  </div>
                </div>
                <span>上传镜像列表</span>
              </button>
            </div>
          </div>
        </div>
      </template>
      <template #button>
        <el-button @click="viewSetting"><el-icon>
            <SetUp />
          </el-icon></el-button>
      </template>
    </table-bar>

    <art-table v-loading="loading" :element-loading-svg="loadSvg" class="custom-loading-svg"
      element-loading-svg-view-box="-10, -10, 50, 50" element-loading-text="服务器正在处理中......" :data="tableData" selection
      :currentPage="currentPage" :pageSize="pageSize" :total="total" @current-change="handlePageChange"
      @size-change="handleSizeChange">
      <template #default>
        <el-table-column label="镜像信息" #default="scope" width="250px" v-if="columns[0].show">
          <el-tooltip placement="top" effect="light">
            <template #content>
              <div style="max-width: 480px"><span style="font-size: small">镜像信息</span>
                <hr />
                <p style="font-size: 15px">
                  漏洞名称:{{ scope.row.image_vul_name }}<br />
                  镜像名称:{{ scope.row.image_name }}
                </p>
              </div>
            </template>
            <div class="info" style="display: flex; align-items: center">
              <div>
                <p class="image-vul-name">{{ scope.row.image_vul_name }}</p>
                <p class="email">{{ scope.row.image_name }}</p>
              </div>
            </div>
          </el-tooltip>
        </el-table-column>
        <el-table-column label="评分" prop="rank" width="75px" v-if="columns[1].show">
        </el-table-column>
        <el-table-column label="镜像描述" prop="image_desc" #default="scope" width="120px" v-if="columns[2].show">
          <el-tooltip placement="top" effect="light">
            <template #content>
              <div style="max-width: 480px"><span style="font-size: small">镜像描述</span>
                <hr />
                <p style="font-size: 15px">{{ scope.row.image_desc }}</p>
              </div>
            </template>
            {{
              scope.row.image_desc?.length > 7
                ? scope.row.image_desc.substring(0, 7) + '...'
                : scope.row.image_desc
            }}
          </el-tooltip>
        </el-table-column>
        <el-table-column label="漏洞类型" v-if="columns[3].show">
          <template #default="scope">
            <el-tag v-for="(type, index) in scope.row.degree.HoleType" :key="index" type="info"
              style="margin-bottom: 5px">
              {{ type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开发语言" v-if="columns[4].show">
          <template #default="scope">
            <el-tag v-for="(type, index) in scope.row.degree.devLanguage" :key="index" type="info"
              style="margin-bottom: 5px">
              {{ type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开发数据库" v-if="columns[5].show">
          <template #default="scope">
            <el-tag v-for="(type, index) in scope.row.degree.devDatabase" :key="index" type="info"
              style="margin-bottom: 5px">
              {{ type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开发框架" v-if="columns[6].show">
          <template #default="scope">
            <el-tag v-for="(type, index) in scope.row.degree.devClassify" :key="index" type="info"
              style="margin-bottom: 5px">
              {{ type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" prop="from" sortable v-if="columns[7].show" />
        <el-table-column fixed="right" label="操作" width="150px">
          <template #default="scope">
            <button id="pull-image" @click="pullImage(scope.row.image_name)">
              <svg aria-hidden="true" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" fill="none"
                xmlns="http://www.w3.org/2000/svg">
                <path stroke-width="2" stroke="#fffffff"
                  d="M13.5 3H12H8C6.34315 3 5 4.34315 5 6V18C5 19.6569 6.34315 21 8 21H11M13.5 3L19 8.625M13.5 3V7.625C13.5 8.17728 13.9477 8.625 14.5 8.625H19M19 8.625V11.8125"
                  stroke-linejoin="round" stroke-linecap="round"></path>
                <path stroke-linejoin="round" stroke-linecap="round" stroke-width="2" stroke="#fffffff"
                  d="M17 15V18M17 21V18M17 18H14M17 18H20"></path>
              </svg>
              拉取镜像
            </button>
          </template>
        </el-table-column>
      </template>
    </art-table>

    <el-dialog v-model="dialogVisible" :title="'镜像配置'" width="30%"> </el-dialog>
    <el-dialog v-model="progressView" title="拉取进度" width="800" :lock-scroll="true" :close-on-click-modal="false">
      <div style="display: flex; flex-direction: row-reverse">
        <button
          class="px-4 z-30 py-2 bg-rose-400 rounded-md text-white relative font-semibold after:-z-20 after:absolute after:h-1 after:w-1 after:bg-rose-800 after:left-3 overflow-hidden after:bottom-0 after:translate-y-full after:rounded-md after:hover:scale-[300] after:hover:transition-all after:hover:duration-700 after:transition-all after:duration-700 transition-all duration-700 [text-shadow:2px_3px_2px_#be123c;] hover:[text-shadow:1px_1px_2px_#fda4af] text-xl"
          @click="stopPullImage">
          终止拉取
        </button>
      </div>
      <el-table :data="progressTableData" style="width: 100%" max-height="300" v-loading="progressLoading"
        :element-loading-svg="loadSvg" class="custom-loading-svg" element-loading-svg-view-box="-10, -10, 50, 50">
        <el-table-column prop="id" label="ID" width="180" />
        <el-table-column prop="status" label="状态" width="180" />
        <el-table-column prop="progress" label="进度" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { FormInstance } from 'element-plus'
import { ElMessage } from 'element-plus'
import api from '@/utils/http'
import { BaseResult } from '@/types/axios'

const progressView = ref(false)
const dialogVisible = ref(false)
const showSetting = ref(false)

const loading = ref(false)
const progressLoading = ref(false)
const loadSvg = `
        <path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>
      `
const viewSetting = () => {
  showSetting.value = !showSetting.value
  api.get<BaseResult>({ url: '/api/v1/vul/getImageLoadConfig' }).then((res) => {
    settingConfig.path = res.data.load_path
    settingConfig.docker_health = res.data.docker_health
  })
}

const settingConfig = reactive({
  path: '',
  docker_health: false
})

const progressTableData = ref([] as any)

const columns = reactive([
  { name: '漏洞名称', show: true },
  { name: '评分', show: true },
  { name: '镜像描述', show: true },
  { name: '漏洞类型', show: true },
  { name: '开发语言', show: true },
  { name: '开发数据库', show: true },
  { name: '开发框架', show: true },
  { name: '来源', show: true }
])

const searchFormRef = ref<FormInstance>()
interface SearchForm {
  name: string
  vulType: string
  devLanguage: string
  devDatabase: string
  devMiddleware: string
}
const searchForm = reactive<SearchForm>({
  name: '',
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
  if (
    searchForm.name ||
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
}
// 初始化时获取用户列表
const init = () => {
  getVulImages().then((data) => {
    imageData.value = data
    updateTableData()
  })
}

onMounted(() => {
  init()
})

function getVulImages() {
  // 模拟获取用户列表的 API 请求
  return new Promise(async (resolve) => {
    loading.value = true
    await api
      .get<BaseResult>({
        url: `/api/v1/vul/getVulImages`
      })
      .then((res) => {
        resolve(res.data)
      })
      .catch((error) => {
        ElMessage.error('获取镜像列表失败')
      })
    loading.value = false
  })
}

const imageInput = ref('')
let ws: WebSocket
const pullImage = async (image_name: string) => {
  progressTableData.value = []
  // 从当前URL获取host
  const wsUrl = import.meta.env.VITE_API_URL
  if (wsUrl) {
    ws = new WebSocket(wsUrl + `/api/v1/vul/pullImage?image=${encodeURIComponent(image_name)}`)
  } else {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    ws = new WebSocket(
      `${protocol}//${host}/api/v1/vul/pullImage?image=${encodeURIComponent(image_name)}`
    )
  }
  loading.value = true
  // 实现进度显示
  progressLoading.value = true
  // WebSocket事件处理
  ws.onmessage = (event) => {
    const response = JSON.parse(event.data)
    if (response.code === 200 && response.message !== '镜像拉取完成') {
      progressView.value = true
      progressLoading.value = false
      // 查找已存在的相同ID进度项
      const existingIndex = progressTableData.value.findIndex(
        (item: any) => item.id === response.data.id
      )

      if (existingIndex >= 0) {
        // 存在则更新对应项
        progressTableData.value.splice(existingIndex, 1, response.data)
      } else {
        // 不存在则新增
        progressTableData.value.push(response.data)
      }
    } else {
      ws.close()
      if (response.message === '镜像拉取完成') {
        ElMessage.success('镜像拉取完成')
      } else {
        ElMessage.error(response.message)
      }
    }
  }

  ws.onerror = (error) => {
    ElMessage.error('WebSocket连接异常')
    loading.value = false
    progressView.value = false
  }

  ws.onclose = () => {
    loading.value = false
    progressView.value = false
  }
}

const stopPullImage = () => {
  try {
    ws.send(
      JSON.stringify({
        code: 200,
        message: '终止请求',
        data: {
          action: 'CANCEL_PULL' // 添加action字段
        }
      })
    )
    ElMessage.success('成功发送终止请求')
  } catch (error) {
    ElMessage.error('WebSocket连接异常')
  }
}

const search = () => {
  conditionalArrays.value = []
  tableData.value = []
  for (const index in imageData.value) {
    const shouldInclude =
      (searchForm.name === '' ||
        imageData.value[index]?.image_vul_name?.includes(searchForm.name)) &&
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
      console.log(searchForm.name, imageData.value[index].image_vul_name)
      conditionalArrays.value.push(imageData.value[index])
    }
  }
  total.value = conditionalArrays.value.length
  handlePageChange(1)
}

const changeColumn = (list: any) => {
  columns.values = list
}

const submitUpload = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'

  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return

    if (!file.name.endsWith('.json')) {
      ElMessage.error('请选择.json格式的文件')
      return
    }
    const reader = new FileReader()
    reader.readAsDataURL(file)

    reader.onload = async () => {
      const base64Data = (reader.result as string).split(',')[1] // 去除data:前缀
      loading.value = true
      await api
        .post<BaseResult>({
          url: `/api/v1/vul/uploadImageFile`,
          data: {
            code: 200,
            message: '上传漏洞镜像列表',
            data: {
              filename: file.name, // 文件名
              base64FileData: base64Data // base64编码的文件内容
            }
          }
        })
        .then((res) => {
          ElNotification({
            title: '提示',
            message: h('i', { style: 'color: teal' }, '文件上传成功。')
          })
          init() // 刷新列表
        })
        .catch((error) => {
          ElMessage.error('上传失败')
        })
      loading.value = false
    }

    reader.onerror = () => {
      ElMessage.error('文件读取失败')
    }
  }

  input.click()
}
</script>

<style lang="scss" scoped>
.image-list {
  width: 100%;
  height: 100%;

  .info {
    >div {
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

/* 拉取制定镜像 */
.SelectImage {
  position: relative;
  font-family: inherit;
  font-weight: 700;
  font-size: 1rem;
  letter-spacing: 0.05em;
  border-radius: 0.5rem;
  cursor: pointer;
  border: none;
  background: linear-gradient(to right, #8e2de2, #4a00e0);
  color: ghostwhite;
  overflow: hidden;
  margin-right: 10px;
  box-shadow:
    0 4px 6px -1px #488aec31,
    0 2px 4px -1px #488aec17;
  transition: all 0.6s ease;
}

.SelectImage svg {
  width: 1.25rem;
  height: 1.25rem;
  margin-right: 0.5em;
}

.SelectImage span {
  position: relative;
  z-index: 10;
  transition: color 0.4s;
  display: inline-flex;
  align-items: center;
  padding: 0.75rem 1.5rem;
  gap: 0.75rem;
}

.SelectImage::before,
.SelectImage::after {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

.SelectImage::before {
  content: "";
  background: #000;
  width: 120%;
  left: -10%;
  transform: skew(30deg);
  transition: transform 0.4s cubic-bezier(0.3, 1, 0.8, 1);
}

.SelectImage:hover::before {
  transform: translate3d(100%, 0, 0);
}

.SelectImage:active {
  transform: scale(0.95);
}
</style>
