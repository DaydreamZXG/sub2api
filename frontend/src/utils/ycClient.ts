export type YcClientPlatform = 'windows' | 'macos'

/** Installers of the desktop client. */
export const YC_CLIENT_DOWNLOADS: Record<YcClientPlatform, string> = {
  windows: 'https://ergou.qzz.io/releases/official/yeschoy-windows-x86_64-installer.exe',
  macos: 'https://ergou.qzz.io/releases/official/yeschoy-macos-universal-installer.dmg'
}

/** The installer that fits this browser, or null on phones, tablets and systems without one. */
export function detectYcClientPlatform(): YcClientPlatform | null {
  const userAgent = navigator.userAgent.toLowerCase()
  const platform = (navigator.platform || '').toLowerCase()
  // iPadOS asks for desktop sites with a Mac user agent; only touch gives it away.
  const touchMac = platform.startsWith('mac') && navigator.maxTouchPoints > 1
  if (/android|iphone|ipad|ipod/.test(userAgent) || touchMac) return null
  if (userAgent.includes('windows') || platform.startsWith('win')) return 'windows'
  if (userAgent.includes('macintosh') || platform.startsWith('mac')) return 'macos'
  return null
}
