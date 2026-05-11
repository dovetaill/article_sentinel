import type { ThemeConfig } from 'antd';
import { theme } from 'antd';

export const adminVisualTokens = {
  colorPrimary: '#2d8cf0',
  sidebarBg: '#191a23',
  pageBg: '#f0f2f5',
  contentBg: '#f5f7f9',
  headerBg: '#ffffff',
  surfaceBg: '#ffffff',
  cardBg: '#ffffff',
  tableBg: '#ffffff',
  tableHeaderBg: '#fafafa',
  borderColor: '#e8eaec',
  borderColorSecondary: '#f0f0f0',
  textColor: '#17233d',
  textColorSecondary: '#808695',
  tabActiveBg: '#e6f4ff',
  tabActiveBorder: '#91caff'
} as const;

export const adminAntdTheme: ThemeConfig = {
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: adminVisualTokens.colorPrimary,
    colorBgBase: adminVisualTokens.surfaceBg,
    colorBgLayout: adminVisualTokens.pageBg,
    colorBgContainer: adminVisualTokens.surfaceBg,
    colorBorder: adminVisualTokens.borderColor,
    colorBorderSecondary: adminVisualTokens.borderColorSecondary,
    colorText: adminVisualTokens.textColor,
    colorTextSecondary: adminVisualTokens.textColorSecondary,
    colorSuccess: '#389e0d',
    colorSuccessBg: '#f6ffed',
    colorSuccessBorder: '#b7eb8f',
    colorWarning: '#d48806',
    colorWarningBg: '#fffbe6',
    colorWarningBorder: '#ffe58f',
    colorError: '#cf1322',
    colorErrorBg: '#fff1f0',
    colorErrorBorder: '#ffccc7',
    colorInfo: adminVisualTokens.colorPrimary,
    colorInfoBg: '#e6f4ff',
    colorInfoBorder: '#91caff',
    borderRadius: 8
  },
  components: {
    Layout: {
      bodyBg: adminVisualTokens.pageBg,
      headerBg: adminVisualTokens.headerBg
    },
    Table: {
      headerBg: adminVisualTokens.tableHeaderBg,
      rowHoverBg: adminVisualTokens.contentBg
    }
  }
};
