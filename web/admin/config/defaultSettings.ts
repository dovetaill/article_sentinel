import type { Settings as LayoutSettings } from '@ant-design/pro-layout';

const settings: LayoutSettings & {
  title?: string;
} = {
  title: '文章哨兵管理台',
  navTheme: 'dark',
  layout: 'mix',
  contentWidth: 'Fluid',
  fixSiderbar: true,
  fixedHeader: true,
  splitMenus: false
};

export default settings;
