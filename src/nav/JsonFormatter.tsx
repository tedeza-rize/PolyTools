import { useState } from 'react';
import { Input, Button, message } from 'antd';
import { useLanguage } from '../contexts/LanguageContext';

const { TextArea } = Input;

const JsonFormatter = () => {
  const { t } = useLanguage();
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');

  const handleFormat = () => {
    try {
      if (input.trim() === '') {
        setOutput('');
        return;
      }
      const formatted = JSON.stringify(JSON.parse(input), null, 4);
      setOutput(formatted);
    } catch (error) {
      message.error('Invalid JSON');
      setOutput('');
    }
  };

  return (
    <div>
      <h1>{t('json_formatter_title')}</h1>
      <TextArea rows={10} value={input} onChange={(e) => setInput(e.target.value)} placeholder={t('json_formatter_placeholder_input')} />
      <Button type="primary" onClick={handleFormat} style={{ margin: '10px 0' }}>{t('json_formatter_button_format')}</Button>
      <TextArea rows={10} value={output} readOnly placeholder={t('json_formatter_placeholder_output')} />
    </div>
  );
};

export default JsonFormatter;