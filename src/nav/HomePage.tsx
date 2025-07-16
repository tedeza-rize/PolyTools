import React from 'react';
import { Card, Row, Col, Typography, Empty, Tag, Space } from 'antd';
import { useLanguage } from '../contexts/LanguageContext';
import { useMenu } from '../contexts/MenuContext';
import { useTheme } from '../contexts/ThemeContext';
import { Tool } from '../components/menuConfig';

// 홈페이지에서 도구 선택(페이지 이동)이 가능하도록 props 정의
interface HomePageProps {
  onToolSelect: (toolId: string) => void;
}

const HomePage: React.FC<HomePageProps> = ({ onToolSelect }) => {
  const { t } = useLanguage();
  const { colors } = useTheme();
    const { tools, categories = [], recentTools, searchTerm, filteredTools } = useMenu();

  // 도구 카드 클릭 시 App의 상태를 변경하여 페이지 이동
  const handleToolClick = (toolId: string) => {
    onToolSelect(toolId);
  };

  // 최근 사용 도구 ID 배열을 실제 도구 객체 배열로 변환
  const recentToolDetails = recentTools
    .map(id => tools.find(tool => tool.id === id))
    .filter((tool): tool is Tool => Boolean(tool));

  // 재사용 가능한 도구 카드 렌더링 함수
  const renderToolCard = (tool: Tool) => (
    <Col key={tool.id} xs={24} sm={12} md={8} lg={8} xl={6}>
      <Card
        hoverable
        onClick={() => handleToolClick(tool.id)}
        style={{ backgroundColor: colors.secondary, border: `1px solid ${colors.border}` }}
        bodyStyle={{ padding: '16px' }}
      >
        <Space align="center">
          <span style={{ fontSize: '22px', color: colors.primary, marginTop: '4px' }}>{tool.icon}</span>
          <Typography.Text style={{ color: colors.text, fontWeight: 500 }}>
            {t(tool.nameKey)}
          </Typography.Text>
        </Space>
      </Card>
    </Col>
  );

  // 사이드바에서 검색어가 있을 때 표시될 화면
  const renderSearchResults = () => (
    <Card
      title={<Typography.Title level={4} style={{ color: colors.text }}>{t('homepage.search_results')}</Typography.Title>}
      style={{ marginBottom: 24, backgroundColor: 'transparent', border: 'none' }}
      headStyle={{ borderBottom: `1px solid ${colors.border}`}}
    >
      {filteredTools.length > 0 ? (
        <Row gutter={[16, 16]}>{filteredTools.map(renderToolCard)}</Row>
      ) : (
        <Empty description={<Typography.Text style={{ color: colors.text }}>{t('common_no_results', '검색 결과가 없습니다.')}</Typography.Text>} />
      )}
    </Card>
  );

  // 기본 화면 (검색어 없을 때)
  const renderDefaultView = () => (
    <>
      {/* 최근 사용 도구 */}
      {recentToolDetails.length > 0 && (
        <Card
          title={<Typography.Title level={4} style={{ color: colors.text }}>{t('homepage.recent')}</Typography.Title>}
          style={{ marginBottom: 32, backgroundColor: 'transparent', border: 'none' }}
          headStyle={{ borderBottom: `1px solid ${colors.border}`}}
        >
          <Row gutter={[16, 16]}>{recentToolDetails.map(renderToolCard)}</Row>
        </Card>
      )}

      {/* 카테고리별 도구 목록 */}
      {categories
        .filter(cat => cat.id !== 'home' && cat.id !== 'settings') // 홈, 설정 제외
        .map(category => {
          const categoryTools = tools.filter(tool => tool.category === category.id);
          if (categoryTools.length === 0) return null;

          return (
            <Card
              key={category.id}
              title={
                <Space align="center">
                  <span style={{ color: category.color, fontSize: '22px', marginTop: '4px' }}>{category.icon}</span>
                  <Typography.Title level={4} style={{ color: colors.text, margin: 0 }}>
                    {t(category.nameKey)}
                  </Typography.Title>
                  <Tag color="default" style={{ backgroundColor: colors.secondary, borderColor: colors.border, color: colors.text }}>
                    {categoryTools.length}
                  </Tag>
                </Space>
              }
              style={{ marginBottom: 32, backgroundColor: 'transparent', border: 'none' }}
              headStyle={{ borderBottom: `1px solid ${colors.border}`}}
            >
              <Row gutter={[16, 16]}>{categoryTools.map(renderToolCard)}</Row>
            </Card>
          );
        })}
    </>
  );

  return (
    <div style={{ padding: '24px 32px', height: '100%', overflowY: 'auto' }}>
      <div style={{ marginBottom: 32 }}>
        <Typography.Title level={2} style={{ color: colors.text }}>
          {t('common_app_name')}
        </Typography.Title>
        <Typography.Paragraph style={{ color: colors.text, fontSize: '16px' }}>
          {t('homepage.welcome')}
        </Typography.Paragraph>
      </div>
      
      {/* 검색어 유무에 따라 다른 뷰를 렌더링 */}
      {searchTerm ? renderSearchResults() : renderDefaultView()}
    </div>
  );
};

export default HomePage;
