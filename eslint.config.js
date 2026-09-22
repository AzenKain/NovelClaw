import frontendConfig from './frontend/eslint.config.js';

export default [
  {
    ignores: [
      '**/node_modules/**',
      '**/dist/**',
      '**/bindings/**',
      'build/**',
      'cmd/**',
      'pkg/**',
      'internal/**',
      'test/**',
      'testdata/**',
      'data/**',
      'docs/**',
      'research/**',
      '.agents/**',
    ],
  },
  ...frontendConfig,
];
