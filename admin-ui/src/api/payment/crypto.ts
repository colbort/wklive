import { get, post, put } from '@/utils/request'
import type {
  CreateCryptoRechargeAddressReq,
  CreateCryptoRechargeTxReq,
  CreateCryptoWalletAccountReq,
  CryptoRechargeAddress,
  CryptoRechargeTx,
  CryptoWalletAccount,
  ListCryptoRechargeAddressesReq,
  ListCryptoRechargeTxsReq,
  ListCryptoWalletAccountsReq,
  RespBase,
  UpdateCryptoRechargeAddressReq,
  UpdateCryptoRechargeTxReq,
  UpdateCryptoWalletAccountReq,
} from '@/services'

export function apiCryptoRechargeAddressList(
  params: ListCryptoRechargeAddressesReq,
): Promise<RespBase<CryptoRechargeAddress[]>> {
  return get<CryptoRechargeAddress[]>('/admin/payment/crypto-recharge-addresses', params)
}

export function apiCryptoRechargeAddressCreate(
  params: CreateCryptoRechargeAddressReq,
): Promise<RespBase> {
  return post('/admin/payment/crypto-recharge-address', params)
}

export function apiCryptoRechargeAddressUpdate(
  params: UpdateCryptoRechargeAddressReq,
): Promise<RespBase> {
  return put('/admin/payment/crypto-recharge-address', params)
}

export function apiCryptoWalletAccountList(
  params: ListCryptoWalletAccountsReq,
): Promise<RespBase<CryptoWalletAccount[]>> {
  return get<CryptoWalletAccount[]>('/admin/payment/crypto-wallet-accounts', params)
}

export function apiCryptoWalletAccountCreate(
  params: CreateCryptoWalletAccountReq,
): Promise<RespBase> {
  return post('/admin/payment/crypto-wallet-account', params)
}

export function apiCryptoWalletAccountUpdate(
  params: UpdateCryptoWalletAccountReq,
): Promise<RespBase> {
  return put('/admin/payment/crypto-wallet-account', params)
}

export function apiCryptoRechargeTxList(
  params: ListCryptoRechargeTxsReq,
): Promise<RespBase<CryptoRechargeTx[]>> {
  return get<CryptoRechargeTx[]>('/admin/payment/crypto-recharge-txs', params)
}

export function apiCryptoRechargeTxCreate(params: CreateCryptoRechargeTxReq): Promise<RespBase> {
  return post('/admin/payment/crypto-recharge-tx', params)
}

export function apiCryptoRechargeTxUpdate(params: UpdateCryptoRechargeTxReq): Promise<RespBase> {
  return put('/admin/payment/crypto-recharge-tx', params)
}
