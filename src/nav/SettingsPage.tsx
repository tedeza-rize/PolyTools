import { useLanguage } from '../contexts/LanguageContext'
import { useTheme } from '../contexts/ThemeContext'

const SettingsPage = () => {
  const { t, language, setLanguage } = useLanguage()
  const { theme, setTheme, colors } = useTheme()

  return (
    <div className="settings-page" style={{ backgroundColor: colors.background, color: colors.text }}>
      <div className="page-header" style={{ borderBottom: `1px solid ${colors.secondary}` }}>
        <h1 style={{ color: colors.primary }}>{t('settings_title')}</h1>
        <p style={{ color: colors.text }}>{t('settings_description')}</p>
      </div>

      <div className="settings-content">
        <div className="setting-section">
          <h2 style={{ color: colors.primary }}>{t('settings_theme')}</h2>
          <div className="setting-group">
            <label className="setting-item">
              <input
                type="radio"
                name="theme"
                value="light"
                checked={theme === 'light'}
                onChange={() => setTheme('light')}
              />
              <span className="setting-label" style={{ color: colors.text }}>
                <span className="setting-icon">☀️</span>
                {t('settings_light')}
              </span>
            </label>
            <label className="setting-item">
              <input
                type="radio"
                name="theme"
                value="dark"
                checked={theme === 'dark'}
                onChange={() => setTheme('dark')}
              />
              <span className="setting-label" style={{ color: colors.text }}>
                <span className="setting-icon">🌙</span>
                {t('settings_dark')}
              </span>
            </label>
          </div>
        </div>

        <div className="setting-section">
          <h2 style={{ color: colors.primary }}>{t('settings_language')}</h2>
          <div className="setting-group">
            <label className="setting-item">
              <input
                type="radio"
                name="language"
                value="ko"
                checked={language === 'ko'}
                onChange={() => setLanguage('ko')}
              />
              <span className="setting-label" style={{ color: colors.text }}>
                <span className="setting-icon">🇰🇷</span>
                {t('settings_korean')}
              </span>
            </label>
            <label className="setting-item">
              <input
                type="radio"
                name="language"
                value="en"
                checked={language === 'en'}
                onChange={() => setLanguage('en')}
              />
              <span className="setting-label" style={{ color: colors.text }}>
                <span className="setting-icon">🇺🇸</span>
                {t('settings_english')}
              </span>
            </label>
          </div>
        </div>

        <div className="setting-section">
          <h2 style={{ color: colors.primary }}>{t('settings_info_section_title')}</h2>
          <div className="setting-group">
            <div className="info-item">
              <span className="info-label" style={{ color: colors.text }}>{t('settings_version_label')}</span>
              <span className="info-value" style={{ color: colors.text }}>v2.0-preview.8</span>
            </div>
            <div className="info-item">
              <span className="info-label" style={{ color: colors.text }}>{t('settings_build_label')}</span>
              <span className="info-value" style={{ color: colors.text }}>2024.07.11</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default SettingsPage