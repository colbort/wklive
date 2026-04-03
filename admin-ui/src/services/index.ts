// 服务层统一导出
export type { BaseService } from './BaseService'
export { BaseServiceImpl } from './BaseService'
export { UserService, userService } from './system/UserService'
export { RoleService, roleService } from './system/RoleService'
export { MenuService, menuService } from './system/MenuService'
export { LogService, logService } from './system/LogService'
export { UploadService, uploadService } from './system/UploadService'
export { ConfigService, configService } from './system/ConfigService'
export { CronJobService, cronJobService } from './system/CronJobService'
export { TenantsService, tenantsService } from './system/TenantsService'
export { CategoriesService, categoriesService } from './itick/CategoriesService'
export { ProductsService, productsService } from './itick/ProductsService'
export { TenantCategoriesService, tenantCategoriesService } from './itick/TenantCategoriesService'
export { TenantProductsService, tenantProductsService } from './itick/TenantProductsService'

// 类型导出
export type { RespBase } from './BaseService'
export type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UserQueryParams,
  SysUserItem,
  Google2FABindInitResp,
} from './system/UserService'
export type {
  Role,
  CreateRoleRequest,
  UpdateRoleRequest,
  RoleQueryParams,
  RoleGrantRequest,
  SysRole,
  RoleListResp,
  RoleItem,
} from './system/RoleService'
export type {
  Menu,
  Permission,
  MenuQueryParams,
  CreateMenuRequest,
  UpdateMenuRequest,
  MenuNode,
  PermItem,
  SysMenuCreateReq,
  SysMenuUpdateReq,
  SysMenuListReq,
  SysMenuListResp,
  SysMenuItem,
  SysMenuTreeItem,
} from './system/MenuService'
export type {
  LoginLog,
  OperationLog,
  LoginLogQueryParams,
  OperationLogQueryParams,
  LoginLogItem,
  LoginLogListReq,
  LoginLogListResp,
  OpLogItem,
  OpLogListReq,
  OpLogListResp,
} from './system/LogService'
export type { UploadFileResp } from './system/UploadService'
export type {
  SysConfigItem,
  SysConfigListReq,
  SysConfigCreateReq,
  SysConfigUpdateReq,
} from './system/ConfigService'
export type {
  SysCronJobItem,
  SysCronJobListReq,
  SysCronJobListResp,
  SysCronJobCreateReq,
  SysCronJobUpdateReq,
  SysCronJobDeleteReq,
  SysCronJobRunReq,
  SysCronJobStartReq,
  SysCronJobStopReq,
  SysCronJobHandler,
  SysCronJobHandlersResp,
  SysCronJobLogItem,
  SysCronJobLogListReq,
  SysCronJobLogListResp,
} from './system/CronJobService'
export type {
  SysTenantItem,
  SysTenantListReq,
  SysTenantCreateReq,
  SysTenantUpdateReq,
} from './system/TenantsService'

export type {
  ListCategoriesReq,
  ItickCategory,
  CreateCategoryReq,
  UpdateCategoryReq,
  SyncCategoryProductsReq,
  SyncCategoryProductsResp,
} from './itick/CategoriesService'

export type {
  ListProductsReq,
  ItickProduct,
  CreateProductReq,
  UpdateProductReq,
  GetProductKlineReq,
  Kline,
} from './itick/ProductsService'
