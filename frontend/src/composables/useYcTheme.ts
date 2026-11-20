import { ref, watch } from 'vue'
import { usePreferredDark } from '@vueuse/core'

export type YcThemeMode = 'system' | 'light' | 'dark'

// Same storage key the admin sidebar toggle writes; no stored value means "follow the system".
const STORAGE_KEY = 'theme'

export function useYcTheme() {
  const saved = localStorage.getItem(STORAGE_KEY)
  const mode = ref<YcThemeMode>(saved === 'dark' || saved === 'light' ? saved : 'system')
  const systemDark = usePreferredDark()

  watch(
    [mode, systemDark],
    () => {
      const dark = mode.value === 'dark' || (mode.value === 'system' && systemDark.value)
      document.documentElement.classList.toggle('dark', dark)
    },
    { immediate: true }
  )

  function setMode(next: YcThemeMode) {
    mode.value = next
    if (next === 'system') localStorage.removeItem(STORAGE_KEY)
    else localStorage.setItem(STORAGE_KEY, next)
  }

  return { mode, setMode }
}
