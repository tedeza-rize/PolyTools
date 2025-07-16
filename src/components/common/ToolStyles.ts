import styled, { createGlobalStyle } from 'styled-components';
import { Card, Row, Input } from 'antd';

const { TextArea } = Input;

export const GlobalStyle = createGlobalStyle`
  body.dark {
    --bg-color: #1a1a1a;
    --text-color: #e0e0e0;
    --card-bg-color: #2a2a2a;
    --border-color: #424242;
    --highlight-bg-color: #4a4a2a;
    --highlight-text-color: #ffff00;
  }
  body.light {
    --bg-color: #f5f5f5;
    --text-color: #212121;
    --card-bg-color: #ffffff;
    --border-color: #e0e0e0;
    --highlight-bg-color: #fffb8f;
    --highlight-text-color: #000000;
  }
`;

export const ToolContainer = styled.div`
  padding: 24px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  color: var(--text-color);
  box-sizing: border-box;
`;

export const ResponsiveRow = styled(Row)`
  flex-grow: 1;
  min-height: 0;
`;

export const StyledCard = styled(Card)`
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

export const StyledTextArea = styled(TextArea)`
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
