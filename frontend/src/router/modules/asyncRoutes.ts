import { RoutesAlias } from './routesAlias'
import { MenuListType } from '@/types/menu'

/**
 * 菜单列表、异步路由
 *
 * 支持两种模式:
 * 1. 前端静态配置 - 直接使用本文件中定义的路由配置
 * 2. 后端动态配置 - 后端返回菜单数据，前端解析生成路由
 *
 * 菜单标题（title）:
 * 可以是 i18n 的 key，也可以是字符串，比如：'用户列表'
 */
export const asyncRoutes: MenuListType[] = [
  {
    id: 13,
    name: 'ImageManager',
    path: '/image-manage',
    component: RoutesAlias.Home,
    meta: {
      title:'menus.ImageManager.ImageManage',
      icon: '&#xe6d2;',
      keepAlive: false,
    },
    children: [
      {
        id: 1301,
        path: 'imageList',
        name: 'ImageList',
        component: RoutesAlias.ImageList,
        meta: {
          title:'menus.ImageManager.ImageList',
          keepAlive: false,
          roles: ['admin']
        }
      },
      {
        id: 1302,
        path: 'createVulEnv',
        name: 'CreateVulEnv',
        component: RoutesAlias.CreateVulEnv,
        meta: {
          title:'menus.ImageManager.CreateVulEnv',
          keepAlive: false,
          roles: ['admin']
        }
      },
      {
        id: 1303,
        path: 'createInstance',
        name: 'CreateInstance',
        component: RoutesAlias.CreateInstance,
        meta: {
          title:'menus.ImageManager.CreateInstance',
          keepAlive: false,
        },
      },
      {
        id: 1304,
        path: 'instanceManage',
        name: 'InstanceManage',
        component: RoutesAlias.InstanceManage,
        meta: {
          title:'menus.ImageManager.InstanceManage',
          keepAlive: false,
          roles: ['admin']
        },
      }
    ]
  },
  {
    id: 2,
    name: 'User',
    path: '/user',
    component: RoutesAlias.Home,
    meta: {
      title: 'menus.user.title',
      icon: '&#xe86e;',
      keepAlive: false
    },
    children: [
      {
        id: 301,
        path: 'account',
        name: 'Account',
        component: RoutesAlias.Account,
        meta: {
          title: 'menus.user.account',
          keepAlive: true,
          roles: ['admin']
        }
      },
      {
        id: 304,
        path: 'user',
        name: 'UserCenter',
        component: RoutesAlias.UserCenter,
        meta: {
          title: 'menus.user.userCenter',
          isHide: true,
          keepAlive: true,
          isHideTab: true
        }
      }
    ]
  },
  {
    id: 8,
    path: '/exception',
    name: 'Exception',
    component: RoutesAlias.Home,
    meta: {
      title: 'menus.exception.title',
      icon: '&#xe820;',
      keepAlive: false
    },
    children: [
      {
        id: 801,
        path: '403',
        name: '403',
        component: RoutesAlias.Exception403,
        meta: {
          title: 'menus.exception.forbidden',
          keepAlive: true
        }
      },
      {
        id: 802,
        path: '404',
        name: '404',
        component: RoutesAlias.Exception404,
        meta: {
          title: 'menus.exception.notFound',
          keepAlive: true
        }
      },
      {
        id: 803,
        path: '500',
        name: '500',
        component: RoutesAlias.Exception500,
        meta: {
          title: 'menus.exception.serverError',
          keepAlive: true
        }
      },
      {
        id: 804,
        path: 'incomplete',
        name: 'Incomplete',
        component: RoutesAlias.Incomplete,
        meta: {
          title:'menus.exception.incomplete',
          keepAlive: true,
        }
      }
    ]
  },
  {
    id: 5,
    path: '/widgets',
    name: 'Widgets',
    component: RoutesAlias.Home,
    meta: {
      title: 'menus.widgets.title',
      icon: '&#xe81a;',
      keepAlive: false
    },
    children: [
      {
        id: 515,
        path: 'fireworks',
        name: 'Fireworks',
        component: RoutesAlias.Fireworks,
        meta: {
          title: 'menus.widgets.fireworks',
          keepAlive: true,
          showTextBadge: 'Hot'
        }
      },
    ]
  },
  {
    id: 9,
    path: '/system',
    name: 'System',
    component: RoutesAlias.Home,
    meta: {
      title: 'menus.system.title',
      icon: '&#xe7b9;',
      keepAlive: false
    },
    children: [
      {
        id: 901,
        path: 'setting',
        name: 'Setting',
        component: RoutesAlias.Setting,
        meta: {
          title: 'menus.system.setting',
          keepAlive: true
        }
      },
    ]
  },
  {
    id: 12,
    name: '帮助中心',
    path: '/help',
    component: RoutesAlias.Home,
    meta: {
      title: 'menus.help.title',
      icon: '&#xe719;',
      keepAlive: false
    },
    children: [
      {
        id: 1101,
        path: 'document',
        name: 'Document',
        component: RoutesAlias.Document,
        meta: {
          title: 'menus.help.document',
          keepAlive: false
        }
      }
    ]
  }
]
