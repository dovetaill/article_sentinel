import { Button, Result } from 'antd';

export default function LoginRedirectPage() {
  return (
    <Result
      status="info"
      title="会话已失效或尚未登录"
      subTitle="请跳转到统一登录入口继续访问管理台。"
      extra={
        <Button type="primary" href="/auth/login">
          前往登录
        </Button>
      }
    />
  );
}
