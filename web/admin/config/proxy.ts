const ADMIN_API_BASE_URL = process.env.ADMIN_API_BASE_URL || 'http://127.0.0.1:8080';

const devProxy = {
  '/api': {
    target: ADMIN_API_BASE_URL,
    changeOrigin: true
  },
  '/auth': {
    target: ADMIN_API_BASE_URL,
    changeOrigin: true
  }
};

const proxy = {
  dev: devProxy,
  development: devProxy,
  test: devProxy,
  pre: devProxy,
  production: {}
};

export default proxy;
