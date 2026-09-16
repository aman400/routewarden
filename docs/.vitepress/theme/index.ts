import DefaultTheme from 'vitepress/theme'
import './custom.css'

const STORAGE_KEY = 'routewarden-preferred-tab'

function getTabTitle(label: HTMLLabelElement): string {
  return (
    label.getAttribute('data-title') ||
    label.textContent ||
    ''
  ).trim()
}

function syncTabs(targetTitle: string, triggeringGroup?: HTMLElement) {
  if (!targetTitle) return
  const normalizedTarget = targetTitle.trim().toLowerCase()

  try {
    localStorage.setItem(STORAGE_KEY, targetTitle.trim())
  } catch {}

  document.querySelectorAll<HTMLElement>('.vp-code-group').forEach((group) => {
    // If this is the group that the user directly clicked, VitePress native listener
    // already handled activating the tab and block. Only sync other groups.
    if (group === triggeringGroup) return

    const labels = Array.from(group.querySelectorAll<HTMLLabelElement>('.tabs label'))
    const targetIdx = labels.findIndex(
      (lbl) => getTabTitle(lbl).toLowerCase() === normalizedTarget
    )

    if (targetIdx === -1) return

    const inputs = Array.from(group.querySelectorAll<HTMLInputElement>('.tabs input'))
    const targetInput = inputs[targetIdx]

    if (targetInput) {
      targetInput.checked = true
    }

    const blocks = group.querySelector<HTMLElement>('.blocks')
    if (blocks) {
      Array.from(blocks.children).forEach((child, idx) => {
        if (idx === targetIdx) {
          child.classList.add('active')
        } else {
          child.classList.remove('active')
        }
      })
    }
  })
}

export default {
  extends: DefaultTheme,
  enhanceApp() {
    if (typeof window !== 'undefined') {
      const handleTabSelection = (target: HTMLElement) => {
        let label: HTMLLabelElement | null = null
        let group: HTMLElement | null = null

        if (target.matches('.vp-code-group input')) {
          const input = target as HTMLInputElement
          group = input.closest<HTMLElement>('.vp-code-group')
          if (group) {
            label = group.querySelector<HTMLLabelElement>(`label[for="${input.id}"]`)
          }
        } else {
          label = target.closest<HTMLLabelElement>('.vp-code-group .tabs label')
          if (label) {
            group = label.closest<HTMLElement>('.vp-code-group')
          }
        }

        if (label && group) {
          const title = getTabTitle(label)
          if (title) {
            syncTabs(title, group)
          }
        }
      }

      // Listen for both clicks and changes on code-group tabs
      window.addEventListener('click', (e) => {
        if (e.target) {
          handleTabSelection(e.target as HTMLElement)
        }
      })

      window.addEventListener('change', (e) => {
        if (e.target) {
          handleTabSelection(e.target as HTMLElement)
        }
      })

      const applyStoredPreference = () => {
        try {
          const saved = localStorage.getItem(STORAGE_KEY)
          if (saved) {
            syncTabs(saved)
          }
        } catch {}
      }

      window.addEventListener('DOMContentLoaded', applyStoredPreference)

      // Support VitePress client-side page navigation
      if (typeof MutationObserver !== 'undefined') {
        let debounceTimer: ReturnType<typeof setTimeout> | null = null
        const observer = new MutationObserver(() => {
          if (debounceTimer) clearTimeout(debounceTimer)
          debounceTimer = setTimeout(() => {
            applyStoredPreference()
          }, 50)
        })
        observer.observe(document.documentElement, { childList: true, subtree: true })
      }
    }
  }
}
