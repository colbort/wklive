import type { RespBase } from '@/services'
import {
  apiCryptoRechargeAddressCreate,
  apiCryptoRechargeAddressList,
  apiCryptoRechargeAddressUpdate,
  apiCryptoRechargeTxCreate,
  apiCryptoRechargeTxList,
  apiCryptoRechargeTxUpdate,
  apiCryptoWalletAccountCreate,
  apiCryptoWalletAccountList,
  apiCryptoWalletAccountUpdate,
} from '@/api/payment/crypto'

export type CryptoRechargeAddress = {
  id: number
  tenantId: number
  userId: number
  walletType: number
  coin: string
  chainCode: number
  address: string
  memo: string
  addressSource: number
  addressType: number
  status: number
  lastUsedTime: number
  createTimes: number
  updateTimes: number
}

export type CryptoWalletAccount = {
  id: number
  tenantId: number
  accountCode: string
  accountName: string
  provider: string
  apiKeyCipher: string
  apiSecretCipher: string
  callbackSecretCipher: string
  extConfig: string
  status: number
  isDefault: number
  createTimes: number
  updateTimes: number
}

export type CryptoRechargeTx = {
  id: number
  tenantId: number
  userId: number
  orderId: number
  orderNo: string
  coin: string
  chainCode: number
  txHash: string
  fromAddress: string
  toAddress: string
  memo: string
  amount: string
  blockHeight: number
  confirmCount: number
  requiredConfirmCount: number
  status: number
  rawData: string
  createTimes: number
  updateTimes: number
}

export type ListCryptoRechargeAddressesReq = Partial<CryptoRechargeAddress> & {
  cursor?: string | null
  limit?: number
}

export type CreateCryptoRechargeAddressReq = Omit<
  CryptoRechargeAddress,
  'id' | 'lastUsedTime' | 'createTimes' | 'updateTimes'
>
export type UpdateCryptoRechargeAddressReq = Pick<
  CryptoRechargeAddress,
  'id' | 'tenantId' | 'address' | 'memo' | 'addressSource' | 'addressType' | 'status'
>

export type ListCryptoWalletAccountsReq = {
  tenantId?: number
  keyword?: string
  provider?: string
  status?: number
  isDefault?: number
  cursor?: string | null
  limit?: number
}
export type CreateCryptoWalletAccountReq = Omit<
  CryptoWalletAccount,
  'id' | 'createTimes' | 'updateTimes'
>
export type UpdateCryptoWalletAccountReq = Omit<
  CryptoWalletAccount,
  'accountCode' | 'createTimes' | 'updateTimes'
>

export type ListCryptoRechargeTxsReq = {
  tenantId?: number
  userId?: number
  orderNo?: string
  coin?: string
  chainCode?: number
  txHash?: string
  toAddress?: string
  status?: number
  cursor?: string | null
  limit?: number
}
export type CreateCryptoRechargeTxReq = Omit<CryptoRechargeTx, 'id' | 'createTimes' | 'updateTimes'>
export type UpdateCryptoRechargeTxReq = Pick<
  CryptoRechargeTx,
  | 'id'
  | 'tenantId'
  | 'orderId'
  | 'orderNo'
  | 'confirmCount'
  | 'requiredConfirmCount'
  | 'status'
  | 'rawData'
>

export class CryptoService {
  listRechargeAddresses(
    params: ListCryptoRechargeAddressesReq,
  ): Promise<RespBase<CryptoRechargeAddress[]>> {
    return apiCryptoRechargeAddressList(params)
  }

  createRechargeAddress(params: CreateCryptoRechargeAddressReq): Promise<RespBase> {
    return apiCryptoRechargeAddressCreate(params)
  }

  updateRechargeAddress(params: UpdateCryptoRechargeAddressReq): Promise<RespBase> {
    return apiCryptoRechargeAddressUpdate(params)
  }

  listWalletAccounts(
    params: ListCryptoWalletAccountsReq,
  ): Promise<RespBase<CryptoWalletAccount[]>> {
    return apiCryptoWalletAccountList(params)
  }

  createWalletAccount(params: CreateCryptoWalletAccountReq): Promise<RespBase> {
    return apiCryptoWalletAccountCreate(params)
  }

  updateWalletAccount(params: UpdateCryptoWalletAccountReq): Promise<RespBase> {
    return apiCryptoWalletAccountUpdate(params)
  }

  listRechargeTxs(params: ListCryptoRechargeTxsReq): Promise<RespBase<CryptoRechargeTx[]>> {
    return apiCryptoRechargeTxList(params)
  }

  createRechargeTx(params: CreateCryptoRechargeTxReq): Promise<RespBase> {
    return apiCryptoRechargeTxCreate(params)
  }

  updateRechargeTx(params: UpdateCryptoRechargeTxReq): Promise<RespBase> {
    return apiCryptoRechargeTxUpdate(params)
  }
}

export const cryptoService = new CryptoService()
