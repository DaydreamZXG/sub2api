import { createApp } from 'vue'
import PortalApp from './PortalApp.vue'

export interface PortalConfig {
  portal: true
  site_name: string
  portal_logo: string
  contact_info: string
  sign: string
}

/** The client pages of the resellers' domain: one small app, with none of the main site's routes or stores. */
export function mountPortal(config: PortalConfig) {
  document.title = config.site_name
  createApp(PortalApp, { config }).mount('#app')
}
