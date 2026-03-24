import { useAuthStore } from '@/stores'
export function setupPermDirective(app) {
  app.directive('perm', {
    mounted(el, binding) {
      const auth = useAuthStore()
      const need = binding.value
      const ok = Array.isArray(need) ? need.some((p) => auth.hasPerm(p)) : auth.hasPerm(need)
      if (!ok) el.remove()
    },
  })
}
