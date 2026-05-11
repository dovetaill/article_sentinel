import type { Settings as LayoutSettings } from '@ant-design/pro-layout';

const settings: LayoutSettings & {
  title?: string;
  colorPrimary?: string;
  siderWidth?: number;
} = {
  title: '文章哨兵管理台',
  navTheme: 'light',
  layout: 'side',
  colorPrimary: '#1677ff',
  contentWidth: 'Fluid',
  fixSiderbar: true,
  fixedHeader: true,
  siderWidth: 232,
  splitMenus: false
};

export default settings;
