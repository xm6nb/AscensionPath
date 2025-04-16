// 首先导入api
import api from '@/utils/http'
import { BaseResult } from '@/types/axios'
import { UserInfo } from '@/types/store'
import { useUserStore } from '@/store/modules/user'
import { delCookie } from '@/utils/utils'

export class UserService {
  // 登录接口
  static login(options: { body: any }): Promise<BaseResult> {
    return new Promise((resolve) => {
      // 修改为正确的api.post调用方式
      api
        .post<BaseResult>({
          url: `/api/v1/users/login`,
          data: options.body
        })
        .then((response) => {
          resolve(response)
        })
        .catch((error) => {
          resolve({
            code: error.response?.status || 500,
            message: error.response?.data?.message || '服务器错误',
            data: null
          })
        })
    })
  }

  // 获取用户信息
  static getUserInfo(): Promise<BaseResult<UserInfo>> {
    const userid = useUserStore().getUserInfo.id

    return new Promise((resolve) => {
      api
        .get<BaseResult>({
          url: `/api/v1/users/getUserInfo`,
          params: {
            id: userid
          }
        })
        .then((response) => {
          resolve({
            code: response.code,
            message: response.message,
            data: response.data
          })
        })
        .catch((error) => {
          // 获取当前登录用户信息失败，修改登录状态
          useUserStore().setLoginStatus(false)
          useUserStore().saveUserData()
          resolve({
            code: error.response?.status || 500,
            message: error.response?.data?.message || '获取用户信息失败',
            data: {} as UserInfo
          })
        })
    })
  }
}
