import React from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { useMenu, Tool } from '../contexts/MenuContext';
import { useTheme } from '../contexts/ThemeContext';

const HomePage: React.FC = () => {
  const { t } = useLanguage();
  const { tools, recentTools, searchTerm, setSearchTerm, filteredTools } = useMenu();
  const { colors } = useTheme();

  const getRecentTools = (): Tool[] =>
    recentTools
      .map(id => tools.find(tool => tool.id === id))
      .filter((tool): tool is Tool => Boolean(tool));

  const getPopularTools = (): Tool[] => tools.filter(tool => tool.popular);

  const displayTools = searchTerm ? filteredTools : tools;

  return (
    <div className="home-page" style={{ backgroundColor: colors.background, color: colors.text }}>
      <div className="home-header" style={{ borderBottom: `1px solid ${colors.secondary}` }}>
        <div className="home-title">
          <h1 style={{ color: colors.primary }}>{t('home.title')}</h1>
          <p className="home-subtitle" style={{ color: colors.text }}>{t('homepage.welcome')}</p>
        </div>

        <div className="home-search">
          <input
            type="text"
            placeholder={t('home.search.placeholder')}
            value={searchTerm}
            onChange={e => setSearchTerm(e.target.value)}
            className="search-input large"
            style={{ backgroundColor: colors.secondary, color: colors.text, borderColor: colors.primary }}
          />
        </div>
      </div>

      <div className="home-content">
        {searchTerm ? (
          <section className="tools-section">
            <h2 style={{ color: colors.primary }}>{t('homepage.search_results')}</h2>
            <div className="tools-grid">
              {displayTools.map(tool => (
                <div key={tool.id} className="tool-card" style={{ backgroundColor: colors.secondary, borderColor: colors.primary }}>
                  <div className="tool-icon" style={{ color: colors.primary }}>{tool.icon}</div>
                  <div className="tool-info">
                    <h3 style={{ color: colors.text }}>{t(tool.name)}</h3>
                    <p style={{ color: colors.text }}>{t(`tool_descriptions.${tool.id}`)}</p>
                  </div>
                  {tool.popular && (
                    <div className="tool-badge" style={{ backgroundColor: colors.primary, color: colors.text }}>
                      {t('common.popular')}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </section>
        ) : (
          <>
            {getRecentTools().length > 0 && (
              <section className="tools-section">
                <h2 style={{ color: colors.primary }}>{t('homepage.recent')}</h2>
                <div className="tools-grid">
                  {getRecentTools().map(tool => (
                    <div key={tool.id} className="tool-card" style={{ backgroundColor: colors.secondary, borderColor: colors.primary }}>
                      <div className="tool-icon" style={{ color: colors.primary }}>{tool.icon}</div>
                      <div className="tool-info">
                        <h3 style={{ color: colors.text }}>{t(tool.name)}</h3>
                        <p style={{ color: colors.text }}>{t(`tool_descriptions.${tool.id}`)}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            )}

            <section className="tools-section">
              <h2 style={{ color: colors.primary }}>{t('homepage.popular_tools')}</h2>
              <div className="tools-grid">
                {getPopularTools().map(tool => (
                  <div key={tool.id} className="tool-card" style={{ backgroundColor: colors.secondary, borderColor: colors.primary }}>
                    <div className="tool-icon" style={{ color: colors.primary }}>{tool.icon}</div>
                    <div className="tool-info">
                      <h3 style={{ color: colors.text }}>{t(tool.name)}</h3>
                      <p style={{ color: colors.text }}>{t(`tool_descriptions.${tool.id}`)}</p>
                    </div>
                    <div className="tool-badge" style={{ backgroundColor: colors.primary, color: colors.text }}>
                      {t('common.popular')}
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className="tools-section">
              <h2 style={{ color: colors.primary }}>{t('homepage.all_tools')}</h2>
              <div className="tools-grid">
                {tools.map(tool => (
                  <div key={tool.id} className="tool-card" style={{ backgroundColor: colors.secondary, borderColor: colors.primary }}>
                    <div className="tool-icon" style={{ color: colors.primary }}>{tool.icon}</div>
                    <div className="tool-info">
                      <h3 style={{ color: colors.text }}>{t(tool.name)}</h3>
                      <p style={{ color: colors.text }}>{t(`tool_descriptions.${tool.id}`)}</p>
                    </div>
                    {tool.popular && (
                      <div className="tool-badge" style={{ backgroundColor: colors.primary, color: colors.text }}>
                        {t('common.popular')}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </section>
          </>
        )}
      </div>
    </div>
  );
};

export default HomePage;
