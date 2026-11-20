import { ref } from 'vue'
import { getReseller } from '@/api/reseller'

// Whether the signed-in user is a reseller, asked once per page load: it only decides a menu entry.
const isReseller = ref(false)
let asked = false

export function useResellerFlag() {
  function loadResellerFlag() {
    if (asked) return
    asked = true
    getReseller()
      .then((reseller) => {
        isReseller.value = reseller !== null
      })
      .catch(() => {
        asked = false
      })
  }
  return { isReseller, loadResellerFlag }
}
