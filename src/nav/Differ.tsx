/* DifferTool.tsx */
import React, {
  useState,
  useMemo,
  useCallback,
  Key,
  ReactNode,
} from "react";
import {
  Input,
  Button,
  Radio,
  Row,
  Col,
  Typography,
  Card,
  ConfigProvider,
  Space,
  Grid,
  theme as antdTheme,
  Flex,
  Progress,
  Tooltip,
  Dropdown,
  message,
} from "antd";
import type { MenuProps } from "antd";
import {
  CopyOutlined,
  SyncOutlined,
  DownOutlined,
  CalculatorOutlined,
} from "@ant-design/icons";
import { diffLines, diffWords } from "diff";
import * as fuzzball from "fuzzball";
import { useLanguage } from "../contexts/LanguageContext";
import { useTheme } from "../contexts/ThemeContext";
import styled, { createGlobalStyle } from "styled-components";
import Tokenizer from "wink-tokenizer";
import lcs from "longest-common-subsequence";
import damerau from "damerau-levenshtein";

/* ------------------------------------------------------------------ */
/* ─────────────────────── 공통 상수 & 싱글턴 ──────────────────────── */
const singletonTokenizer = new Tokenizer();
const { TextArea } = Input;
const { Title, Text, Paragraph } = Typography;
const { useBreakpoint } = Grid;
const LINE_NUM_WIDTH = "64px";
const SIGN_WIDTH = "24px";

/* ------------------------------------------------------------------ */
/* ─────────────────────────── 전역 스타일 ─────────────────────────── */
const GlobalStyle = createGlobalStyle`
  body {
    transition: background-color 0.3s ease, color 0.3s ease;
  }
  .diff-line {
    display: grid;
    grid-template-columns: ${LINE_NUM_WIDTH} ${LINE_NUM_WIDTH} ${SIGN_WIDTH} 1fr;
    column-gap: 8px;
    padding: 4px 12px;
    font-family: 'JetBrains Mono', monospace;
    white-space: pre-wrap;
    border-left: 3px solid transparent;
    border-radius: 4px;
  }
  .diff-line-added {
    background-color: ${(p) =>
      p.theme.isDark ? "rgba(34,197,94,.15)" : "rgba(34,197,94,.1)"};
    border-left-color: #22c55e;
  }
  .diff-line-removed {
    background-color: ${(p) =>
      p.theme.isDark ? "rgba(239,68,68,.15)" : "rgba(239,68,68,.1)"};
    border-left-color: #ef4444;
  }
  .diff-word-added {
    background-color: ${(p) =>
      p.theme.isDark ? "rgba(34,197,94,.3)" : "rgba(34,197,94,.2)"};
    padding: 2px 4px;
    border-radius: 4px;
    font-weight: 500;
  }
  .diff-word-removed {
    background-color: ${(p) =>
      p.theme.isDark ? "rgba(239,68,68,.3)" : "rgba(239,68,68,.2)"};
    text-decoration: line-through;
    padding: 2px 4px;
    border-radius: 4px;
    font-weight: 500;
  }
  .diff-num {
    text-align: right;
    font-weight: 600;
    color: ${(p) => (p.theme.isDark ? "#64748b" : "#94a3b8")};
    user-select: none;
  }
  .diff-sign {
    text-align: center;
    font-weight: 700;
  }
  .diff-blank {
    visibility: hidden;
  }

  /* ───── 작은 화면(UI 보정) ───── */
  @media (max-width: 767px) {
    .diff-line { column-gap: 4px; padding: 2px 8px; font-size: 13px; }
    .diff-num { font-size: 12px; }
  }
`;

/* ------------------------------------------------------------------ */
/* ───────────────────────── styled-components ─────────────────────── */
const PageContainer = styled.div`
  padding: 24px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: ${(p) =>
    p.theme.isDark
      ? "linear-gradient(135deg,#0f172a 0%,#1e293b 100%)"
      : "linear-gradient(135deg,#f8fafc 0%,#e2e8f0 100%)"};
  transition: background 0.3s ease;
`;

const MainContent = styled(Flex)`
  flex-grow: 1;
`;

const ResultCard = styled(Card)`
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  border-radius: 12px !important;
  border: 2px solid
    ${(p) =>
      p.theme.isDark ? "rgba(71,85,105,.3)" : "rgba(226,232,240,.8)"} !important;
  background: ${(p) =>
    p.theme.isDark ? "rgba(30,41,59,.5)" : "rgba(255,255,255,.8)"} !important;
  backdrop-filter: blur(10px);
  box-shadow: ${(p) =>
    p.theme.isDark
      ? "0 8px 32px rgba(0,0,0,.3)"
      : "0 8px 32px rgba(0,0,0,.1)"};

  .ant-card-head {
    background: transparent !important;
    border-bottom: 1px solid
      ${(p) =>
        p.theme.isDark
          ? "rgba(71,85,105,.3)"
          : "rgba(226,232,240,.8)"} !important;
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
    background: ${(p) => (p.theme.isDark ? "#0f172a" : "#ffffff")};
    padding: 0 8px;
    font-size: 12px;
    font-weight: 500;
    color: ${(p) => (p.theme.isDark ? "#94a3b8" : "#64748b")};
    z-index: 1;
  }
`;

/* 반응형: column / row 전환 */
const ControlsContainer = styled(Space)<{ $mobile: boolean }>`
  flex-wrap: ${({ $mobile }) => ($mobile ? "wrap" : "nowrap")};
  flex-direction: ${({ $mobile }) => ($mobile ? "column" : "row")};
  gap: ${({ $mobile }) => ($mobile ? "8px" : "16px")};
  align-items: ${({ $mobile }) => ($mobile ? "stretch" : "center")};
  background: ${(p) =>
    p.theme.isDark ? "rgba(15,23,42,.8)" : "rgba(248,250,252,.8)"};
  padding: ${({ $mobile }) => ($mobile ? "12px" : "8px 16px")};
  border-radius: 8px;
  border: 1px solid
    ${(p) =>
      p.theme.isDark ? "rgba(71,85,105,.3)" : "rgba(226,232,240,.8)"};
`;

const StyledButton = styled(Button)`
  border-radius: 8px !important;
  font-weight: 500 !important;
  transition: 0.2s;
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
`;

/* ------------------------------------------------------------------ */
/* ───────────────────────────── 타입 ──────────────────────────────── */
type Engine = "diff" | "fuzzball" | "damerau" | "lcs" | "cosine" | "all";

/* ------------------------------------------------------------------ */
/* ──────────────────────────── 컴포넌트 ──────────────────────────── */
const DifferTool: React.FC = () => {
  /* ----- 공용 컨텍스트 ----- */
  const { t } = useLanguage();
  const { theme } = useTheme();
  const isDark = theme === "dark";

  /* ----- 반응형 브레이크포인트 ----- */
  const screens = useBreakpoint();
  const isMobile = !screens.md; // md(≥768px) 미만 = mobile

  /* ----- 상태 ----- */
  const [leftText, setLeftText] = useState(
    'React is a JavaScript library for building user interfaces.\nIt lets you compose complex UIs from small and isolated pieces of code called "components".'
  );
  const [rightText, setRightText] = useState(
    'React is a powerful JavaScript library for creating interactive user interfaces.\nIt allows you to build complex UIs from reusable pieces of code called "components".'
  );
  const [engine, setEngine] = useState<Engine>("diff");
  const [diffType, setDiffType] = useState<"lines" | "words">("lines");
  const [similarity, setSimilarity] = useState<Record<Engine, number> | null>(
    null
  );

  /* ----- antd ConfigProvider 토큰 ----- */
  const antdConfig = useMemo(
    () => ({
      token: {
        colorPrimary: "#6366f1",
        borderRadius: 8,
        colorBgContainer: isDark
          ? "rgba(30,41,59,.8)"
          : "rgba(255,255,255,.8)",
        colorText: isDark ? "#f1f5f9" : "#334155",
        colorBorder: isDark
          ? "rgba(71,85,105,.3)"
          : "rgba(226,232,240,.8)",
      },
      algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    }),
    [isDark]
  );

  /* ------------------------------------------------------------------ */
  /* ──────────────────────── 유틸 함수 (Cosine) ─────────────────────── */
  const calculateTf = (tokens: string[]) => {
    const tf: Record<string, number> = {};
    tokens.forEach((token) => {
      tf[token] = (tf[token] || 0) + 1;
    });
    const len = tokens.length;
    Object.keys(tf).forEach((k) => (tf[k] /= len));
    return tf;
  };

  const calculateIdf = (docs: string[][]) => {
    const idf: Record<string, number> = {};
    const docCnt = docs.length;
    const allTokens = new Set(docs.flat());
    allTokens.forEach((token) => {
      const numDocs = docs.filter((d) => d.includes(token)).length;
      idf[token] = Math.log(docCnt / (1 + numDocs)) + 1;
    });
    return idf;
  };

  const createVector = (
    tf: Record<string, number>,
    idf: Record<string, number>
  ) => {
    const vec: Record<string, number> = {};
    Object.keys(tf).forEach((k) => (vec[k] = tf[k] * idf[k]));
    return vec;
  };

  const cosineSimilarity = (
    v1: Record<string, number>,
    v2: Record<string, number>
  ) => {
    let dot = 0;
    const union = new Set([...Object.keys(v1), ...Object.keys(v2)]);
    union.forEach((k) => (dot += (v1[k] || 0) * (v2[k] || 0)));
    const mag = (v: Record<string, number>) =>
      Math.sqrt(Object.values(v).reduce((s, x) => s + x * x, 0));
    const m1 = mag(v1);
    const m2 = mag(v2);
    return m1 === 0 || m2 === 0 ? 100 : (dot / (m1 * m2)) * 100;
  };

  /* ------------------------------------------------------------------ */
  /* ────────────────────── Similarity 계산 로직 ────────────────────── */
  const calcSimilarity = useCallback(
    (method: Engine) => {
      /* 토큰화 1회만 수행 */
      const lt = singletonTokenizer
        .tokenize(leftText)
        .map((t) => t.value.toLowerCase());
      const rt = singletonTokenizer
        .tokenize(rightText)
        .map((t) => t.value.toLowerCase());

      switch (method) {
        case "fuzzball":
          return fuzzball.ratio(leftText, rightText);
        case "damerau": {
          const dist = damerau(leftText, rightText).steps;
          const longer = Math.max(leftText.length, rightText.length);
          return longer ? (1 - dist / longer) * 100 : 100;
        }
        case "lcs": {
          const ls = leftText.toLowerCase();
          const rs = rightText.toLowerCase();
          const lcsLen = lcs(ls, rs).length;
          const avg = (ls.length + rs.length) / 2;
          return avg ? (lcsLen / avg) * 100 : 100;
        }
        case "cosine": {
          const idf = calculateIdf([lt, rt]);
          return cosineSimilarity(
            createVector(calculateTf(lt), idf),
            createVector(calculateTf(rt), idf)
          );
        }
        default:
          return 0;
      }
    },
    [leftText, rightText]
  );

  /* ------------------------------------------------------------------ */
  /* ─────────────────────────── Diff 결과 ──────────────────────────── */
  const { diffResult } = useMemo(() => {
    if (engine !== "diff") return { diffResult: [] };
    const opts = { newlineIsToken: true } as const;
    return {
      diffResult:
        diffType === "words"
          ? diffWords(leftText, rightText)
          : diffLines(leftText, rightText, opts),
    };
  }, [leftText, rightText, diffType, engine]);

  /* ------------------------------------------------------------------ */
  /* ───────────────────────── 이벤트 핸들러 ────────────────────────── */
  const handleCompare = () => {
    if (engine === "diff") return;
    if (engine === "all") {
      const methods: Engine[] = ["fuzzball", "damerau", "lcs", "cosine"];
      const res = methods.reduce((acc, m) => {
        acc[m] = calcSimilarity(m);
        return acc;
      }, {} as Record<Engine, number>);
      setSimilarity(res);
    } else {
      setSimilarity({ [engine]: calcSimilarity(engine) } as Record<
        Engine,
        number
      >);
    }
    message.success(t("differ_tool_calc_done") ?? "Calculation complete");
  };

  const copyResult = () => {
    let txt = "";
    if (engine === "diff") txt = diffResult.map((p) => p.value).join("");
    else if (similarity)
      txt = Object.entries(similarity)
        .map(
          ([m, v]) => `${t(`differ_tool_engine_${m}`)}: ${v.toFixed(2)}%`
        )
        .join("\n");
    navigator.clipboard.writeText(txt);
    message.success(t("differ_tool_copy_done") ?? "Copied!");
  };

  /* ------------------------------------------------------------------ */
  /* ───────────────────────── JSX 부분 분리 ────────────────────────── */
  /* Radio Group(모드) */
  const modeRadio = (
    <Radio.Group
      value={engine === "diff" ? "diff" : "similarity"}
      onChange={(e) => {
        const v = e.target.value;
        setEngine(v === "diff" ? "diff" : "fuzzball");
        setSimilarity(null);
      }}
      buttonStyle="solid"
    >
      <Radio.Button value="diff">
        {t("differ_tool_engine_diff") ?? "Diff"}
      </Radio.Button>
      <Radio.Button value="similarity">
        {t("differ_tool_engine_similarity") ?? "Similarity"}
      </Radio.Button>
    </Radio.Group>
  );

  /* diffType or similarity method 선택 파트 */
  const diffRadio =
    engine === "diff" ? (
      <Radio.Group
        value={diffType}
        onChange={(e) => setDiffType(e.target.value)}
        buttonStyle="solid"
      >
        <Radio.Button value="lines">
          {t("differ_tool_by_lines") ?? "Lines"}
        </Radio.Button>
        <Radio.Button value="words">
          {t("differ_tool_by_words") ?? "Words"}
        </Radio.Button>
      </Radio.Group>
    ) : (
      <Dropdown
        menu={{
          items: [
            {
              type: "group",
              label: t("differ_tool_group_edit_distance"),
              children: [
                { key: "fuzzball", label: t("differ_tool_engine_fuzzball") },
                { key: "damerau", label: t("differ_tool_engine_damerau") },
                { key: "lcs", label: t("differ_tool_engine_lcs") },
              ],
            },
            {
              type: "group",
              label: t("differ_tool_group_statistical"),
              children: [
                { key: "cosine", label: t("differ_tool_engine_cosine") },
              ],
            },
          ],
          onClick: ({ key }: { key: Key }) => setEngine(key as Engine),
        }}
      >
        <Button>
          {t(`differ_tool_engine_${engine}`) ?? "Select Method"}{" "}
          <DownOutlined />
        </Button>
      </Dropdown>
    );

  /* 비교 / 모든 방법 버튼 */
  const compareButtons: ReactNode =
    engine !== "diff" ? (
      <Space>
        <StyledButton
          type="primary"
          icon={<CalculatorOutlined />}
          onClick={handleCompare}
        >
          {t("differ_tool_compare")}
        </StyledButton>
        <StyledButton onClick={() => setEngine("all")}>
          {t("differ_tool_engine_all")}
        </StyledButton>
      </Space>
    ) : null;

  /* 결과: Diff or Similarity */
  const diffView = () => {
    if (!leftText && !rightText)
      return (
        <Flex
          align="center"
          justify="center"
          style={{ height: 200 }}
        >
          <Text type="secondary">
            {t("differ_tool_start_prompt") ??
              "Enter text in both fields to see the differences"}
          </Text>
        </Flex>
      );

    /* ------- line diff ------- */
    if (diffType === "lines") {
      let ln = 0,
        rn = 0;
      return (
        <div>
          {diffResult.flatMap((p, i) =>
            p.value.replace(/\n$/, "").split("\n").map((line, j) => {
              const add = p.added,
                del = p.removed;
              if (!add && !del) {
                ln++;
                rn++;
              } else if (add) rn++;
              else if (del) ln++;
              const mainNo = add ? rn : ln;
              return (
                <div
                  key={`${i}-${j}`}
                  className={`diff-line ${
                    add ? "diff-line-added" : del ? "diff-line-removed" : ""
                  }`}
                >
                  <span className="diff-num">{mainNo}</span>
                  <span className="diff-num diff-blank" />
                  <span
                    className="diff-sign"
                    style={{
                      color: add
                        ? "#22c55e"
                        : del
                        ? "#ef4444"
                        : "transparent",
                    }}
                  >
                    {add ? "+" : del ? "-" : ""}
                  </span>
                  <span>{line || " "}</span>
                </div>
              );
            })
          )}
        </div>
      );
    }

    /* ------- word diff ------- */
    return (
      <Paragraph
        style={{
          padding: 16,
          fontFamily: "JetBrains Mono,monospace",
          fontSize: 14,
          lineHeight: 1.6,
        }}
      >
        {diffResult.map((p, i) => (
          <span
            key={i}
            className={
              p.added
                ? "diff-word-added"
                : p.removed
                ? "diff-word-removed"
                : ""
            }
          >
            {p.value}
          </span>
        ))}
      </Paragraph>
    );
  };

  const similarityView = () => (
    <Flex
      vertical
      align="center"
      justify="center"
      style={{ height: "100%", minHeight: 200, padding: 24 }}
    >
      {similarity ? (
        Object.entries(similarity).map(([m, v]) => (
          <Flex
            vertical
            align="center"
            key={m}
            style={{ marginBottom: 24, width: "100%" }}
          >
            <Title
              level={4}
              style={{ color: isDark ? "#c7d2fe" : "#4338ca" }}
            >
              {t(`differ_tool_engine_${m}`)}
            </Title>
            <Progress
              percent={Number(v.toFixed(2))}
              strokeColor={{ "0%": "#6366f1", "100%": "#8b5cf6" }}
              trailColor={
                isDark ? "rgba(71,85,105,.3)" : "rgba(226,232,240,.8)"
              }
              size={120}
              strokeWidth={8}
              type="dashboard"
            />
            <Title
              level={3}
              style={{
                marginTop: 16,
                color: isDark ? "#f1f5f9" : "#334155",
                fontWeight: 600,
              }}
            >
              {v.toFixed(2)}%
            </Title>
          </Flex>
        ))
      ) : (
        <Text type="secondary">
          {t("differ_tool_similarity_prompt")}
        </Text>
      )}
    </Flex>
  );

  /* ------------------------------------------------------------------ */
  /* ───────────────────────────── 렌더 ─────────────────────────────── */
  return (
    <ConfigProvider theme={antdConfig}>
      <GlobalStyle theme={{ isDark }} />
      <PageContainer theme={{ isDark }}>
        {/* ─── 헤더 ─── */}
        <Flex align="center" justify="space-between" style={{ marginBottom: 24 }}>
          <div style={{ flex: 1, textAlign: "center" }}>
            <Title
              level={2}
              style={{ margin: 0, color: isDark ? "#f1f5f9" : "#334155" }}
            >
              {t("differ_tool_title") ?? "Text Differ"}
            </Title>
          </div>
          <Tooltip title={t("differ_tool_swap_tooltip") ?? "Swap text fields"}>
            <StyledButton
              shape="circle"
              icon={<SyncOutlined />}
              onClick={() => {
                setLeftText(rightText);
                setRightText(leftText);
              }}
            />
          </Tooltip>
        </Flex>

        {/* ─── 본문 ─── */}
        <MainContent vertical gap={24}>
          {/* 입력 영역 */}
          <Row gutter={24} style={{ flex: "0 0 auto" }}>
            <Col xs={24} md={12}>
              <TextAreaContainer theme={{ isDark }}>
                <div className="textarea-label">Original Text</div>
                <TextArea
                  value={leftText}
                  onChange={(e) => setLeftText(e.target.value)}
                  placeholder={
                    t("differ_tool_placeholder_left") ??
                    "Enter original text here..."
                  }
                  autoSize={{ minRows: 15, maxRows: 30 }}
                  style={{ minHeight: isMobile ? 280 : 450 }}
                />
              </TextAreaContainer>
            </Col>
            <Col xs={24} md={12}>
              <TextAreaContainer theme={{ isDark }}>
                <div className="textarea-label">Modified Text</div>
                <TextArea
                  value={rightText}
                  onChange={(e) => setRightText(e.target.value)}
                  placeholder={
                    t("differ_tool_placeholder_right") ??
                    "Enter modified text here..."
                  }
                  autoSize={{ minRows: 15, maxRows: 30 }}
                  style={{ minHeight: isMobile ? 280 : 450 }}
                />
              </TextAreaContainer>
            </Col>
          </Row>

          {/* 결과 카드 */}
          <ResultCard
            theme={{ isDark }}
            style={{ maxHeight: isMobile ? "60vh" : undefined }}
            title={
              <Flex
                justify="space-between"
                align={isMobile ? "stretch" : "center"}
                vertical={isMobile}
                gap={isMobile ? 12 : 0}
              >
                {/* 컨트롤 박스 */}
                <ControlsContainer $mobile={isMobile} theme={{ isDark }}>
                  {modeRadio}
                  {diffRadio}
                  {compareButtons}
                </ControlsContainer>

                {/* 복사 버튼 */}
                <Tooltip
                  title={
                    t("differ_tool_copy_tooltip") ?? "Copy result to clipboard"
                  }
                >
                  <StyledButton
                    icon={<CopyOutlined />}
                    onClick={copyResult}
                    type="primary"
                    block={isMobile}
                  >
                    {t("differ_tool_copy") ?? "Copy"}
                  </StyledButton>
                </Tooltip>
              </Flex>
            }
          >
            {engine === "diff" ? diffView() : similarityView()}
          </ResultCard>
        </MainContent>
      </PageContainer>
    </ConfigProvider>
  );
};

export default DifferTool;
