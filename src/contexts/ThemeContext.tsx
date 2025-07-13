import { createContext, useContext, useState, useEffect, ReactNode } from 'react'

type ThemeName = 'light' | 'dark'
type ThemeColors = { [key: string]: string }

interface ThemeContextType {
  theme: ThemeName
  setTheme: (themeName: ThemeName) => void
  colors: ThemeColors
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setThemeState] = useState<ThemeName>(() => {
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme) {
      return savedTheme as ThemeName
    }
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    return systemPrefersDark ? 'dark' : 'light'
  })
  const [colors, setColors] = useState<ThemeColors>({})

  useEffect(() => {
    const loadTheme = async () => {
      try {
        const themeModule = await import(`../themes/${theme}.json`)
        const data = themeModule.default
        setColors(data.colors)

        // Apply CSS variables
        for (const [key, value] of Object.entries(data.colors)) {
          document.documentElement.style.setProperty(`--color-${key}`, value as string)
        }
        document.documentElement.setAttribute('data-theme', theme)

      } catch (error) {
        console.error("Error loading theme:", error)
        // Fallback to light theme if loading fails
        const fallbackModule = await import('../themes/light.json')
        const data = fallbackModule.default
        setColors(data.colors)
        for (const [key, value] of Object.entries(data.colors)) {
          document.documentElement.style.setProperty(`--color-${key}`, value as string)
        }
        document.documentElement.setAttribute('data-theme', 'light')
      }
    }
    loadTheme()
    localStorage.setItem('theme', theme)
  }, [theme])

  const setTheme = (themeName: ThemeName) => {
    setThemeState(themeName)
  }

  return (
    <ThemeContext.Provider value={{ theme, setTheme, colors }}>
      {children}
    </ThemeContext.Provider>
  )
}

export const useTheme = () => {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}