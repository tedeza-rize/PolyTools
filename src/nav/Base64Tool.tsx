import { useState } from 'react';
import { Input, Button, Radio, message } from 'antd';
import { useLanguage } from '../contexts/LanguageContext';

const { TextArea } = Input;

const Base64Tool = () => {
  const { t } = useLanguage();
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [mode, setMode] = useState('encode'); // 'encode' or 'decode'

  const handleConvert = () => {
    try {
      if (mode === 'encode') {
        setOutput(btoa(input));
      } else {
        setOutput(atob(input));
      }
    } catch (error) {
      message.error('Invalid input for ' + mode);
    }
  };

  return (
    <div>
      <h1>{t('base64_tool_title')}</h1>
      <Radio.Group onChange={(e) => setMode(e.target.value)} value={mode} style={{ marginBottom: 10 }}>
        <Radio.Button value="encode">{t('base64_tool_encode')}</Radio.Button>
        <Radio.Button value="decode">{t('base64_tool_decode')}</Radio.Button>
      </Radio.Group>
      <TextArea rows={10} value={input} onChange={(e) => setInput(e.target.value)} placeholder={t('base64_tool_placeholder_input')} />
      <Button type="primary" onClick={handleConvert} style={{ margin: '10px 0' }}>
        {mode === 'encode' ? t('base64_tool_encode') : t('base64_tool_decode')}
      </Button>
      <TextArea rows={10} value={output} readOnly placeholder={t('base64_tool_placeholder_output')} />
    </div>
  );
};

export default Base64Tool;