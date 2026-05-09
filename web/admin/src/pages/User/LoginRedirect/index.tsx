import { useEffect } from 'react';

import { Button, Result } from 'antd';

import { redirectToFixedLogin } from '@/utils/auth';

export default function LoginRedirectPage() {
  useEffect(() => {
    redirectToFixedLogin();
  }, []);

  return (
    <Result
      status="info"
      title="会话已失效或尚未登录"
      subTitle="请跳转到统一登录入口继续访问管理台。"
      extra={
        <Button type="primary" onClick={redirectToFixedLogin}>
          前往登录
        </Button>
      }
    />
  );
}
