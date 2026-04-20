export interface RespBase {
  code: number
  msg: string
  total?: number
  nextCursor?: string | number | null
  prevCursor?: string | number | null
  hasNext?: boolean
  hasPrev?: boolean
}

export interface ApiResp<T> extends RespBase {
  data?: T
}

export interface PageReq {
  cursor?: number
  limit?: number
}

export interface TimeRange {
  startTime?: number
  endTime?: number
}
