module.exports = {
    ci: {
        collect: {
            url: ['http://localhost:3000'],
            startServerCommand: 'npm run dev',
            startServerReadyPattern: 'ready',
            startServerReadyTimeout: 30000,
            settings: {
                chromeFlags: '--headless --no-sandbox --disable-setuid-sandbox --disable-dev-shm-usage --disable-gpu',
            },
        },
        assert: {
            assertions: {
                'categories:performance': ['warn', { minScore: 0.6 }],
                'categories:accessibility': ['error', { minScore: 0.9 }],
                'categories:best-practices': ['warn', { minScore: 0.8 }],
                'categories:seo': ['warn', { minScore: 0.8 }],
            },
        },
        upload: {
            target: 'filesystem',
            outputDir: './lighthouse-reports',
        },
    },
};
