import React, { useState, useEffect, useMemo } from 'react';
import { 
  Input, Button, Radio, Row, Col, Typography, Card, ConfigProvider, 
  Space, theme as antdTheme, Flex, Progress, Tooltip, Switch 
} from 'antd';
import { CopyOutlined, SyncOutlined, MoonOutlined, SunOutlined } from '@ant-design/icons';
import { diffLines, diffWords } from 'diff';
import * as fuzzball from 'fuzzball';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import styled, { createGlobalStyle } from 'styled-components';

const { TextArea } = Input;
const { Title, Text, Paragraph } = Typography;

// 개선된 GlobalStyle - 더 나은 다크모드 대응
const GlobalStyle = createGlobalStyle`
  body {
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  
  .diff-line {
    font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
    white-space: pre-wrap;
    padding: 8px 16px;
    border-radius: 4px;
    margin: 2px 0;
    transition: all 0.2s ease;
  }
  
  .diff-line-added {
    background-color: ${props => props.theme.isDark ? 'rgba(34, 197, 94, 0.15)' : 'rgba(34, 197, 94, 0.1)'};
    border-left: 3px solid #22c55e;
  }
  
  .diff-line-removed {
    background-color: ${props => props.theme.isDark ? 'rgba(239, 68, 68, 0.15)' : 'rgba(239, 68, 68, 0.1)'};
    border-left: 3px solid #ef4444;
  }
  
  .diff-word-added {
    background-color: ${props => props.theme.isDark ? 'rgba(34, 197, 94, 0.3)' : 'rgba(34, 197, 94, 0.2)'};
    padding: 2px 4px;
    border-radius: 4px;
    font-weight: 500;
  }
  
  .diff-word-removed {
    background-color: ${props => props.theme.isDark ? 'rgba(239, 68, 68, 0.3)' : 'rgba(239, 68, 68, 0.2)'};
    text-decoration: line-through;
    padding: 2px 4px;
    border-radius: 4px;
    font-weight: 500;
  }
`;

const PageContainer = styled.div`
  padding: 24px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: ${props => props.theme.isDark ? 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)' : 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)'};
  transition: background 0.3s ease;
`;

const HeaderContainer = styled(Flex)`
  background: ${props => props.theme.isDark ? 'rgba(30, 41, 59, 0.8)' : 'rgba(255, 255, 255, 0.8)'};
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 16px 24px;
  margin-bottom: 24px;
  border: 1px solid ${props => props.theme.isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'};
  box-shadow: ${props => props.theme.isDark ? '0 4px 6px -1px rgba(0, 0, 0, 0.3)' : '0 4px 6px -1px rgba(0, 0, 0, 0.1)'};
`;

const MainContent = styled(Flex)`
  flex-grow: 1;
`;

const StyledTextArea = styled(TextArea)`
  height: 100% !important;
  min-height: 200px;
  resize: none;
  font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
  font-size: 14px;
  line-height: 1.6;
  border-radius: 12px !important;
  border: 2px solid ${props => props.theme.isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'} !important;
  background: ${props => props.theme.isDark ? 'rgba(30, 41, 59, 0.5)' : 'rgba(255, 255, 255, 0.8)'} !important;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease !important;
  
  &:hover {
    border-color: ${props => props.theme.isDark ? 'rgba(99, 102, 241, 0.5)' : 'rgba(99, 102, 241, 0.3)'} !important;
    box-shadow: ${props => props.theme.isDark ? '0 0 0 3px rgba(99, 102, 241, 0.1)' : '0 0 0 3px rgba(99, 102, 241, 0.05)'};
  }
  
  &:focus {
    border-color: #6366f1 !important;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1) !important;
  }
`;

const ResultCard = styled(Card)`
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  border-radius: 12px !important;
  border: 2px solid ${props => props.theme.isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'} !important;
  background: ${props => props.theme.isDark ? 'rgba(30, 41, 59, 0.5)' : 'rgba(255, 255, 255, 0.8)'} !important;
  backdrop-filter: blur(10px);
  box-shadow: ${props => props.theme.isDark ? '0 8px 32px rgba(0, 0, 0, 0.3)' : '0 8px 32px rgba(0, 0, 0, 0.1)'};
  
  .ant-card-head {
    background: transparent !important;
    border-bottom: 1px solid ${props => props.theme.isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'} !important;
  }
  
  .ant-card-body {
    flex-grow: 1;
    overflow: auto;
    padding: 16px !important;
    background: transparent !important;
  }
`;

const TextAreaContainer = styled.div`
  position: relative;
  display: flex;
  flex-direction: column;
  
  .textarea-label {
    position: absolute;
    top: -10px;
    left: 16px;
    background: ${props => props.theme.isDark ? '#0f172a' : '#ffffff'};
    padding: 0 8px;
    font-size: 12px;
    font-weight: 500;
    color: ${props => props.theme.isDark ? '#94a3b8' : '#64748b'};
    z-index: 1;
  }
`;

const ControlsContainer = styled(Space)`
  background: ${props => props.theme.isDark ? 'rgba(15, 23, 42, 0.8)' : 'rgba(248, 250, 252, 0.8)'};
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid ${props => props.theme.isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'};
`;

const StyledButton = styled(Button)`
  border-radius: 8px !important;
  font-weight: 500 !important;
  transition: all 0.2s ease !important;
  
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
`;

const ThemeToggle = styled(Switch)`
  .ant-switch-handle {
    &::before {
      background: ${props => props.checked ? '#fbbf24' : '#60a5fa'} !important;
    }
  }
`;

const DifferTool: React.FC = () => {
  const { t } = useLanguage();
  const { theme, toggleTheme } = useTheme();
  const { token } = antdTheme.useToken();
  const isDark = theme === 'dark';

  const [leftText, setLeftText] = useState('React is a JavaScript library for building user interfaces.\nIt lets you compose complex UIs from small and isolated pieces of code called "components".');
  const [rightText, setRightText] = useState('React is a powerful JavaScript library for creating interactive user interfaces.\nIt allows you to build complex UIs from reusable pieces of code called "components".');
  const [engine, setEngine] = useState<'diff' | 'fuzzball'>('diff');
  const [diffType, setDiffType] = useState<'lines'>('lines');

  // 개선된 Ant Design 테마 설정
  const antdConfig = useMemo(() => ({
    token: { 
      colorPrimary: '#6366f1',
      borderRadius: 8,
      colorBgContainer: isDark ? 'rgba(30, 41, 59, 0.8)' : 'rgba(255, 255, 255, 0.8)',
      colorText: isDark ? '#f1f5f9' : '#334155',
      colorBorder: isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'
    },
    algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
  }), [isDark]);

  const { diffResult, similarity } = useMemo(() => {
    if (engine === 'diff') {
      const options = { newlineIsToken: true };
      const result = diffType === 'words' 
        ? diffWords(leftText, rightText) 
        : diffLines(leftText, rightText, options);
      return { diffResult: result, similarity: null };
    } else {
      const ratio = fuzzball.ratio(leftText, rightText);
      return { diffResult: [], similarity: ratio };
    }
  }, [leftText, rightText, diffType, engine]);

  const copyResult = () => {
    let textToCopy = '';
    if (engine === 'diff') {
      textToCopy = diffResult.map(part => part.value).join('');
    } else if (similarity !== null) {
      textToCopy = `${t('differ_tool_similarity')}: ${similarity}%`;
    }
    navigator.clipboard.writeText(textToCopy);
  };

  const renderDiffResult = () => {
    if (!leftText && !rightText) {
      return (
        <div style={{ 
          padding: '48px', 
          textAlign: 'center',
          color: isDark ? '#94a3b8' : '#64748b'
        }}>
          <Text type="secondary" style={{ fontSize: '16px' }}>
            {t('differ_tool_start_prompt') || 'Enter text in both fields to see the differences'}
          </Text>
        </div>
      );
    }

    if (diffType === 'lines') {
      return (
        <div style={{ fontFamily: 'JetBrains Mono, monospace' }}>
          {diffResult.map((part, index) => {
            const lines = part.value.replace(/\n$/, '').split('\n');
            return lines.map((line, lineIndex) => (
              <Flex 
                key={`${index}-${lineIndex}`} 
                align="flex-start" 
                className={`diff-line ${part.added ? 'diff-line-added' : ''} ${part.removed ? 'diff-line-removed' : ''}`}
              >
                <Text 
                  type="secondary" 
                  style={{ 
                    width: 40, 
                    textAlign: 'center', 
                    userSelect: 'none',
                    fontWeight: 'bold',
                    color: part.added ? '#22c55e' : part.removed ? '#ef4444' : isDark ? '#64748b' : '#94a3b8'
                  }}
                >
                  {part.added ? '+' : part.removed ? '-' : ' '}
                </Text>
                <div style={{ flex: 1, paddingLeft: '8px' }}>{line}</div>
              </Flex>
            ));
          })}
        </div>
      );
    } else {
      return (
        <Paragraph style={{ 
          padding: '16px', 
          fontFamily: 'JetBrains Mono, monospace',
          fontSize: '14px',
          lineHeight: '1.6'
        }}>
          {diffResult.map((part, index) => (
            <span key={index} className={part.added ? 'diff-word-added' : part.removed ? 'diff-word-removed' : ''}>
              {part.value}
            </span>
          ))}
        </Paragraph>
      );
    }
  };

  const renderSimilarityResult = () => (
    <Flex vertical align="center" justify="center" style={{ height: '100%', minHeight: 200 }}>
      <Progress 
        type="dashboard" 
        percent={similarity ?? 0} 
        strokeColor={{
          '0%': '#6366f1',
          '100%': '#8b5cf6',
        }}
        trailColor={isDark ? 'rgba(71, 85, 105, 0.3)' : 'rgba(226, 232, 240, 0.8)'}
        size={160}
        strokeWidth={8}
      />
      <Title level={3} style={{ 
        marginTop: 24, 
        color: isDark ? '#f1f5f9' : '#334155',
        fontWeight: 600
      }}>
        {t('differ_tool_similarity') || 'Similarity'}: {similarity ?? 0}%
      </Title>
    </Flex>
  );

  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle theme={{ isDark }} />
      <PageContainer theme={{ isDark }}>
        <HeaderContainer justify="space-between" align="center" theme={{ isDark }}>
          <Title level={2} style={{ margin: 0, color: isDark ? '#f1f5f9' : '#334155' }}>
            {t('differ_tool_title') || 'Text Differ'}
          </Title>
          <Space size="middle">
            <Tooltip title={isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'}>
              <ThemeToggle
                checked={isDark}
                onChange={toggleTheme}
                checkedChildren={<MoonOutlined />}
                unCheckedChildren={<SunOutlined />}
              />
            </Tooltip>
            <Tooltip title={t('differ_tool_swap_tooltip') || 'Swap text fields'}>
              <StyledButton 
                shape="circle" 
                icon={<SyncOutlined />} 
                onClick={() => { 
                  setLeftText(rightText); 
                  setRightText(leftText); 
                }}
              />
            </Tooltip>
          </Space>
        </HeaderContainer>

        <MainContent vertical gap={24}>
          <Row gutter={24} style={{ flex: '0 0 auto' }}>
            <Col span={12}>
              <TextAreaContainer theme={{ isDark }}>
                <div className="textarea-label">Original Text</div>
                <StyledTextArea 
                  value={leftText} 
                  onChange={e => setLeftText(e.target.value)} 
                  placeholder={t('differ_tool_placeholder_left') || 'Enter original text here...'}
                  theme={{ isDark }}
                />
              </TextAreaContainer>
            </Col>
            <Col span={12}>
              <TextAreaContainer theme={{ isDark }}>
                <div className="textarea-label">Modified Text</div>
                <StyledTextArea 
                  value={rightText} 
                  onChange={e => setRightText(e.target.value)} 
                  placeholder={t('differ_tool_placeholder_right') || 'Enter modified text here...'}
                  theme={{ isDark }}
                />
              </TextAreaContainer>
            </Col>
          </Row>

          <ResultCard 
            theme={{ isDark }}
            title={
              <Flex justify="space-between" align="center" style={{ width: '100%' }}>
                <ControlsContainer wrap theme={{ isDark }}>
                  <Text strong style={{ color: isDark ? '#f1f5f9' : '#334155' }}>
                    {t('differ_tool_result') || 'Result'}
                  </Text>
                  <Radio.Group value={engine} onChange={e => setEngine(e.target.value)} buttonStyle="solid">
                    <Radio.Button value="diff">
                      {t('differ_tool_engine_diff') || 'Diff'}
                    </Radio.Button>
                    <Radio.Button value="fuzzball">
                      {t('differ_tool_engine_fuzzball') || 'Similarity'}
                    </Radio.Button>
                  </Radio.Group>
                  {engine === 'diff' && (
                    <Radio.Group value={diffType} onChange={e => setDiffType(e.target.value)} buttonStyle="solid">
                      <Radio.Button value="lines">
                        {t('differ_tool_by_lines') || 'Lines'}
                      </Radio.Button>
                      <Radio.Button value="words">
                        {t('differ_tool_by_words') || 'Words'}
                      </Radio.Button>
                    </Radio.Group>
                  )}
                </ControlsContainer>
                <Tooltip title={t('differ_tool_copy_tooltip') || 'Copy result to clipboard'}>
                  <StyledButton 
                    icon={<CopyOutlined />} 
                    onClick={copyResult}
                    type="primary"
                  >
                    {t('differ_tool_copy') || 'Copy'}
                  </StyledButton>
                </Tooltip>
              </Flex>
            }
          >
            {engine === 'diff' ? renderDiffResult() : renderSimilarityResult()}
          </ResultCard>
        </MainContent>
      </PageContainer>
    </ConfigProvider>
  );
};

export default DifferTool;