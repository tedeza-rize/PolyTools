import { useState } from 'react'
import { MenuProvider } from './contexts/MenuContext'
import { ThemeProvider } from './contexts/ThemeContext'
import { LanguageProvider } from './contexts/LanguageContext'
import Sidebar from './components/Sidebar'
import MainContent from './components/MainContent'
import './App.css'

function App() {
  const [selectedTool, setSelectedTool] = useState('home')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  return (
    <LanguageProvider>
      <ThemeProvider>
        <MenuProvider>
          <div className="app">
            <Sidebar 
              selectedTool={selectedTool}
              onToolSelect={setSelectedTool}
              collapsed={sidebarCollapsed}
              onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
            />
            <MainContent 
              selectedTool={selectedTool}
              onToolSelect={setSelectedTool} // 페이지 이동 함수 전달
              sidebarCollapsed={sidebarCollapsed}
            />
          </div>
        </MenuProvider>
      </ThemeProvider>
    </LanguageProvider>
  )
}

export default App