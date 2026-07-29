import { create } from 'zustand'
import en from '../locales/en.json'
import fa from '../locales/fa.json'
import es from '../locales/es.json'
import ru from '../locales/ru.json'
import zh from '../locales/zh.json'
import ar from '../locales/ar.json'
import de from '../locales/de.json'
import fr from '../locales/fr.json'
import pt from '../locales/pt.json'
import tr from '../locales/tr.json'

export type Locale = 'en' | 'fa' | 'es' | 'ru' | 'zh' | 'ar' | 'de' | 'fr' | 'pt' | 'tr'

export interface LocaleInfo {
  code: Locale
  name: string
  nativeName: string
  dir: 'ltr' | 'rtl'
  flag: string
}

export const localeList: LocaleInfo[] = [
  { code: 'en', name: 'English', nativeName: 'English', dir: 'ltr', flag: '🇺🇸' },
  { code: 'fa', name: 'Persian', nativeName: 'فارسی', dir: 'rtl', flag: '🇮🇷' },
  { code: 'ar', name: 'Arabic', nativeName: 'العربية', dir: 'rtl', flag: '🇸🇦' },
  { code: 'es', name: 'Spanish', nativeName: 'Español', dir: 'ltr', flag: '🇪🇸' },
  { code: 'ru', name: 'Russian', nativeName: 'Русский', dir: 'ltr', flag: '🇷🇺' },
  { code: 'zh', name: 'Chinese', nativeName: '中文', dir: 'ltr', flag: '🇨🇳' },
  { code: 'de', name: 'German', nativeName: 'Deutsch', dir: 'ltr', flag: '🇩🇪' },
  { code: 'fr', name: 'French', nativeName: 'Français', dir: 'ltr', flag: '🇫🇷' },
  { code: 'pt', name: 'Portuguese', nativeName: 'Português', dir: 'ltr', flag: '🇧🇷' },
  { code: 'tr', name: 'Turkish', nativeName: 'Türkçe', dir: 'ltr', flag: '🇹🇷' },
]

const rtlLocales = new Set<Locale>(['fa', 'ar'])

const translations: Record<Locale, Record<string, any>> = {
  en, fa, es, ru, zh, ar, de, fr, pt, tr,
}

function getInitialLocale(): Locale {
  const stored = localStorage.getItem('vortex-locale') as Locale | null
  if (stored && translations[stored]) return stored
  const browserLang = navigator.language?.slice(0, 2)
  // Map browser language to our supported locale
  const browserMap: Record<string, Locale> = {
    fa: 'fa', ar: 'ar', es: 'es', ru: 'ru', zh: 'zh',
    de: 'de', fr: 'fr', pt: 'pt', tr: 'tr',
  }
  if (browserLang && browserMap[browserLang]) return browserMap[browserLang]
  return 'en'
}

function resolveKey(obj: any, path: string): string {
  const keys = path.split('.')
  let current = obj
  for (const key of keys) {
    if (current?.[key] === undefined) return path
    current = current[key]
  }
  return typeof current === 'string' ? current : path
}

function t(locale: Locale, path: string, params?: Record<string, string | number>): string {
  const lang = translations[locale] || translations.en
  let text = resolveKey(lang, path)

  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      text = text.replace(new RegExp(`{${key}}`, 'g'), String(value))
    })
  }

  return text
}

interface I18nState {
  locale: Locale
  isRTL: boolean
  setLocale: (l: Locale) => void
  t: (path: string, params?: Record<string, string | number>) => string
}

const initial = getInitialLocale()

export const useI18nStore = create<I18nState>((set, get) => ({
  locale: initial,
  isRTL: rtlLocales.has(initial),

  setLocale: (l: Locale) => {
    localStorage.setItem('vortex-locale', l)
    const root = document.documentElement
    const isRTL = rtlLocales.has(l)
    if (isRTL) {
      root.setAttribute('dir', 'rtl')
    } else {
      root.setAttribute('dir', 'ltr')
    }
    root.setAttribute('lang', l)
    set({ locale: l, isRTL })
  },

  t: (path: string, params?: Record<string, string | number>) => {
    return t(get().locale, path, params)
  },
}))

// Apply initial locale on load
const root = document.documentElement
const isRTL = rtlLocales.has(initial)
root.setAttribute('dir', isRTL ? 'rtl' : 'ltr')
root.setAttribute('lang', initial)

// Hook for convenience
export function useI18n() {
  const store = useI18nStore()
  return { t: store.t, locale: store.locale, isRTL: store.isRTL, setLocale: store.setLocale }
}
