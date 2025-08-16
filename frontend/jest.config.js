// 前端测试配置
module.exports = {
  projects: [
    {
      displayName: 'frontend',
      testEnvironment: 'jsdom',
      testMatch: ['<rootDir>/frontend/**/*.test.{js,jsx,ts,tsx}'],
      transform: {
        '^.+\\.(js|jsx|ts|tsx)$': ['babel-jest', { configFile: './frontend/.babelrc' }],
      },
      moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/frontend/src/$1',
        '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
      },
      setupFilesAfterEnv: ['<rootDir>/frontend/jest.setup.js'],
      collectCoverageFrom: [
        'frontend/src/**/*.{js,jsx,ts,tsx}',
        '!frontend/src/**/*.d.ts',
        '!frontend/src/**/*.stories.{js,jsx,ts,tsx}',
        '!frontend/src/**/__tests__/**',
      ],
      coverageThreshold: {
        global: {
          branches: 70,
          functions: 70,
          lines: 70,
          statements: 70,
        },
      },
    },
  ],
  coverageReporters: ['text', 'lcov', 'html'],
  coverageDirectory: '<rootDir>/coverage',
};