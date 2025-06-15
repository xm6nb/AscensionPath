<template>
  <div class="page-content article-list custom-loading-svg" v-loading="loading" :element-loading-svg="svg"
    element-loading-text="上传中..." element-loading-svg-view-box="-10, -10, 50, 50">
    <div style="display: flex; flex-grow: 1">
      <el-input v-model="searchVal" :prefix-icon="Search" clearable placeholder="根据名称搜索" @keyup.enter="searchVulEnv"
        style="margin-right: 5px" />
      <div class="custom-segmented" style="margin-right: 5px; display: flex; align-items: center">
        <el-segmented v-model="defaultType" :options="options" @change="searchEnvByType" />
      </div>
      <div style="display: flex; align-items: center">
        <button class="cssbuttons-io-button" style="white-space: nowrap" @click="searchVulEnv">
          搜索
          <div class="icon">
            <svg height="24" width="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path d="M0 0h24v24H0z" fill="none"></path>
              <path d="M16.172 11l-5.364-5.364 1.414-1.414L20 12l-7.778 7.778-1.414-1.414L16.172 13H4v-2z"
                fill="currentColor"></path>
            </svg>
          </div>
        </button>
        <div class="flex justify-center items-center gap-2 h-full" style="white-space: nowrap">
          <div class="bg-gradient-to-b from-stone-300/40 to-transparent p-[2px] rounded-[12px]">
            <button
              class="group p-[2px] rounded-[8px] bg-gradient-to-b from-white to-stone-200/40 shadow-[0_1px_2px_rgba(0,0,0,0.3)] active:shadow-[0_0px_1px_rgba(0,0,0,0.2)] active:scale-[0.98]"
              style="height: 2.8em; padding: 0.35em 1.2em 0.35em 1.2em" @click="submitUpload">
              <div class="bg-gradient-to-b from-stone-200/40 to-white/80 rounded-[6px] px-1.5 py-1">
                <div class="flex gap-1 items-center">
                  <span class="font-semibold text-sm">上传Docker Compose压缩包</span>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="list custom-loading-svg" v-loading="isLoading" :element-loading-svg="svg" element-loading-text="加载中..."
      element-loading-svg-view-box="-10, -10, 50, 50" style="min-height: 500px;">
      <div class="offset">
        <div class="item" v-for="(item, index) in vulImageList" :key="index + (currentPage - 1) * pageSize"
          @click="toDetail(item)" v-memo="[isLoading]">
          <!-- 骨架屏 -->
          <el-skeleton animated :loading="isLoading" style="width: 100%; height: 100%">
            <template #template>
              <div class="top">
                <el-skeleton-item variant="image" style="width: 100%; height: 100%; border-radius: 10px" />
                <div style="padding: 16px 0">
                  <el-skeleton-item variant="p" style="width: 80%" />
                  <el-skeleton-item variant="p" style="width: 40%; margin-top: 10px" />
                </div>
              </div>
            </template>

            <template #default>
              <div class="top">
                <el-image class="cover" :src="RandomPngImg(Math.random() * 10)" lazy fit="cover">
                  <template #error>
                    <div class="image-slot">
                      <el-icon><icon-picture /></el-icon>
                    </div>
                  </template>
                </el-image>

                <span class="type">{{ item.env_type }}</span>
              </div>
              <div class="bottom">
                <h2>{{ item.env_name }}</h2>
                <div class="info">
                  <div class="text">
                    <i class="iconfont-sys">&#xe6f7;</i>
                    <span>评分:{{ item.rank }}</span>
                    <div class="line"></div>
                    <i class="iconfont-sys">&#xe689;</i>
                    <span>{{ item.from }}</span>
                  </div>
                  <el-button size="small">详情</el-button>
                </div>
              </div>
            </template>
          </el-skeleton>
        </div>
      </div>
    </div>

    <div style="margin-top: 16vh" v-if="showEmpty">
      <el-empty :description="`未找到相关数据 ${EmojiText[0]}`" />
    </div>

    <div style="display: flex; justify-content: center; margin-top: 20px">
      <el-pagination background v-model:current-page="currentPage" v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]" layout="prev, pager, next, sizes, total, jumper" :total="total"
        :hide-on-single-page="true" @update:current-page="handleCurrentChange" />
    </div>
    <el-dialog v-model="dialogTableVisible" title="镜像信息" width="800">
      <el-descriptions class="margin-top" title="" size="default" direction="vertical" border>
        <template #extra>
          <el-button type="primary" @click="createVulEnv">创建漏洞环境</el-button>
        </template>
        <el-descriptions-item :width="140">
          <template #label>
            <div class="cell-item">
              <el-icon>
                <user />
              </el-icon>
              漏洞名称
            </div>
          </template>
          {{ vulInfo.env_name }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <iphone />
              </el-icon>
              镜像名称
            </div>
          </template>
          {{ vulInfo.base_image }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <location />
              </el-icon>
              镜像类型
            </div>
          </template>
          {{ vulInfo.env_type }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <tickets />
              </el-icon>
              评分
            </div>
          </template>
          <el-tag size="small" type="warning">{{ vulInfo.rank }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <office-building />
              </el-icon>
              来源
            </div>
          </template>
          {{ vulInfo.from }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              漏洞类型
            </div>
          </template>
          <template v-if="vulInfo.degree && vulInfo.degree.HoleType && vulInfo.degree.HoleType.length > 0">
            <ElTag v-for="tag in vulInfo.degree.HoleType" :key="tag">{{ tag }}</ElTag>
          </template>
          <template v-else>
            <ElTag>暂无数据</ElTag>
          </template>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              开发语言
            </div>
          </template>
          <template v-if="
            vulInfo.degree && vulInfo.degree.devLanguage && vulInfo.degree.devLanguage.length > 0
          ">
            <ElTag v-for="tag in vulInfo.degree.devLanguage" :key="tag">{{ tag }}</ElTag>
          </template>
          <template v-else>
            <ElTag>暂无数据</ElTag>
          </template>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              开发数据库
            </div>
          </template>
          <template v-if="
            vulInfo.degree && vulInfo.degree.devDatabase && vulInfo.degree.devDatabase.length > 0
          ">
            <ElTag v-for="tag in vulInfo.degree.devDatabase" :key="tag">{{ tag }}</ElTag>
          </template>
          <template v-else>
            <ElTag>暂无数据</ElTag>
          </template>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              开发框架
            </div>
          </template>
          <template v-if="
            vulInfo.degree && vulInfo.degree.devClassify && vulInfo.degree.devClassify.length > 0
          ">
            <ElTag v-for="tag in vulInfo.degree.devClassify" :key="tag">{{ tag }}</ElTag>
          </template>
          <template v-else>
            <ElTag>暂无数据</ElTag>
          </template>
        </el-descriptions-item>
        <el-descriptions-item :rowspan="2" :span="2">
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              镜像描述
            </div>
          </template>
          {{ vulInfo.env_desc }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <calendar />
              </el-icon>
              路径
            </div>
          </template>
          {{ vulInfo.base_compose }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
    <el-dialog v-model="dialogCreateVulVisible" title="创建漏洞环境" width="800">
      <el-form ref="ruleFormRef" style="max-width: 100%; width: 100%" :model="ruleForm" :rules="rules"
        label-width="auto" class="demo-ruleForm custom-loading-svg" :size="formSize" status-icon
        v-loading="createFormLoading" :element-loading-svg="svg" element-loading-svg-view-box="-10, -10, 50, 50">
        <el-form-item label="漏洞环境名称" prop="env_name">
          <el-input v-model="ruleForm.env_name" />
        </el-form-item>
        <el-form-item label="漏洞种类" prop="env_type">
          <el-select v-model="ruleForm.env_type" placeholder="漏洞种类">
            <el-option label="单镜像" value="单镜像" />
            <el-option label="复合环境" value="复合环境" />
            <el-option label="SQL注入" value="SQL注入" />
            <el-option label="跨站脚本攻击(XSS)" value="跨站脚本攻击(XSS)" />
            <el-option label="跨站请求伪造(CSRF)" value="跨站请求伪造(CSRF)" />
            <el-option label="文件包含漏洞" value="文件包含漏洞" />
            <el-option label="命令执行" value="命令执行" />
            <el-option label="目录遍历" value="目录遍历" />
            <el-option label="XML外部实体注入(XXE)" value="XML外部实体注入(XXE)" />
            <el-option label="服务端请求伪造(SSRF)" value="服务端请求伪造(SSRF)" />
            <el-option label="反序列化漏洞" value="反序列化漏洞" />
            <el-option label="身份验证绕过" value="身份验证绕过" />
            <el-option label="文件上传漏洞" value="文件上传漏洞" />
            <el-option label="配置错误漏洞" value="配置错误漏洞" />
          </el-select>
        </el-form-item>
        <el-form-item label="基础镜像" prop="base_image" v-if="currentItem.base_image !== ''">
          <el-input v-model="ruleForm.base_image" />
        </el-form-item>
        <el-form-item label="Docker compose文件路径" prop="base_compose" v-if="currentItem.base_compose !== ''">
          <el-input v-model="ruleForm.base_compose" />
        </el-form-item>
        <el-form-item>
          <el-col :span="12">
            <el-form-item label="评分" prop="rank">
              <el-input-number v-model="ruleForm.rank" class="mx-4" :min="0" :max="10" :step="0.5"
                controls-position="right" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="来源" prop="from">
              <el-input v-model="ruleForm.from" />
            </el-form-item>
          </el-col>
        </el-form-item>
        <el-form-item>
          <el-col :span="12">
            <el-form-item label="开销" prop="cost">
              <el-input-number v-model="ruleForm.cost" class="mx-4" :min="0" :max="10" controls-position="right" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="开放等级" prop="is_open">
              <el-input-number v-model="ruleForm.is_open" class="mx-4" :min="1" :max="1000" controls-position="right" />
            </el-form-item>
          </el-col>
        </el-form-item>
        <el-form-item label="漏洞环境描述" prop="env_desc">
          <el-input v-model="ruleForm.env_desc" type="textarea" :autosize="{ minRows: 2, maxRows: 6 }"
            placeholder="最好描述一下漏洞环境吧😁" :resize="'none'" />
        </el-form-item>
        <el-form-item style="display: flex; flex-direction: column; margin-bottom: 0">
          <el-button type="primary" @click="submitForm(ruleFormRef)"> 创建 </el-button>
          <el-button @click="resetForm(ruleFormRef)">重置</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
    <el-dialog v-model="progressView" title="拉取进度" width="800" :lock-scroll="true" :close-on-click-modal="false">
      <div style="display: flex; flex-direction: row-reverse">
        <button
          class="px-4 z-30 py-2 bg-rose-400 rounded-md text-white relative font-semibold after:-z-20 after:absolute after:h-1 after:w-1 after:bg-rose-800 after:left-3 overflow-hidden after:bottom-0 after:translate-y-full after:rounded-md after:hover:scale-[300] after:hover:transition-all after:hover:duration-700 after:transition-all after:duration-700 transition-all duration-700 [text-shadow:2px_3px_2px_#be123c;] hover:[text-shadow:1px_1px_2px_#fda4af] text-xl"
          @click="stopPullImage">
          终止
        </button>
      </div>
      <el-table :data="progressTableData" style="width: 100%" max-height="300" :element-loading-svg="svg"
        class="custom-loading-svg" element-loading-svg-view-box="-10, -10, 50, 50">
        <el-table-column prop="id" label="ID" width="180" />
        <el-table-column prop="status" label="状态" width="180" />
        <el-table-column prop="progress" label="进度" />
      </el-table>
      <el-input v-model="buildOutput" style="width: 100%" :autosize="{ minRows: 1, maxRows: 5 }" type="textarea" resize="none" disabled v-if="buildOutput != ''"
        placeholder="镜像构建输出" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Picture as IconPicture } from '@element-plus/icons-vue'
import { ref, onMounted, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import type { ComponentSize, FormInstance, FormRules } from 'element-plus'
import EmojiText from '@/utils/emojo'
import { useCommon } from '@/composables/useCommon'
import api from '@/utils/http'
import { BaseResult } from '@/types/axios'
import { RandomPngImg } from '@/utils/utils'
import { router } from '@/router'
import { RoutesAlias } from '@/router/modules/routesAlias'

const defaultType = ref('All')
const dialogTableVisible = ref(false)

const options = ['All', '单镜像', '复合环境']

const searchVal = ref('')
const vulImageList = ref([] as any)
const vulImageListTmp = ref([] as any)
const AllVulImageList = ref([] as any)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref()
const isLoading = ref(true)
const loading = ref(false)
const createFormLoading = ref(false)
const dialogCreateVulVisible = ref(false)
const progressTableData = ref([] as any)
let currentItem: any = null
const svg = `
        <path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>
      `

const vulInfo = reactive({
  env_name: '',
  env_desc: '',
  env_type: '',
  base_image: '',
  base_compose: '',
  rank: 3.5,
  from: '',
  degree: {} as any
})

const showEmpty = computed(() => {
  return vulImageList.value.length === 0 && !isLoading.value
})

onMounted(() => {
  getImageInfo({ backTop: false })
})

// 搜索环境
const searchVulEnv = async () => {
  isLoading.value = true
  if (searchVal.value !== '') {
    vulImageListTmp.value = AllVulImageList.value.filter((item: any) =>
      item.env_name.includes(searchVal.value)
    )
  } else {
    vulImageListTmp.value = AllVulImageList.value
  }
  total.value = vulImageListTmp.value.length
  await handleCurrentChange(1)
  isLoading.value = false
}

// 根据种类筛选环境
const searchEnvByType = async () => {
  isLoading.value = true
  if (defaultType.value === 'All') {
    vulImageListTmp.value = AllVulImageList.value
  } else {
    vulImageListTmp.value = AllVulImageList.value.filter(
      (item: any) => item.env_type === defaultType.value
    )
  }
  total.value = vulImageListTmp.value.length
  await handleCurrentChange(1)
  isLoading.value = false
}

const getImageInfo = async ({ backTop = false }) => {
  await api
    .get<BaseResult>({
      url: '/api/v1/vul/getVulEnv'
    })
    .then((res) => {
      vulImageListTmp.value = res.data
      AllVulImageList.value = vulImageListTmp.value
      vulImageList.value = vulImageListTmp.value.slice(
        (currentPage.value - 1) * pageSize.value,
        currentPage.value * pageSize.value
      )
      total.value = res.data.length
    })
  if (backTop) {
    useCommon().scrollToTop()
  }
  isLoading.value = false
}

const handleCurrentChange = async (val: number) => {
  currentPage.value = val
  vulImageList.value = vulImageListTmp.value.slice(
    (currentPage.value - 1) * pageSize.value,
    currentPage.value * pageSize.value
  )
}

const toDetail = (item: any) => {
  currentItem = item
  dialogTableVisible.value = true
  vulInfo.env_name = item.env_name
  vulInfo.env_desc = item.env_desc
  vulInfo.env_type = item.env_type
  vulInfo.base_image = item.base_image
  vulInfo.base_compose = item.base_compose
  vulInfo.rank = item.rank
  vulInfo.from = item.from
  vulInfo.degree = item.degree
}

const submitUpload = async () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.zip'

  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return

    if (!file.name.endsWith('.zip')) {
      ElMessage.error('请选择.zip格式的文件')
      return
    }
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = async () => {
      loading.value = true
      const base64Data = (reader.result as string).split(',')[1] // 去除data:前缀
      await api
        .post<BaseResult>({
          url: '/api/v1/vul/uploadVulZip',
          data: {
            code: 200,
            message: '上传漏洞镜像列表',
            data: {
              filename: file.name, // 文件名
              base64FileData: base64Data // base64编码的文件内容
            }
          },
          timeout: 300 * 1000
        })
        .then((res) => {
          ElMessage.success('上传成功')
          getImageInfo({ backTop: true })
        })
        .catch((err) => { })
      loading.value = false
    }

    reader.onerror = () => {
      ElMessage.error('文件读取失败')
    }
  }

  input.click()
}

interface RuleForm {
  env_name: string
  env_type: string
  base_image: string
  base_compose: string
  rank: number
  from: string
  env_desc: string
  degree: any
  cost: number
  is_open: number
}

const formSize = ref<ComponentSize>('default')
const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive<RuleForm>({
  env_name: '',
  env_type: '',
  base_image: '',
  base_compose: '',
  rank: 3.5,
  from: '',
  env_desc: '',
  degree: {} as any,
  cost: 0,
  is_open: 1000
})

const progressView = ref(false)

const rules = reactive<FormRules<RuleForm>>({
  env_name: [
    { required: true, message: '请输入漏洞环境名称', trigger: 'blur' },
    { min: 1, max: 50, message: '长度限制 1-50 个字符', trigger: 'blur' }
  ]
})

const buildOutput = ref('')
let ws: WebSocket
const submitForm = async (formEl: FormInstance | undefined) => {
  if (!formEl) return
  await formEl.validate((valid, fields) => {
    if (valid) {
      // 从当前URL获取host
      const wsUrl = import.meta.env.VITE_API_URL
      if (wsUrl) {
        ws = new WebSocket(wsUrl + '/api/v1/vul/createVulEnv')
      } else {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        ws = new WebSocket(`${protocol}//${host}/api/v1/vul/createVulEnv`)
      }
      progressTableData.value = []
      buildOutput.value = ''
      createFormLoading.value = true
      // 2. 连接成功后发送数据
      ws.onopen = () => {
        ws.send(
          JSON.stringify({
            code: 200,
            message: '发送创建漏洞相关数据',
            data: ruleForm
          })
        )
      }

      // 3. 处理服务器响应
      ws.onmessage = (event) => {
        const response = JSON.parse(event.data)
        // 处理拉取镜像进度条数据
        if (response.code === 200 && response.message !== '漏洞环境创建成功') {
          progressView.value = true
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
          if (response.message === '漏洞环境创建成功') {
            ElNotification({
              title: '提示',
              message: `创建成功: ${response.message}`,
              type: 'success'
            })
            dialogCreateVulVisible.value = false
            router.push(RoutesAlias.CreateInstance)
          } else {
            ElMessage.error(`创建失败: ${response.message}`)
          }
        }
        // 处理构建镜像数据
        if (response.code === 200 && response.message === '构建日志') {
          buildOutput.value = buildOutput.value + response.data.stream
        }
      }

      // 4. 错误处理
      ws.onerror = (error) => {
        ElMessage.error(`WebSocket 连接异常: ${error.type}`) // 正确访问event属性
        ws.close()
      }

      // 5. 关闭连接
      ws.onclose = (event) => {
        console.log('WebSocket 连接已关闭:', event.code, event.reason)
        progressView.value = false
        createFormLoading.value = false
      }
    } else {
    }
  })
}

const resetForm = (formEl: FormInstance | undefined) => {
  if (!formEl) return
  formEl.resetFields()
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

// 创建漏洞环境
const createVulEnv = () => {
  dialogTableVisible.value = false
  dialogCreateVulVisible.value = true
  ruleForm.env_name = currentItem.env_name
  ruleForm.env_type = currentItem.env_type
  ruleForm.base_image = currentItem.base_image
  ruleForm.base_compose = currentItem.base_compose
  ruleForm.rank = currentItem.rank
  ruleForm.from = currentItem.from
  ruleForm.env_desc = currentItem.env_desc
  ruleForm.degree = currentItem.degree
}
</script>

<style lang="scss" scoped>
.article-list {
  .custom-segmented .el-segmented {
    height: 40px;
    padding: 6px;

    --el-border-radius-base: 8px;
  }

  .list {
    margin-top: 20px;

    .offset {
      display: flex;
      flex-wrap: wrap;
      width: calc(100% + 20px);

      .item {
        box-sizing: border-box;
        width: calc(20% - 20px);
        margin: 0 20px 20px 0;
        cursor: pointer;
        border: 1px solid var(--art-border-color);
        border-radius: calc(var(--custom-radius) / 2 + 2px) !important;

        &:hover {
          .el-button {
            opacity: 1 !important;
          }
        }

        .top {
          position: relative;
          aspect-ratio: 16/9.5;

          .cover {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            height: 100%;
            object-fit: cover;
            background: var(--art-gray-200);
            border-radius: calc(var(--custom-radius) / 2 + 2px) calc(var(--custom-radius) / 2 + 2px) 0 0;

            .image-slot {
              font-size: 26px;
              color: var(--art-gray-400);
            }
          }

          .type {
            position: absolute;
            top: 5px;
            right: 5px;
            padding: 5px 4px;
            font-size: 12px;
            color: rgba(#fff, 0.8);
            background: rgba($color: #000, $alpha: 60%);
            border-radius: 4px;
          }
        }

        .bottom {
          padding: 5px 10px;

          h2 {
            font-size: 16px;
            font-weight: 500;
            color: #333;

            @include ellipsis();
          }

          .info {
            display: flex;
            justify-content: space-between;
            width: 100%;
            height: 25px;
            margin-top: 6px;
            line-height: 25px;

            .text {
              display: flex;
              align-items: center;
              color: var(--art-text-gray-600);

              i {
                margin-right: 5px;
                font-size: 14px;
              }

              span {
                font-size: 13px;
                color: var(--art-gray-600);
              }

              .line {
                width: 1px;
                height: 12px;
                margin: 0 15px;
                background-color: var(--art-border-dashed-color);
              }
            }

            .el-button {
              opacity: 0;
              transition: all 0.3s;
            }
          }
        }
      }
    }
  }
}

@media only screen and (max-width: $device-notebook) {
  .article-list {
    .list {
      .offset {
        .item {
          width: calc(25% - 20px);
        }
      }
    }
  }
}

@media only screen and (max-width: $device-ipad-pro) {
  .article-list {
    .list {
      .offset {
        .item {
          width: calc(33.333% - 20px);

          .bottom {
            h2 {
              font-size: 16px;
            }
          }
        }
      }
    }
  }
}

@media only screen and (max-width: $device-ipad) {
  .article-list {
    .list {
      .offset {
        .item {
          width: calc(50% - 20px);
        }
      }
    }
  }
}

@media only screen and (max-width: $device-phone) {
  .article-list {
    .list {
      .offset {
        .item {
          width: calc(100% - 20px);
        }
      }
    }
  }
}

.cell-item {
  white-space: nowrap;
}

.cssbuttons-io-button {
  background: #000000;
  color: white;
  font-family: inherit;
  padding: 0.35em;
  padding-left: 1.2em;
  font-size: 17px;
  font-weight: 500;
  border-radius: 0.9em;
  border: none;
  letter-spacing: 0.05em;
  display: flex;
  align-items: center;
  box-shadow: inset 0 0 1.6em -0.6em #714da6;
  overflow: hidden;
  position: relative;
  height: 2.8em;
  padding-right: 3.3em;
  cursor: pointer;
}

.cssbuttons-io-button .icon {
  background: white;
  margin-left: 1em;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 2.2em;
  width: 2.2em;
  border-radius: 0.7em;
  box-shadow: 0.1em 0.1em 0.6em 0.2em #7b52b9;
  right: 0.3em;
  transition: all 0.3s;
}

.cssbuttons-io-button:hover .icon {
  width: calc(100% - 0.6em);
}

.cssbuttons-io-button .icon svg {
  width: 1.1em;
  transition: transform 0.3s;
  color: #7b52b9;
}

.cssbuttons-io-button:hover .icon svg {
  transform: translateX(0.1em);
}

.cssbuttons-io-button:active .icon {
  transform: scale(0.95);
}
</style>
