import { useEffect, useMemo, useRef, useState } from "react";
import {
  Typography,
  Row,
  Col,
  ConfigProvider,
  theme as antdTheme,
  Select,
  Button,
  Spin,
  Space,
  Switch,
  message,
  Tag,
} from "antd";
import { useLanguage } from "../contexts/LanguageContext";
import { useTheme } from "../contexts/ThemeContext";
import {
  GlobalStyle,
  ToolContainer,
  StyledCard,
} from "../components/common/ToolStyles";

const { Title, Text } = Typography;
const { Option } = Select;

/**
 * Format a date with current locale + local timezone suffix.
 * Returns "--" when null.
 */
const formatDate = (d: Date | null) =>
  d
    ? `${d.toLocaleString()} (${Intl.DateTimeFormat().resolvedOptions().timeZone})`
    : "--";

const TimeTool = () => {
  const { t } = useLanguage();
  const { theme } = useTheme();

  /* ------------------------------------------------- STATE */
  const [pcTime, setPcTime] = useState<Date>(new Date());

  const [ntpBaseTime, setNtpBaseTime] = useState<Date | null>(null); // 서버에서 받은 시각
  const [ntpFetchedAt, setNtpFetchedAt] = useState<number | null>(null); // 해당 시각을 수신한 PC 시각

  const [loading, setLoading] = useState<boolean>(false);
  const [latency, setLatency] = useState<number | null>(null);

  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  const [refreshInterval, setRefreshInterval] = useState<number>(60); // sec
  const [selectedNtpServer, setSelectedNtpServer] = useState<string>(
    "time.cloudflare.com",
  );

  const refreshTimer = useRef<NodeJS.Timeout>();

  /* ------------------------------------------------- CONST */
  const ntpServers = [
    "time.cloudflare.com",
    "time.google.com",
    "time.facebook.com",
    "time.windows.com",
    "time.apple.com",
    "time2.kriss.re.kr",
  ];

  const intervals = [15, 30, 60, 300];

  const antdConfig = useMemo(
    () => ({ algorithm: theme === "dark" ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm }),
    [theme],
  );

  /* ------------------------------------------------- HELPERS */
  /** 최신 NTP 시각 계산 (tab background 일 때 interval throttle되어도 정확) */
  const currentNtpTime = useMemo(() => {
    if (!ntpBaseTime || !ntpFetchedAt) return null;
    return new Date(ntpBaseTime.getTime() + (Date.now() - ntpFetchedAt));
  }, [ntpBaseTime, ntpFetchedAt, pcTime]); // pcTime이 주기적 re-render 트리거

  const diffSec = useMemo(() => {
    if (!currentNtpTime) return null;
    return (currentNtpTime.getTime() - Date.now()) / 1000; // 실시간 오차(초) – 소수 포함
  }, [currentNtpTime]);

  const diffColor = useMemo(() => {
    if (diffSec === null) return "default";
    const abs = Math.abs(diffSec);
    if (abs <= 0.1) return "success"; // ±0.10 초 이내 OK
    if (abs <= 1) return "warning"; // ±1 초 주의
    return "error"; // 그 이상 오류
  }, [diffSec]);

  /* ------------------------------------------------- NTP FETCH */
  const fetchNtpTime = async (server: string) => {
    setLoading(true);
    const start = Date.now();
    try {
      const time: Date = await window.ipcRenderer.getNtpTime(server);
      const end = Date.now();
      setLatency(end - start);
      setNtpBaseTime(time);
      setNtpFetchedAt(end);
    } catch (error) {
      console.error("Failed to fetch NTP time:", error);
      message.error(t("Failed to fetch NTP time"));
      setNtpBaseTime(null);
      setNtpFetchedAt(null);
      setLatency(null);
    } finally {
      setLoading(false);
    }
  };

  /* ------------------------------------------------- AUTO REFRESH */
  useEffect(() => {
    if (refreshTimer.current) clearInterval(refreshTimer.current);

    if (autoRefresh) {
      refreshTimer.current = setInterval(() => {
        fetchNtpTime(selectedNtpServer);
      }, refreshInterval * 1000);
    }

    return () => {
      if (refreshTimer.current) clearInterval(refreshTimer.current);
    };
  }, [autoRefresh, refreshInterval, selectedNtpServer]);

  /* ------------------------------------------------- 최초 1회 */
  useEffect(() => {
    fetchNtpTime(selectedNtpServer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  /* ------------------------------------------------- PC CLOCK TICK  (250 ms) */
  useEffect(() => {
    const tick = setInterval(() => {
      setPcTime(new Date());
    }, 250);
    return () => clearInterval(tick);
  }, []);

  /* ------------------------------------------------- RENDER */
  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle />
      <ToolContainer>
        <Title level={2} style={{ margin: "0 0 24px 0" }}>
          {t("tools_time")}
        </Title>

        {/* ---------------- SETTINGS CARD ---------------- */}
        <StyledCard title={t("NTP Settings") as string} style={{ marginBottom: 24 }}>
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            <Row gutter={[16, 16]} align="middle">
              <Col flex="auto">
                <Select
                  value={selectedNtpServer}
                  onChange={setSelectedNtpServer}
                  style={{ width: "100%" }}
                >
                  {ntpServers.map((server) => (
                    <Option key={server} value={server}>
                      {server}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col>
                <Button type="primary" onClick={() => fetchNtpTime(selectedNtpServer)} loading={loading}>
                  {t("Sync Now")}
                </Button>
              </Col>
            </Row>

            <Row gutter={[16, 16]} align="middle">
              <Col>
                <Switch checked={autoRefresh} onChange={setAutoRefresh} />
              </Col>
              <Col flex="none">
                <Text>{t("Auto refresh")}</Text>
              </Col>
              <Col flex="auto">
                <Select
                  disabled={!autoRefresh}
                  value={refreshInterval}
                  onChange={setRefreshInterval}
                  style={{ width: 120 }}
                >
                  {intervals.map((sec) => (
                    <Option key={sec} value={sec}>
                      {sec} s
                    </Option>
                  ))}
                </Select>
              </Col>
            </Row>
          </Space>
        </StyledCard>

        {/* ---------------- TIMES ---------------- */}
        <Row gutter={[24, 24]}>
          <Col xs={24} md={12}>
            <StyledCard title={t("NTP Time") as string}>
              {loading ? (
                <Spin size="large" />
              ) : (
                <Space direction="vertical" size="small">
                  <Text style={{ fontSize: "1.5em" }}>{formatDate(currentNtpTime)}</Text>
                  {latency !== null && (
                    <Text type="secondary">{t("Latency")}: {latency} ms</Text>
                  )}
                </Space>
              )}
            </StyledCard>
          </Col>

          <Col xs={24} md={12}>
            <StyledCard title={t("PC Time") as string}>
              <Space direction="vertical" size="small">
                <Text style={{ fontSize: "1.5em" }}>{formatDate(pcTime)}</Text>
                {diffSec !== null && (
                  <Space>
                    <Text>{t("Difference")}: </Text>
                    <Tag color={diffColor}>{diffSec.toFixed(2)} s</Tag>
                  </Space>
                )}
              </Space>
            </StyledCard>
          </Col>
        </Row>
      </ToolContainer>
    </ConfigProvider>
  );
};

export default TimeTool;
