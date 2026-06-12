import type { App, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores'

type PermValue = string | string[] | undefined

function hasPermission(value: PermValue) {
  const auth = useAuthStore()
  const need = Array.isArray(value) ? value.filter(Boolean) : value ? [value] : []

  if (!need.length) return true
  if (auth.user?.userType === 1 || auth.user?.isOwner === 2) return true

  return need.some((perm) => auth.hasPerm(perm))
}

function updateElement(el: HTMLElement, value: PermValue) {
  if (hasPermission(value)) return
  el.remove()
}

export function setupPermDirective(app: App) {
  app.directive('perm', {
    mounted(el: HTMLElement, binding: DirectiveBinding<PermValue>) {
      updateElement(el, binding.value)
    },
    updated(el: HTMLElement, binding: DirectiveBinding<PermValue>) {
      updateElement(el, binding.value)
    },
  })
}
