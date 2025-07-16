import React, { useState, useEffect, useMemo } from 'react';
import { Input, Row, Col, Typography, Checkbox, ConfigProvider, theme as antdTheme } from 'antd';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import { GlobalStyle, ToolContainer, StyledCard } from '../components/common/ToolStyles';
import styled from 'styled-components';
import debounce from 'lodash.debounce';
import Prism from 'prismjs';
import 'prismjs/components/prism-regex';
import 'prismjs/themes/prism.css';

const { TextArea } = Input;
const { Title, Text } = Typography;
const CheckboxGroup = Checkbox.Group;

const HighlightedText = styled.div`
  white-space: pre-wrap;
  word-wrap: break-word;
  .match {
    background-color: var(--highlight-bg-color);
    color: var(--highlight-text-color);
    font-weight: bold;
  }
  code.language-regex {
    background: transparent;
    border: none;
    font-size: 1em;
  }
`;

// HTML escape to prevent XSS
const escapeHtml = (text: string) =>
  text.replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

const RegexTester = () => {
  const { t } = useLanguage();
  const { theme } = useTheme();
  const [pattern, setPattern] = useState('');
  const [testString, setTestString] = useState('');
  const [flags, setFlags] = useState(['g', 'i']);
  const [error, setError] = useState('');

  // Debounced input handlers for performance
  const onPatternChange = useMemo(
    () => debounce((value: string) => setPattern(value), 300),
    []
  );
  const onTestChange = useMemo(
    () => debounce((value: string) => setTestString(value), 300),
    []
  );

  const antdConfig = useMemo(
    () => ({ algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm }),
    [theme]
  );

  useEffect(() => { document.body.className = theme; }, [theme]);

  const highlightedOutput = useMemo(() => {
    if (!pattern || !testString) {
      setError('');
      return <Text style={{ color: 'var(--text-color)' }}>{escapeHtml(testString)}</Text>;
    }

    try {
      const regex = new RegExp(pattern, flags.join(''));
      const matches = [...testString.matchAll(regex)];

      if (matches.length === 0) {
        setError('');
        return <Text style={{ color: 'var(--text-color)' }}>{escapeHtml(testString)}</Text>;
      }

      setError('');
      let resultHtml = '';
      let lastIndex = 0;

      matches.forEach((m, i) => {
        const [matchText] = m;
        const idx = m.index ?? 0;
        resultHtml += escapeHtml(testString.slice(lastIndex, idx));
        resultHtml += `<span class=\"match\">${escapeHtml(matchText)}</span>`;
        lastIndex = idx + matchText.length;
      });
      resultHtml += escapeHtml(testString.slice(lastIndex));

      return (
        <HighlightedText>
          {/* Syntax-highlight regex source */}
          <code className="language-regex">{pattern}</code>
          {/* Highlighted text */}
          <div dangerouslySetInnerHTML={{ __html: resultHtml }} />
        </HighlightedText>
      );
    } catch (e: any) {
      setError(e.message);
      return <Text style={{ color: 'var(--text-color)' }}>{escapeHtml(testString)}</Text>;
    }
  }, [pattern, testString, flags, theme]);

  // Prism highlight after render
  useEffect(() => { Prism.highlightAll(); }, [pattern, highlightedOutput]);

  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle />
      <ToolContainer>
        <Title level={2} style={{ color: 'var(--text-color)', margin: '0 0 16px 0' }}>{t('tools_regex')}</Title>

        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col flex={1}>
            <Input
              placeholder={t('regex_tester_placeholder_pattern')}
              onChange={(e) => onPatternChange(e.target.value)}
              style={{
                backgroundColor: 'var(--card-bg-color)',
                color: 'var(--text-color)',
                borderColor: 'var(--border-color)',
                caretColor: 'var(--text-color)'
              }}
            />
          </Col>
          <Col>
            <CheckboxGroup options={['g', 'i', 'm']} value={flags} onChange={(vals) => setFlags(vals as string[])} />
          </Col>
        </Row>
        {error && <Text type="danger" style={{ marginBottom: 16 }}>{error}</Text>}

        <Row gutter={[16, 16]} style={{ flex: 1, minHeight: 0 }}>
          <Col xs={24} md={12} style={{ display: 'flex', flexDirection: 'column' }}>
            <StyledCard title={t('regex_tester_test_string_label')}>
              <TextArea
                onChange={(e) => onTestChange(e.target.value)}
                placeholder={t('regex_tester_placeholder_test_string')}
                rows={10}
                style={{
                  flex: 1,
                  resize: 'none',
                  border: 'none',
                  backgroundColor: 'transparent',
                  color: 'var(--text-color)',
                  caretColor: 'var(--text-color)' 
                }}
              />
            </StyledCard>
          </Col>
          <Col xs={24} md={12} style={{ display: 'flex', flexDirection: 'column' }}>
            <StyledCard title={t('regex_tester_matches_label')}>
              {highlightedOutput}
            </StyledCard>
          </Col>
        </Row>
      </ToolContainer>
    </ConfigProvider>
  );
};

export default RegexTester;
