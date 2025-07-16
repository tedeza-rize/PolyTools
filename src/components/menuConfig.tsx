import { HomeOutlined, SwapOutlined, LockOutlined, FileTextOutlined, ThunderboltOutlined, EditOutlined, BgColorsOutlined, SettingOutlined, LinkOutlined, KeyOutlined, NumberOutlined, CodeOutlined, CheckCircleOutlined, BranchesOutlined, PictureOutlined, BgColorsOutlined as PaletteOutlined, ClockCircleOutlined, BarsOutlined, BarcodeOutlined } from '@ant-design/icons'

export interface Tool {
  id: string
  nameKey: string
  category: string
  icon: React.ReactNode
}

export interface Category {
  id: string
  nameKey: string
  icon: React.ReactNode
  color: string
}

export const tools: Tool[] = [
  // ────────── Encoders ────────── //
  { id: 'base64', nameKey: 'tools_base64', category: 'encoders', icon: <LockOutlined /> },
  { id: 'url', nameKey: 'tools_url', category: 'encoders', icon: <LinkOutlined /> },

  // ────────── Formatting Tools ────────── //
  { id: 'formatter', nameKey: 'tools_formatter', category: 'code_optimizer', icon: <CodeOutlined /> },
  { id: 'minifier', nameKey: 'tools_minifier', category: 'code_optimizer', icon: <BarsOutlined /> },
  { id: 'validator', nameKey: 'tools_validator', category: 'code_optimizer', icon: <CheckCircleOutlined /> },

  // ────────── Generators ────────── //
  { id: 'uuid', nameKey: 'tools_uuid', category: 'generators', icon: <ThunderboltOutlined /> },
  { id: 'password', nameKey: 'tools_password', category: 'generators', icon: <KeyOutlined /> },
  { id: 'hash', nameKey: 'tools_hash', category: 'generators', icon: <NumberOutlined /> },

  // ────────── Text Utilities ────────── //
  { id: 'regex', nameKey: 'tools_regex', category: 'text', icon: <BranchesOutlined /> },
  { id: 'diff', nameKey: 'tools_diff', category: 'text', icon: <FileTextOutlined /> },

  // ────────── Graphic ────────── //
  { id: 'color', nameKey: 'tools_color', category: 'graphic', icon: <PaletteOutlined /> },
  { id: 'image', nameKey: 'tools_image', category: 'graphic', icon: <PictureOutlined /> },

  // ────────── Converters ────────── //
  { id: 'hex', nameKey: 'tools_hex', category: 'converters', icon: <BarcodeOutlined /> },
  { id: 'timestamp', nameKey: 'tools_timestamp', category: 'converters', icon: <ClockCircleOutlined /> },
]

export const categories: Category[] = [
  { id: 'home', nameKey: 'common_home', icon: <HomeOutlined />, color: '#1890ff' },
  { id: 'converters', nameKey: 'common_converters', icon: <SwapOutlined />, color: '#52c41a' },
  { id: 'encoders', nameKey: 'common_encoders', icon: <LockOutlined />, color: '#722ed1' },
  { id: 'code_optimizer', nameKey: 'code_optimizer', icon: <FileTextOutlined />, color: '#faad14' },
  { id: 'generators', nameKey: 'common_generators', icon: <ThunderboltOutlined />, color: '#fadb14' },
  { id: 'text', nameKey: 'common_text', icon: <EditOutlined />, color: '#eb2f96' },
  { id: 'graphic', nameKey: 'common_graphic', icon: <BgColorsOutlined />, color: '#f5222d' },
  { id: 'settings', nameKey: 'common_settings', icon: <SettingOutlined />, color: '#8c8c8c' },
]
