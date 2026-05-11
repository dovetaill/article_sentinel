import { Empty, Typography } from 'antd';

export default function WorkspaceEmptyPage() {
  return (
    <section className="admin-workspace-empty admin-light-surface">
      <Empty description="当前没有打开的工作标签" image={Empty.PRESENTED_IMAGE_SIMPLE}>
        <Typography.Text type="secondary">请从左侧菜单重新进入业务页面。</Typography.Text>
      </Empty>
    </section>
  );
}
