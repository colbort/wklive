import { get, post, put, del } from '@/utils/request'

import type {
  RespBase,
  SysCronJobListReq,
  SysCronJobItem,
  SysCronJobCreateReq,
  SysCronJobUpdateReq,
  SysCronJobHandler,
  SysCronJobLogItem,
  SysCronJobLogListReq,
} from '@/services'

export function apiSysCronJobList(params: SysCronJobListReq): Promise<RespBase<SysCronJobItem[]>> {
  return get<SysCronJobItem[]>('/admin/system/jobs', params)
}

export function apiSysCronJobCreate(data: SysCronJobCreateReq): Promise<RespBase> {
  return post('/admin/system/jobs', data)
}

export function apiSysCronJobUpdate(data: SysCronJobUpdateReq): Promise<RespBase> {
  return put('/admin/system/jobs', data)
}

export function apiSysCronJobDelete(id: number): Promise<RespBase> {
  return del(`/admin/system/jobs/${id}`)
}

export function apiSysCronJobRun(id: number): Promise<RespBase> {
  return post(`/admin/system/jobs/${id}/run`)
}

export function apiSysCronJobStart(id: number): Promise<RespBase> {
  return post(`/admin/system/jobs/${id}/start`)
}

export function apiSysCronJobStop(id: number): Promise<RespBase> {
  return post(`/admin/system/jobs/${id}/stop`)
}

export function apiSysCronJobHandlers(): Promise<RespBase<SysCronJobHandler[]>> {
  return get<SysCronJobHandler[]>('/admin/system/jobs/handlers')
}

export function apiSysCronJobLogList(
  params: SysCronJobLogListReq,
): Promise<RespBase<SysCronJobLogItem[]>> {
  return get<SysCronJobLogItem[]>('/admin/system/logs/job', params)
}
