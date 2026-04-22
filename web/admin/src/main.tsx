import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import 'antd/dist/reset.css';

import App from './App';
import './styles.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: '#2f4b6f',
          colorInfo: '#2f4b6f',
          colorSuccess: '#3b6c5f',
          colorWarning: '#9a6a2f',
          colorError: '#b33a3a',
          colorText: '#1f2937',
          colorTextSecondary: '#5b6778',
          colorBgLayout: '#f3f6fb',
          colorBgContainer: '#ffffff',
          borderRadius: 18,
          fontFamily: '"Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
        }
      }}
    >
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </ConfigProvider>
  </React.StrictMode>,
);
