import { useEffect, useRef } from 'react'
import { useMenu } from '../contexts/MenuContext'
import HomePage from '../nav/HomePage'
import SettingsPage from '../nav/SettingsPage'
import Base64Tool from '../nav/Base64Tool'
import UrlTool from '../nav/UrlTool' // 방금 만든 컴포넌트
import Differ from '../nav/Differ'
import JsonFormatter from '../nav/JsonFormatter'
// import ComingSoon from '../nav/ComingSoon'

interface MainContentProps {
  selectedTool: string
  sidebarCollapsed: boolean
}

const MainContent = ({ selectedTool, sidebarCollapsed }: MainContentProps) => {
  const { addToRecent } = useMenu()
  const prevSelectedTool = useRef<string>('')

  // 도구 선택 시 최근 사용 목록에 추가 (렌더링과 분리)
  useEffect(() => {
    // 이전 선택과 다르고, home/settings가 아닌 경우에만 추가
    if (
      selectedTool !== prevSelectedTool.current &&
      selectedTool !== 'home' &&
      selectedTool !== 'settings'
    ) {
      addToRecent(selectedTool)
    }
    prevSelectedTool.current = selectedTool
  }, [selectedTool])

  const renderContent = () => {
    switch (selectedTool) {
      case 'home':
        return <HomePage />
      case 'settings':
        return <SettingsPage />
      case 'base64':
        return <Base64Tool />
      case 'url':
        return <UrlTool />
      case 'json':
        return <JsonFormatter />
	  case 'diff':
	    return <Differ />
      default:
        return <JsonFormatter />
    }
  }

  return (
    <main
      className={`main-content ${sidebarCollapsed ? 'sidebar-collapsed' : ''}`}
      style={{
        height: '100vh',
        overflowY: 'auto',
        minHeight: 0,
        background: '#fff', // 필요에 따라 조정
      }}
    >
      <div
        className="content-wrapper"
        style={{
          minHeight: 0,
          height: '100%',
          overflowY: 'auto',
        }}
      >
        {renderContent()}
      </div>
    </main>
  )
}

export default MainContent
