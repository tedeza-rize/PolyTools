import { createContext, useContext, useState, ReactNode } from 'react'

import { tools as allTools, Tool } from '../components/menuConfig'

interface MenuContextType {
  tools: Tool[]
  recentTools: string[]
  addToRecent: (toolId: string) => void
  searchTerm: string
  setSearchTerm: (term: string) => void
  filteredTools: Tool[]
}



const MenuContext = createContext<MenuContextType | undefined>(undefined)

export const MenuProvider = ({ children }: { children: ReactNode }) => {
  const [recentTools, setRecentTools] = useState<string[]>(() => {
    const saved = localStorage.getItem('recentTools')
    return saved ? JSON.parse(saved) : []
  })
  
  const [searchTerm, setSearchTerm] = useState('')

  const addToRecent = (toolId: string) => {
    setRecentTools(prev => {
      const filtered = prev.filter(id => id !== toolId)
      const updated = [toolId, ...filtered].slice(0, 5)
      localStorage.setItem('recentTools', JSON.stringify(updated))
      return updated
    })
  }

  const filteredTools = allTools.filter(tool => {
    if (!searchTerm) return true
    return tool.nameKey.toLowerCase().includes(searchTerm.toLowerCase())
  })

  return (
    <MenuContext.Provider value={{
      tools: allTools,
      recentTools,
      addToRecent,
      searchTerm,
      setSearchTerm,
      filteredTools
    }}>
      {children}
    </MenuContext.Provider>
  )
}

export const useMenu = () => {
  const context = useContext(MenuContext)
  if (!context) {
    throw new Error('useMenu must be used within a MenuProvider')
  }
  return context
}