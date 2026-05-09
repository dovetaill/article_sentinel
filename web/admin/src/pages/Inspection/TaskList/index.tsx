import { PageContainer, ProCard } from '@ant-design/pro-components';
import { Empty } from 'antd';

export default function TaskListPage() {
  return (
    <PageContainer title={false}>
      <ProCard>
        <Empty description="检测任务页将在后续任务中迁移。" />
      </ProCard>
    </PageContainer>
  );
}
