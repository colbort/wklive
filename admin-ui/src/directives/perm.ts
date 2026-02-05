import type { App, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function setupPermDirective(app: App) {
  app.directive('perm', {
    mounted(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
      const auth = useAuthStore()
      const need = binding.value
      const ok = Array.isArray(need) ? need.some((p) => auth.hasPerm(p)) : auth.hasPerm(need)
      if (!ok) el.remove()
    },
  })
}
