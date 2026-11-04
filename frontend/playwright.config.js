import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: './tests',
    workers: 1,
    timeout: 30_000,
    use: {
        baseURL: 'http://localhost:34115',
        browserName: 'chromium',
        channel: 'chromium',
        launchOptions: { args: ['--autoplay-policy=no-user-gesture-required'] },
    },
    webServer: {
        command: 'npm run dev -- --host localhost --port 34115 --strictPort',
        url: 'http://localhost:34115',
        reuseExistingServer: !process.env.CI,
        timeout: 30_000,
    },
});
