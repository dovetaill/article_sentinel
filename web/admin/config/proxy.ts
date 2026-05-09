const ADMIN_API_BASE_URL = process.env.ADMIN_API_BASE_URL || 'http://127.0.0.1:8080';

const proxy = {
  dev: {
    '/api': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    },
    '/auth': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    }
  },
  development: {
    '/api': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    },
    '/auth': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    }
  },
  test: {
    '/api': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    },
    '/auth': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    }
  },
  pre: {
    '/api': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    },
    '/auth': {
      target: ADMIN_API_BASE_URL,
      changeOrigin: true
    }
  },
  production: {}
};

export default proxy;
