import { createContext, useContext, useState, ReactNode } from 'react'

interface Tool {
  id: string
  name: string
  icon: string
  category: string
  description: string
  popular?: boolean
}

interface MenuContextType {
  tools: Tool[]
  recentTools: string[]
  addToRecent: (toolId: string) => void
  searchTerm: string
  setSearchTerm: (term: string) => void
  filteredTools: Tool[]
}

const tools: Tool[] = [
  // 인코더/디코더
  { id: 'base64', name: 'tool.base64', icon: '🔐', category: 'encoders', description: 'Base64 데이터를 인코딩 및 디코딩', popular: true },
  { id: 'url', name: 'tool.url', icon: '🔗', category: 'encoders', description: 'URL을 인코딩 및 디코딩' },
  { id: 'html', name: 'tool.html', icon: '📝', category: 'encoders', description: 'HTML 텍스트를 인코딩 및 디코딩' },
  { id: 'jwt', name: 'tool.jwt', icon: '🎫', category: 'encoders', description: 'JWT 토큰을 디코딩' },
  
  // 포맷터
  { id: 'json', name: 'tool.json', icon: '📄', category: 'formatters', description: 'JSON 데이터를 포맷팅', popular: true },
  { id: 'xml', name: 'tool.xml', icon: '📋', category: 'formatters', description: 'XML 데이터를 포맷팅' },
  { id: 'yaml', name: 'tool.yaml', icon: '📑', category: 'formatters', description: 'YAML 데이터를 포맷팅' },
  
  // 생성기
  { id: 'uuid', name: 'tool.uuid', icon: '🆔', category: 'generators', description: 'UUID를 생성', popular: true },
  { id: 'password', name: 'tool.password', icon: '🔑', category: 'generators', description: '안전한 패스워드를 생성' },
  { id: 'hash', name: 'tool.hash', icon: '#️⃣', category: 'generators', description: '해시값을 생성' },
  { id: 'qr', name: 'tool.qr', icon: '📱', category: 'generators', description: 'QR 코드를 생성' },
  { id: 'lorem', name: 'tool.lorem', icon: '📝', category: 'generators', description: 'Lorem Ipsum 텍스트를 생성' },
  
  // 텍스트
  { id: 'regex', name: 'tool.regex', icon: '🔍', category: 'text', description: '정규식을 테스트' },
  { id: 'diff', name: 'tool.diff', icon: '📊', category: 'text', description: '텍스트 간의 차이를 비교' },
  { id: 'markdown', name: 'tool.markdown', icon: '📝', category: 'text', description: '마크다운을 프리뷰' },
  
  // 변환기
  { id: 'cron', name: 'tool.cron', icon: '⏰', category: 'converters', description: 'Cron 표현식을 해석' },
  { id: 'timestamp', name: 'tool.timestamp', icon: '📅', category: 'converters', description: '타임스탬프를 변환', popular: true },
  { id: 'color', name: 'tool.color', icon: '🎨', category: 'converters', description: '색상 포맷을 변환' },
  
  // 그래픽
  { id: 'image', name: 'tool.image', icon: '🖼️', category: 'graphic', description: '이미지를 변환' },
]

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

  const filteredTools = tools.filter(tool => {
    if (!searchTerm) return true
    return tool.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
           tool.description.toLowerCase().includes(searchTerm.toLowerCase())
  })

  return (
    <MenuContext.Provider value={{
      tools,
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