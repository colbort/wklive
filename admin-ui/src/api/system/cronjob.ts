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
  return get<SysCronJobItem[]>('/admin/jobs', params)
}

export function apiSysCronJobCreate(data: SysCronJobCreateReq): Promise<RespBase> {
  return post('/admin/jobs', data)
}

export function apiSysCronJobUpdate(data: SysCronJobUpdateReq): Promise<RespBase> {
  return put('/admin/jobs', data)
}

export function apiSysCronJobDelete(id: number): Promise<RespBase> {
  return del(`/admin/jobs/${id}`)
}

export function apiSysCronJobRun(id: number): Promise<RespBase> {
  return post(`/admin/jobs/${id}/run`)
}

export function apiSysCronJobStart(id: number): Promise<RespBase> {
  return post(`/admin/jobs/${id}/start`)
}

export function apiSysCronJobStop(id: number): Promise<RespBase> {
  return post(`/admin/jobs/${id}/stop`)
}

export function apiSysCronJobHandlers(): Promise<RespBase<SysCronJobHandler[]>> {
  return get<SysCronJobHandler[]>('/admin/jobs/handlers')
}

export function apiSysCronJobLogList(
  params: SysCronJobLogListReq,
): Promise<RespBase<SysCronJobLogItem[]>> {
  return get<SysCronJobLogItem[]>('/admin/logs/job', params)
}
