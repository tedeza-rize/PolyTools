import { useState, useMemo, useCallback, memo, useEffect } from 'react'
import { Layout, Menu, Input, Button, Badge, Tooltip, Typography, Space } from 'antd'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SearchOutlined,
  ToolOutlined,
} from '@ant-design/icons'

import { useLanguage } from '../contexts/LanguageContext'
import { useTheme } from '../contexts/ThemeContext'
// 새 파일에서 메뉴 정의 import
import { tools as TOOL_LIST, categories as CATEGORY_LIST } from './menuConfig'

const { Sider } = Layout
const { Text } = Typography

/* ------------------------------------------------------------------ */
/*  useMenu : 도구 정의 + 검색어 상태                                   */
/* ------------------------------------------------------------------ */
export const useMenu = () => {
  const { t } = useLanguage()
  const [searchTerm, setSearchTerm] = useState('')

  // i18n 적용된 툴 리스트 생성 (memoized)
  const tools = useMemo(
    () =>
      TOOL_LIST.map((tl) => ({
        ...tl,
        name: t(tl.nameKey), // 실제 문자열로 변환
      })),
    [t],
  )

  return { tools, searchTerm, setSearchTerm }
}

/* ------------------------------------------------------------------ */
/*  Sidebar 컴포넌트                                                    */
/* ------------------------------------------------------------------ */
interface SidebarProps {
  selectedTool: string
  onToolSelect: (tool: string) => void
  collapsed: boolean
  onToggleCollapse: () => void
}

const Sidebar = ({ selectedTool, onToolSelect, collapsed, onToggleCollapse }: SidebarProps) => {
  const { t } = useLanguage()
  const { colors, theme } = useTheme()
  const { tools, searchTerm, setSearchTerm } = useMenu()

  /* ---------------- 선택된 카테고리 키 계산 ---------------- */
  const selectedCategory = useMemo(() => {
    const found = tools.find((tl) => tl.id === selectedTool)
    return found ? found.category : selectedTool
  }, [selectedTool, tools])

  /* ------------- Menu 표시용 선택 키 배열 (다단계 강조) ------------- */
  const selectedKeys = useMemo<string[]>(() => {
    const keys = [selectedTool]
    if (selectedTool !== selectedCategory) keys.push(selectedCategory)
    return keys
  }, [selectedTool, selectedCategory])

  const getToolsByCategory = useCallback(
    (cat: string) =>
      tools.filter(
        (tool) =>
          tool.category === cat &&
          (searchTerm === '' || tool.name.toLowerCase().includes(searchTerm.toLowerCase())),
      ),
    [tools, searchTerm],
  )

  /* ---------------- openKeys : 부모 SubMenu 펼침 ---------------- */
  const [openKeys, setOpenKeys] = useState<string[]>([])
  useEffect(() => {
    if (searchTerm) {
      // 검색어가 있을 때는 검색 결과가 있는 카테고리를 모두 펼침
      const matched = CATEGORY_LIST.filter((c) => getToolsByCategory(c.id).length > 0).map((c) => c.id)
      setOpenKeys(matched)
    } else {
      // 검색어가 없으면 선택된 카테고리만 펼침
      setOpenKeys([selectedCategory])
    }
  }, [searchTerm, selectedCategory, getToolsByCategory])

  /* ---------------- 메뉴 아이템 생성 (메모이즈) ---------------- */
  const menuItems = useMemo(
    () =>
      CATEGORY_LIST.map((category) => {
        const items = getToolsByCategory(category.id)
        const isSimple = category.id === 'home' || category.id === 'settings'
        const isSelectedParent = selectedKeys.includes(category.id)

        // 검색 중 하위 결과가 없으면 해당 카테고리 숨김
        if (!isSimple && items.length === 0 && searchTerm !== '') return null

        /* 부모·자식 공통 스타일 */
        const parentStyle = isSelectedParent ? { backgroundColor: colors.primary, borderRadius: 6 } : undefined
        const parentIconColor = isSelectedParent ? colors.text : category.color

        /* collapsed 상태일 때만 아이콘에 커스텀 Tooltip 부여 */
        const iconNode = collapsed ? (
          <Tooltip
            title={t(category.nameKey)}
            placement="right"
            mouseEnterDelay={0.5}
            color={theme === 'light' ? '#fff' : colors.secondary}
            styles={{ body: { color: colors.text } }} // << CHANGED
          >
            <span style={{ color: parentIconColor }}>{category.icon}</span>
          </Tooltip>
        ) : (
          <span style={{ color: parentIconColor }}>{category.icon}</span>
        )

        return {
          key: category.id,
          icon: iconNode,
          /* label 은 펼쳐진 상태에서만 보이므로 기존 그대로 두기 */
          label: (
            <Tooltip
              title={collapsed ? '' : t(category.nameKey)}
              placement="right"
              mouseEnterDelay={0.5}
              color={theme === 'light' ? '#fff' : colors.secondary}
              styles={{ body: { color: colors.text } }} // << CHANGED
            >
              <Space>
                <Text strong={isSelectedParent} style={{ color: colors.text }}>
                  {t(category.nameKey)}
                </Text>
                {!isSimple && !collapsed && items.length > 0 && (
                  <Badge count={items.length} size="small" style={{ backgroundColor: category.color }} />
                )}
              </Space>
            </Tooltip>
          ),
          title: '', // 기본 AntD 툴팁 비활성화
          style: parentStyle,
          /* 자식(툴) 아이템 */
          children: isSimple
            ? undefined
            : items.map((tool) => ({
                key: tool.id,
                icon: collapsed ? (
                  /* collapsed 상태에선 자식들도 아이콘에 Tooltip */
                  <Tooltip
                    title={tool.name}
                    placement="right"
                    mouseEnterDelay={0.5}
                    color={colors.secondary}
                    styles={{ body: { color: colors.text } }} // << CHANGED
                  >
                    <span style={{ fontSize: 14 }}>{tool.icon}</span>
                  </Tooltip>
                ) : (
                  <span style={{ fontSize: 14 }}>{tool.icon}</span>
                ),
                label: (
                  <Tooltip
                    title={collapsed ? '' : tool.name}
                    placement="right"
                    mouseEnterDelay={0.5}
                    color={colors.secondary}
                    styles={{ body: { color: colors.text } }} // << CHANGED
                  >
                    <Text style={{ color: colors.text }}>{tool.name}</Text>
                  </Tooltip>
                ),
                title: '', // 기본 툴팁 비활성화
                onClick: () => onToolSelect(tool.id),
              })),
          onClick: isSimple ? () => onToolSelect(category.id) : undefined,
        }
      }).filter(Boolean) as any[],
    [collapsed, colors.primary, colors.text, colors.secondary, getToolsByCategory, searchTerm, selectedKeys, t, onToolSelect, theme],
  )

  /* ------------------------------------------------------------------ */
  /*  placeholder 컬러를 동적으로 세팅 (다크모드 가독성 개선)            */
  /* ------------------------------------------------------------------ */
  useEffect(() => {
    const styleId = 'sidebar-placeholder-style'
    let styleEl = document.getElementById(styleId) as HTMLStyleElement | null
    if (!styleEl) {
      styleEl = document.createElement('style')
      styleEl.id = styleId
      document.head.appendChild(styleEl)
    }
    styleEl.innerHTML = `
      .sidebar-search-input::placeholder {
        color: ${theme === 'dark' ? 'rgba(255,255,255,0.55)' : 'rgba(0,0,0,0.45)'};
      }
    `
  }, [theme])

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={onToggleCollapse}
      trigger={null}
      width={280}
      collapsedWidth={80}
      style={{
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        zIndex: 1000,
        boxShadow: '2px 0 6px rgba(0,0,0,0.1)',
        backgroundColor: colors.background,
      }}
    >
      <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        {/* ---------- 헤더(로고 + 토글) ---------- */}
        <div
          style={{
            height: 64,
            padding: '0 16px',
            borderBottom: `1px solid ${colors.secondary}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'space-between',
            flexShrink: 0,
          }}
        >
          {!collapsed && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <div
                style={{
                  width: 32,
                  height: 32,
                  background: colors.primary,
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 16,
                }}
              >
                <ToolOutlined style={{ color: colors.text }} />
              </div>
              <Text strong style={{ fontSize: 16, color: colors.text }}>
                {t('common_app_name')}
              </Text>
            </div>
          )}
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={onToggleCollapse}
            style={{ fontSize: 16, width: 40, height: 40, color: colors.text }}
          />
        </div>

        {/* ---------- 검색창 ---------- */}
        {!collapsed && (
          <div style={{ padding: '16px', flexShrink: 0, borderBottom: `1px solid ${colors.secondary}` }}>
            <Input
              className="sidebar-search-input"
              placeholder={t('home_search_placeholder')}
              prefix={<SearchOutlined style={{ color: colors.text }} />}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              allowClear
              style={{ borderRadius: 6, backgroundColor: colors.secondary, color: colors.text }}
            />
          </div>
        )}

        {/* ---------- 메뉴 ---------- */}
        <div
          className="sidebar-menu-container"
          style={{ flex: 1, minHeight: 0, overflowY: 'auto', overflowX: 'hidden' }}
        >
          <Menu
            mode="inline"
            theme={theme}
            selectedKeys={selectedKeys}
            openKeys={openKeys}
            onOpenChange={(keys) => setOpenKeys(keys as string[])}
            style={{ border: 'none', background: 'transparent', color: colors.text, padding: '16px 0' }}
            items={menuItems}
            inlineIndent={24}
          />
        </div>

        {/* ---------- 축약 로고(접힘 상태) ---------- */}
        {collapsed && (
          <div
            style={{
              flexShrink: 0,
              textAlign: 'center',
              padding: '20px 0',
              borderTop: `1px solid ${colors.secondary}`,
            }}
          >
            <div
              style={{
                width: 40,
                height: 40,
                background: colors.primary,
                borderRadius: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 18,
                marginBottom: 8,
                marginLeft: 'auto',
                marginRight: 'auto',
              }}
            >
              <ToolOutlined style={{ color: colors.text }} />
            </div>
            <Text style={{ fontSize: 12, color: colors.text, opacity: 0.8 }}>{t('common_app_name_short')}</Text>
          </div>
        )}
      </div>
    </Sider>
  )
}

export default memo(Sidebar)