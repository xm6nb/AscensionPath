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
      </div>
    </div>

    <div class="list custom-loading-svg" v-loading="isLoading" :element-loading-svg="svg" element-loading-text="加载中..."
      element-loading-svg-view-box="-10, -10, 50, 50" style="min-height: 500px">
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
    <el-dialog v-model="dialogTableVisible" title="场景信息" width="800">
      <el-descriptions class="margin-top custom-loading-svg" title="" size="default" direction="vertical" border
        v-loading="createloading" :element-loading-svg="svg" element-loading-svg-view-box="-10, -10, 50, 50">
        <template #extra style="display: flex">
          <el-popconfirm width="220" :icon="InfoFilled" icon-color="#626AEF" title="是否删除对应镜像" v-if="dialogTableVisible">
            <template #reference>
              <el-button type="info" v-if="useUserStore().getUserInfo.role === 'admin'"
                color="#696969">删除该漏洞环境</el-button>
            </template>
            <template #actions="">
              <el-button size="small" @click="removeVulEnv(vulInfo.vulEnvID, false)">不删除</el-button>
              <el-button type="danger" size="small" @click="removeVulEnv(vulInfo.vulEnvID, true)">
                删除
              </el-button>
            </template>
          </el-popconfirm>
          <el-popconfirm class="box-item" title="是否删除对应镜像" placement="top-start">
            <template #reference> </template>
          </el-popconfirm>
          <el-button type="primary" @click="createInstance" v-if="vulInfo.status !== 1">创建场景实例</el-button>
          <el-button type="warning" @click="extendTime(vulInfo.id)" v-if="vulInfo.status == 1">延长实例时间</el-button>
          <el-button type="danger" @click="removeInstance" v-if="vulInfo.status == 1">移除场景实例</el-button>
        </template>
        <el-descriptions-item :width="140">
          <template #label>
            <div class="cell-item">
              <el-icon>
                <user />
              </el-icon>
              场景名称
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
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <MilkTea />
              </el-icon>
              开销
            </div>
          </template>
          {{ vulInfo.cost }}
        </el-descriptions-item>
      </el-descriptions>
      <el-descriptions class="margin-top" :column="2" border v-if="vulInfo.status !== 0">
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <Bowl />
              </el-icon>
              开启时间
            </div>
          </template>
          {{ vulInfo.start_time }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <Watermelon />
              </el-icon>
              销毁时间
            </div>
          </template>
          {{ vulInfo.expire_time }}
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            <div class="cell-item">
              <el-icon>
                <IceCreamSquare />
              </el-icon>
              访问地址
            </div>
          </template>
          <el-link type="primary" v-for="link in vulInfo.links" :href="link" target="_blank"
            style="margin-right: 10px">{{
            link }}</el-link>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Picture as IconPicture, InfoFilled } from '@element-plus/icons-vue'
import { ref, onMounted, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import EmojiText from '@/utils/emojo'
import { useCommon } from '@/composables/useCommon'
import api from '@/utils/http'
import { BaseResult } from '@/types/axios'
import { RandomPngImg } from '@/utils/utils'
import { formatDate } from '@/utils/utils'
import { useUserStore } from '@/store/modules/user'

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
let currentItem: any = null
const createloading = ref(false)
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
  id: -1,
  env_name: '',
  env_desc: '',
  env_type: '',
  base_image: '',
  userID: -1,
  vulEnvID: -1,
  cost: 0,
  start_time: '',
  status: 0,
  expire_time: '',
  base_compose: '',
  rank: 3.5,
  from: '',
  links: [] as string[],
  degree: {} as any
})

const showEmpty = computed(() => {
  return vulImageList.value.length === 0 && !isLoading.value
})

onMounted(() => {
  getCreatedVulEnv({ backTop: false })
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

const getCreatedVulEnv = async ({ backTop = false }) => {
  isLoading.value = true
  await api
    .get<BaseResult>({
      url: '/api/v1/vul/getCreatedVulEnv'
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
    .catch((err) => { })
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
  vulInfo.id = item.id
  vulInfo.userID = item.user_id
  vulInfo.vulEnvID = item.vul_env_id
  vulInfo.env_name = item.env_name
  vulInfo.env_desc = item.env_desc
  vulInfo.env_type = item.env_type
  vulInfo.base_image = item.base_image
  vulInfo.base_compose = item.base_compose
  vulInfo.rank = item.rank
  vulInfo.from = item.from
  vulInfo.degree = item.degree
  vulInfo.cost = item.cost
  vulInfo.status = item.status
  vulInfo.start_time = formatDate(item.start_time)
  vulInfo.expire_time = formatDate(item.expire_time)
  // 解析端口并生成访问链接
  vulInfo.links = convertPortToLink(item.ports)
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

interface RuleForm {
  vul_env_id: number
  env_name: string
}

const ruleForm = reactive<RuleForm>({
  vul_env_id: 0,
  env_name: ''
})

// 创建漏洞环境
const createInstance = async () => {
  ruleForm.vul_env_id = currentItem.vul_env_id
  ruleForm.env_name = currentItem.env_name
  createloading.value = true
  await api
    .post<BaseResult>({
      url: '/api/v1/vul/createVulInstance',
      data: {
        code: 200,
        message: '创建场景实例',
        data: ruleForm
      }
    })
    .then(async (res) => {
      ElNotification({
        title: '提示',
        message: '成功创建实例场景',
        type: 'success'
      })
      await getCreatedVulEnv({ backTop: false })
      // 从数组中查找匹配的元素
      const matchedItem = AllVulImageList.value.find(
        (item: any) => item.vul_env_id === vulInfo.vulEnvID
      )
      if (matchedItem) {
        vulInfo.id = matchedItem.id
        vulInfo.start_time = formatDate(matchedItem.start_time)
        vulInfo.expire_time = formatDate(matchedItem.expire_time)
        vulInfo.links = convertPortToLink(matchedItem.ports)
        vulInfo.status = matchedItem.status
        vulInfo.userID = matchedItem.user_id
        vulInfo.vulEnvID = matchedItem.vul_env_id
      }
    })
    .catch((err) => { })
  createloading.value = false
}

// 移除漏洞环境
const removeVulEnv = async (id: number, isDeleteImage: boolean) => {
  createloading.value = true
  await api
    .post<BaseResult>({
      url: `/api/v1/vul/deleteVulEnv`,
      data: {
        code: 200,
        message: '创建场景实例',
        data: {
          vul_env_id: id,
          is_delete_image: isDeleteImage
        }
      }
    })
    .then((res) => {
      ElNotification({
        title: '提示',
        message: '成功删除漏洞环境',
        type: 'success'
      })
      dialogTableVisible.value = false
      getCreatedVulEnv({ backTop: false })
    })
    .catch((err) => { })
  createloading.value = false
}

const extendTime = async (id: number) => {
  await api
    .get<BaseResult>({
      url: `/api/v1/vul/extendExpireTime`,
      params: {
        id: id
      }
    })
    .then(async (res) => {
      ElNotification({
        title: '提示',
        message: '该实例已成功延长30分钟',
        type: 'success'
      })
      await getCreatedVulEnv({ backTop: false })
      // 从数组中查找匹配的元素
      const matchedItem = AllVulImageList.value.find(
        (item: any) => item.vul_env_id === vulInfo.vulEnvID
      )
      if (matchedItem) {
        vulInfo.expire_time = formatDate(matchedItem.expire_time)
      }
    })
    .catch((error) => { })
}

const removeInstance = async () => {
  createloading.value = true
  await api
    .post<BaseResult>({
      url: '/api/v1/vul/removeInstance',
      data: {
        code: 200,
        message: '移除场景实例',
        data: {
          user_id: vulInfo.userID,
          vul_env_id: vulInfo.vulEnvID
        }
      }
    })
    .then((res) => {
      ElNotification({
        title: '提示',
        message: '成功移除实例场景',
        type: 'success'
      })
      getCreatedVulEnv({ backTop: false })
      vulInfo.status = 0
    })
    .catch((err) => { })
  createloading.value = false
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
