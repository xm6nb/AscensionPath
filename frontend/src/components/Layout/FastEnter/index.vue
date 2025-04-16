<template>
  <el-popover
    ref="popoverRef"
    :width="700"
    trigger="hover"
    popper-class="fast-enter-popover"
    :show-arrow="false"
    placement="bottom-start"
    :offset="0"
    popper-style="border: 1px solid var(--art-border-dashed-color); border-radius: calc(var(--custom-radius) / 2 + 4px); "
  >
    <template #reference>
      <div class="fast-enter-trigger">
        <div class="btn">
          <i class="iconfont-sys">&#xe81a;</i>
          <span class="red-dot"></span>
        </div>
      </div>
    </template>

    <div class="fast-enter">
      <div class="apps-section">
        <div class="apps-grid">
          <!-- 左侧应用列表 -->
          <div
            class="app-item"
            v-for="app in filteredApplications"
            :key="app.name"
            @click="handleAppClick(app.path)"
          >
            <div class="app-icon">
              <i class="iconfont-sys" v-html="app.icon" :style="{ color: app.iconColor }"></i>
            </div>
            <div class="app-info">
              <h3>{{ app.name }}</h3>
              <p>{{ app.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="quick-links">
        <h3>快速链接</h3>
        <ul>
          <li v-for="link in quickLinks" :key="link.name" @click="handleAppClick(link.path)">
            <span>{{ link.name }}</span>
          </li>
        </ul>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { ref } from 'vue'
  import { RoutesAlias } from '@/router/modules/routesAlias'
  import { useUserStore } from '@/store/modules/user'

  const router = useRouter()
  const popoverRef = ref()

  interface Application {
    name: string
    description: string
    icon: string
    iconColor: string
    path: string
    role: string[]
  }

  interface QuickLink {
    name: string
    path: string
  }

  const userRole = computed(() => useUserStore().getUserInfo.role || '')

  const filteredApplications = computed(() => 
    applications.filter(app => 
      app.role.length === 0 || app.role.includes(userRole.value)
    )
  )

  const applications: Application[] = [
    {
      name: '镜像列表',
      description: '系统从JSON中读取的镜像',
      icon: '&#xe721;',
      iconColor: '#377dff',
      path: RoutesAlias.ImageList,
      role: ['admin']
    },
    {
      name: '漏洞环境配置',
      description: '创建漏洞环境',
      icon: '&#xe812;',
      iconColor: '#ff3b30',
      path: RoutesAlias.CreateVulEnv,
      role: ['admin']
    },
    {
      name: '创建实例',
      description: '将漏洞环境部署到实例',
      icon: '&#xe788;',
      iconColor: '#ffb100',
      path: RoutesAlias.CreateInstance,
      role:[]
    },
    {
      name: '实例管理',
      description: '管理所有用户开启的实例环境',
      icon: '&#xe86e;',
      iconColor: '#ff6b6b',
      path: RoutesAlias.InstanceManage,
      role: ['admin']
    },
    {
      name: '账号管理',
      description: '管理所有系统账号',
      icon: '&#xe70a;',
      iconColor: '#13DEB9',
      path: RoutesAlias.Account,
      role: ['admin']
    },
    {
      name: '礼花效果',
      description: '动画特效展示',
      icon: '&#xe7ed;',
      iconColor: '#7A7FFF',
      path: RoutesAlias.Fireworks,
      role: []
    },
  ]

  const quickLinks: QuickLink[] = [
    { name: '登录', path: '/login' },
    { name: '注册', path: '/register' },
    { name: '个人中心', path: '/user/user' },
  ]

  const handleAppClick = (path: string) => {
    if (path.startsWith('http')) {
      window.open(path, '_blank')
    } else {
      router.push(path)
    }
    popoverRef.value?.hide()
  }
</script>

<style lang="scss" scoped>
  .fast-enter-trigger {
    display: flex;
    gap: 8px;
    align-items: center;

    .btn {
      position: relative;
      display: block;
      width: 38px;
      height: 38px;
      line-height: 38px;
      text-align: center;
      cursor: pointer;
      border-radius: 6px;
      transition: all 0.2s;

      i {
        display: block;
        font-size: 19px;
        color: var(--art-gray-600);
      }

      &:hover {
        color: var(--main-color);
        background-color: rgba(var(--art-gray-200-rgb), 0.7);
      }

      .red-dot {
        position: absolute;
        top: 8px;
        right: 8px;
        width: 6px;
        height: 6px;
        background-color: var(--el-color-danger);
        border-radius: 50%;
      }
    }
  }

  .fast-enter {
    display: grid;
    grid-template-columns: 2fr 0.8fr;

    .apps-section {
      .apps-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 6px;
      }

      .app-item {
        display: flex;
        gap: 12px;
        align-items: center;
        padding: 8px 12px;
        margin-right: 12px;
        cursor: pointer;
        border-radius: 8px;

        &:hover {
          background-color: rgba(var(--art-gray-200-rgb), 0.7);

          .app-icon {
            background-color: transparent !important;
          }
        }

        .app-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 46px;
          height: 46px;
          background-color: rgba(var(--art-gray-200-rgb), 0.7);
          border-radius: 8px;

          i {
            font-size: 20px;
          }
        }

        .app-info {
          h3 {
            margin: 0;
            font-size: 14px;
            font-weight: 500;
            color: var(--art-text-gray-800);
          }

          p {
            margin: 4px 0 0;
            font-size: 12px;
            color: var(--art-text-gray-500);
          }
        }
      }
    }

    .quick-links {
      padding: 8px 0 0 24px;
      border-left: 1px solid var(--el-border-color-lighter);

      h3 {
        margin: 0 0 10px;
        font-size: 16px;
        font-weight: 500;
        color: var(--art-text-gray-800);
      }

      ul {
        li {
          padding: 8px 0;
          cursor: pointer;

          &:hover {
            span {
              color: var(--el-color-primary);
            }
          }

          span {
            color: var(--art-text-gray-600);
            text-decoration: none;
          }
        }
      }
    }
  }
</style>
