import { createContext, useContext, useState, useEffect, ReactNode } from 'react'

type Language = 'ko' | 'en'
type Translations = { [key: string]: string | Translations }

interface LanguageContextType {
  language: Language
  setLanguage: (lang: Language) => void
  t: (key: string) => string
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined)

export const LanguageProvider = ({ children }: { children: ReactNode }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    const savedLanguage = localStorage.getItem('language')
    if (savedLanguage) {
      return savedLanguage as Language
    }
    const systemLanguage = navigator.language.startsWith('ko') ? 'ko' : 'en'
    return systemLanguage
  })
  const [translations, setTranslations] = useState<Translations>({})

  useEffect(() => {
    const loadTranslations = async () => {
      try {
        const translationsModule = await import(`../locales/${language}.json`)
        setTranslations(translationsModule.default)
      } catch (error) {
        console.error("Error loading translations:", error)
        // Fallback to English if loading fails
        const fallbackModule = await import('../locales/en.json')
        setTranslations(fallbackModule.default)
      }
    }
    loadTranslations()
    localStorage.setItem('language', language)
  }, [language])

  const setLanguage = (lang: Language) => {
    setLanguageState(lang)
  }

  const t = (key: string): string => {
    const keys = key.split('.');
    let result: string | Translations | undefined = translations;
    for (const k of keys) {
      if (result && typeof result === 'object' && k in result) {
        result = result[k] as string | Translations;
      } else {
        return key;
      }
    }
    return typeof result === 'string' ? result : key;
  }

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  )
}

export const useLanguage = () => {
  const context = useContext(LanguageContext)
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider')
  }
  return context
}