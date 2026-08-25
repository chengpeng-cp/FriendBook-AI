/** 用户认证 / 个人信息 API */
import { get, post, put } from './client'
import type { LoginOutput, User } from '@/types'

export const authApi = {
  /** 注册 */
  register: (payload: { username: string; password: string; nickname?: string }) =>
    post<{ id: number; username: string }>('/api/v1/users/register', payload),

  /** 登录（返回 JWT） */
  login: (payload: { username: string; password: string }) =>
    post<LoginOutput>('/api/v1/users/login', payload),

  /** 获取个人信息 */
  profile: () => get<User>('/api/v1/users/profile'),

  /** 修改个人信息 */
  updateProfile: (payload: { nickname?: string; avatar?: string; bio?: string }) =>
    put<null>('/api/v1/users/profile', payload),

  /** 修改密码 */
  changePassword: (payload: { old_password: string; new_password: string }) =>
    put<null>('/api/v1/users/password', payload),
}
