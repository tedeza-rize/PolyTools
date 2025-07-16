import React, { useState, useEffect, useMemo } from 'react';
import { Input, Button, Radio, message, Row, Col, Typography, Switch, Card, ConfigProvider, theme as antdTheme, Space } from 'antd';
import { CopyOutlined, SwapOutlined } from '@ant-design/icons';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import styled, { createGlobalStyle } from 'styled-components';

const { TextArea } = Input;
const { Title, Text } = Typography;

const GlobalStyle = createGlobalStyle`
  body.dark {
    --bg-color: #1a1a1a;
    --text-color: #e0e0e0;
    --card-bg-color: #2a2a2a;
    --border-color: #424242;
  }
  body.light {
    --bg-color: #f5f5f5;
    --text-color: #212121;
    --card-bg-color: #ffffff;
    --border-color: #e0e0e0;
  }
`;

const ToolContainer = styled.div`
  padding: 24px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  color: var(--text-color);
  box-sizing: border-box;
`;

const ResponsiveRow = styled(Row)`
  flex-grow: 1;
  min-height: 0;
`;

const StyledCard = styled(Card)`
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--card-bg-color) !important;
  border: 1px solid var(--border-color);
  .ant-card-head {
    color: var(--text-color);
    border-bottom: 1px solid var(--border-color);
  }
  .ant-card-body {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    padding: 0;
  }
`;

const StyledTextArea = styled(TextArea)`
  flex-grow: 1;
  resize: none;
  border: none;
  background-color: transparent !important;
  color: var(--text-color) !important;
  padding: 12px;
  &:focus {
    box-shadow: none !important;
  }
`;

const UrlTool = () => {
  const { t } = useLanguage();
  const { theme } = useTheme();
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [mode, setMode] = useState('encode');
  const [autoConvert, setAutoConvert] = useState(true);

  const antdConfig = useMemo(() => ({
    algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
  }), [theme]);

  useEffect(() => {
    document.body.className = theme;
  }, [theme]);

  const handleConvert = () => {
    if (!input) {
      setOutput('');
      return;
    }
    try {
      if (mode === 'encode') {
        setOutput(encodeURIComponent(input));
      } else {
        setOutput(decodeURIComponent(input));
      }
    } catch (error) {
      setOutput('');
      if (!autoConvert) {
        message.error(t('url_tool_invalid_input'));
      }
    }
  };

  useEffect(() => {
    if (autoConvert) {
      handleConvert();
    }
  }, [input, mode, autoConvert]);

  const handleCopy = () => {
    if (!output) return;
    navigator.clipboard.writeText(output);
    message.success(t('base64_tool_copied_message'));
  };

  const handleSwap = () => {
    setInput(output);
    setMode(mode === 'encode' ? 'decode' : 'encode');
  };

  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle />
      <ToolContainer>
        <Title level={2} style={{ color: 'var(--text-color)', margin: '0 0 16px 0' }}>{t('url_tool_title')}</Title>
        <Row justify="space-between" align="middle" style={{ marginBottom: 24 }}>
          <Col>
            <Radio.Group onChange={(e) => setMode(e.target.value)} value={mode}>
              <Radio.Button value="encode">{t('base64_tool_encode')}</Radio.Button>
              <Radio.Button value="decode">{t('base64_tool_decode')}</Radio.Button>
            </Radio.Group>
          </Col>
          <Col>
            <Space>
              <Text style={{ color: 'var(--text-color)' }}>{t('common_auto')}</Text>
              <Switch checked={autoConvert} onChange={setAutoConvert} />
              {!autoConvert && (
                <Button type="primary" onClick={handleConvert}>
                  {mode === 'encode' ? t('base64_tool_encode') : t('base64_tool_decode')}
                </Button>
              )}
            </Space>
          </Col>
        </Row>
        <ResponsiveRow gutter={[16, 16]}>
          <Col xs={24} md={12} lg={24} style={{ display: 'flex', flexDirection: 'column' }}>
            <StyledCard title={t('base64_tool_input_label')}>
              <StyledTextArea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder={t('base64_tool_placeholder_input')}
              />
            </StyledCard>
          </Col>
          <Col xs={24} md={12} lg={24} style={{ display: 'flex', flexDirection: 'column' }}>
            <StyledCard 
              title={t('base64_tool_output_label')}
              extra={
                <Space>
                  <Button icon={<SwapOutlined />} onClick={handleSwap} title={t('base64_tool_swap')} />
                  <Button icon={<CopyOutlined />} onClick={handleCopy} title={t('base64_tool_copy')} />
                </Space>
              }
            >
              <StyledTextArea
                value={output}
                readOnly
                placeholder={t('base64_tool_placeholder_output')}
              />
            </StyledCard>
          </Col>
        </ResponsiveRow>
      </ToolContainer>
    </ConfigProvider>
  );
};

export default UrlTool;
