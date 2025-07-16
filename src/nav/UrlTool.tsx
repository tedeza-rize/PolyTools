import React, { useState, useEffect, useMemo } from 'react';
import { Input, Button, Radio, message, Row, Col, Typography, Switch, ConfigProvider, theme as antdTheme, Space } from 'antd';
import { CopyOutlined, SwapOutlined } from '@ant-design/icons';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import { GlobalStyle, ToolContainer, ResponsiveRow, StyledCard, StyledTextArea } from '../components/common/ToolStyles';

const { Title, Text } = Typography;

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
