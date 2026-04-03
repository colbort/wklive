import type { RespBase, BaseService } from '@/services'
import {
  apiSysCronJobList,
  apiSysCronJobCreate,
  apiSysCronJobUpdate,
  apiSysCronJobDelete,
  apiSysCronJobRun,
  apiSysCronJobStart,
  apiSysCronJobStop,
  apiSysCronJobHandlers,
  apiSysCronJobLogList,
} from '@/api/system/cronjob'

export type SysCronJobItem = {
  id: number
  jobName: string
  jobGroup: string
  invokeTarget: string
  cronExpression: string
  status: number
  remark?: string
  createBy: string
  createTime: number
  updateBy: string
  updateTime: number
}

export type SysCronJobListReq = {
  cursor?: string | null
  limit?: number
  keyword?: string
  jobName?: string
  jobGroup?: string
  status?: number
}

export type SysCronJobListResp = RespBase<SysCronJobItem[]>

export type SysCronJobCreateReq = {
  jobName: string
  jobGroup: string
  invokeTarget: string
  cronExpression: string
  status: number
  remark?: string
}

export type SysCronJobUpdateReq = {
  id: number
  jobName?: string
  jobGroup?: string
  invokeTarget?: string
  cronExpression?: string
  status?: number
  remark?: string
}

export type SysCronJobDeleteReq = {
  id: number
}

export type SysCronJobRunReq = {
  id: number
}

export type SysCronJobStartReq = {
  id: number
}

export type SysCronJobStopReq = {
  id: number
}

export type SysCronJobHandler = {
  invokeTarget: string // 定时任务处理器唯一标识（如 sys:demo:handler）
  jobName: string // 定时任务处理器名称（如 sys:demo:handler）
}

export type SysCronJobHandlersResp = RespBase<SysCronJobHandler[]>

export type SysCronJobLogItem = {
  id: number
  jobId: number
  jobName: string
  invokeTarget: string
  cronExpression: string
  status: number
  message?: string
  exceptionInfo?: string
  startTime: number
  endTime: number
  createTime: number
}

export type SysCronJobLogListReq = {
  cursor?: string | null
  limit?: number
  jobId?: number
  jobName?: string
  invokeTarget?: string
  status?: number
}

export type SysCronJobLogListResp = RespBase<SysCronJobLogItem[]>

export class CronJobService implements BaseService {
  async getList(params: SysCronJobListReq): Promise<RespBase<SysCronJobItem[]>> {
    return apiSysCronJobList(params)
  }
  async create(data: SysCronJobCreateReq): Promise<RespBase> {
    return apiSysCronJobCreate(data)
  }
  async update(id: string | number, data: Partial<SysCronJobUpdateReq>): Promise<RespBase> {
    return apiSysCronJobUpdate({ id: Number(id), ...data })
  }
  async delete(id: number): Promise<RespBase> {
    return apiSysCronJobDelete(id)
  }
  async run(id: number): Promise<RespBase> {
    return apiSysCronJobRun(id)
  }
  async start(id: number): Promise<RespBase> {
    return apiSysCronJobStart(id)
  }
  async stop(id: number): Promise<RespBase> {
    return apiSysCronJobStop(id)
  }
  async handlers(): Promise<RespBase<SysCronJobHandler[]>> {
    return apiSysCronJobHandlers()
  }
  async getLogList(params: SysCronJobLogListReq): Promise<RespBase<SysCronJobLogItem[]>> {
    return apiSysCronJobLogList(params)
  }
}

export const cronJobService = new CronJobService()
