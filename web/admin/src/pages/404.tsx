import { Button, Result } from 'antd';
import { history } from '@umijs/max';

export default function NotFoundPage() {
  return (
    <Result
      status="404"
      title="404"
      subTitle="未找到对应的管理台页面。"
      extra={
        <Button type="primary" onClick={() => history.push('/inspection/tasks')}>
          返回检测任务
        </Button>
      }
    />
  );
}
