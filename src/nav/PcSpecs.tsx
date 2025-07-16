import { useEffect, useState, useMemo } from 'react';
import { Typography, Row, Col, ConfigProvider, theme as antdTheme } from 'antd';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import { GlobalStyle, ToolContainer, StyledCard } from '../components/common/ToolStyles';

const { Title, Text } = Typography;

declare global {
  interface Window {
    ipcRenderer: {
      getSystemInfo: () => Promise<any>;
    };
  }
}

const PcSpecs = () => {
  const { t } = useLanguage();
  const { theme } = useTheme();
  const [specs, setSpecs] = useState<any>(null);

  const antdConfig = useMemo(
    () => ({ algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm }),
    [theme]
  );

  useEffect(() => {
    const getSpecs = async () => {
      try {
        const systemInfo = await window.ipcRenderer.getSystemInfo();
        console.log('Received system info:', systemInfo);
        setSpecs(systemInfo);
      } catch (error) {
        console.error('Error fetching system info:', error);
      }
    };

    getSpecs();
  }, []);

  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle />
      <ToolContainer>
        <Title level={2} style={{ color: 'var(--text-color)', margin: '0 0 16px 0' }}>{t('tools_pc_specs')}</Title>
        <Row gutter={[16, 16]}>
          <Col xs={24}>
            <StyledCard title="CPU">
              {specs ? (
                <Text>{specs.cpu.manufacturer} {specs.cpu.brand} @ {specs.cpu.speed}GHz</Text>
              ) : (
                <Text>Loading...</Text>
              )}
            </StyledCard>
          </Col>
          <Col xs={24}>
            <StyledCard title="RAM">
              {specs ? (
                <Text>{Math.round(specs.mem.total / 1024 / 1024 / 1024)} GB</Text>
              ) : (
                <Text>Loading...</Text>
              )}
            </StyledCard>
          </Col>
          <Col xs={24}>
            <StyledCard title="OS">
              {specs ? (
                <Text>{specs.os.distro} {specs.os.release}</Text>
              ) : (
                <Text>Loading...</Text>
              )}
            </StyledCard>
          </Col>
          <Col xs={24}>
            <StyledCard title="GPU">
              {specs ? (
                specs.graphics.controllers.map((controller: any, index: number) => (
                  <Text key={index}>GPU {index + 1}: {controller.model}</Text>
                ))
              ) : (
                <Text>Loading...</Text>
              )}
            </StyledCard>
          </Col>
          <Col xs={24}>
            <StyledCard title="Disk">
              {specs ? (
                specs.disk.map((d: any, index: number) => (
                  <Text key={index}>Disk {index + 1}: {d.name} {Math.round(d.size / 1024 / 1024 / 1024)} GB</Text>
                ))
              ) : (
                <Text>Loading...</Text>
              )}
            </StyledCard>
          </Col>
        </Row>
      </ToolContainer>
    </ConfigProvider>
  );
};

export default PcSpecs;