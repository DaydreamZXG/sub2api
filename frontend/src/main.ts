import { isIOSDevice } from '@/utils/device'
import './style.css'
import './styles/yc-theme.css'

function initIOSViewportZoomFix() {
  // iOS Safari 在输入框字号小于 16px 时聚焦会自动放大页面，且失焦后不会恢复。
  // 限制 maximum-scale 可阻止该行为；iOS 10+ 用户仍可双指手动缩放，不影响可访问性。
  // 仅在 iOS 设备上注入，避免影响 Android Chrome 的手动缩放能力。
  if (!isIOSDevice()) return

  const viewport = document.querySelector('meta[name="viewport"]')
  if (!viewport) return

  const content = viewport.getAttribute('content') || ''
  if (/maximum-scale/i.test(content)) return
  viewport.setAttribute('content', `${content}, maximum-scale=1.0`)
}

function initThemeClass() {
  const savedTheme = localStorage.getItem('theme')
  const shouldUseDark =
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

async function bootstrap() {
  // Apply theme class globally before app mount to keep all routes consistent.
  initThemeClass()
  initIOSViewportZoomFix()

  // On the resellers' client domain the server sends its own small config: only the client pages
  // run there. The site itself is a separate chunk, so none of its code is loaded on that domain.
  const injected = window.__APP_CONFIG__ as unknown as { portal?: boolean } | undefined
  if (injected?.portal === true) {
    const { mountPortal } = await import('./portal/mount')
    mountPortal(injected as unknown as import('./portal/mount').PortalConfig)
    return
  }

  const { bootstrapMainApp } = await import('./mainApp')
  await bootstrapMainApp()
}

bootstrap()
