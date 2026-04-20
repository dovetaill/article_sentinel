const vendorPrefix = '/node_modules/';

export function chunkForModule(id: string): string | undefined {
  if (!id.includes(vendorPrefix)) {
    return undefined;
  }
  if (id.includes('/react-router') || id.includes('/@remix-run/router/')) {
    return 'router-vendor';
  }
  if (id.includes('/react-dom/') || id.includes('/react/')) {
    return 'react-vendor';
  }
  const proChunk = proChunkName(id);
  if (proChunk) {
    return proChunk;
  }
  if (id.includes('/@ant-design/icons/')) {
    return 'icons-vendor';
  }
  if (id.includes('/rc-') || id.includes('/@rc-component/')) {
    return 'rc-vendor';
  }
  const antdChunk = antdChunkName(id);
  if (antdChunk) {
    return antdChunk;
  }
  return 'vendor';
}

function antdChunkName(id: string): string | undefined {
  const match = id.match(/\/antd\/(?:es|lib)\/([^/]+)/);
  if (!match || !match[1]) {
    return undefined;
  }
  return `antd-${match[1]}`;
}

function proChunkName(id: string): string | undefined {
  const match = id.match(/\/@ant-design\/(pro-[^/]+)\//);
  if (!match || !match[1]) {
    return undefined;
  }
  return match[1];
}
