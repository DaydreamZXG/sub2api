import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { issueConnectCode, openDesktopApp, stationOrigin } from '@/api/desktop'

/** Hands this signed-in account to the installed desktop assistant with a one-time code. */
export function useYcAppConnect() {
  const { t } = useI18n()
  const appStore = useAppStore()
  const handingOver = ref(false)

  async function connectApp() {
    if (!stationOrigin().startsWith('https://')) {
      appStore.showError(t('yc.connect.appConnectNeedsHttps'))
      return
    }
    handingOver.value = true
    try {
      openDesktopApp(await issueConnectCode())
      appStore.showInfo(t('yc.connect.appConnectHint'))
    } catch {
      appStore.showError(t('yc.connect.appConnectFailed'))
    } finally {
      handingOver.value = false
    }
  }

  return { handingOver, connectApp }
}
