import process from 'node:process';
import { URL, fileURLToPath } from 'node:url';
import { defineConfig, loadEnv } from 'vite';
import { setupVitePlugins } from './build/plugins';
import { assetFileNames, chunkGroups } from './build/config/chunks';
import { createViteProxy, getBuildTime } from './build/config';

export default defineConfig(configEnv => {
  const viteEnv = loadEnv(configEnv.mode, process.cwd()) as unknown as Env.ImportMeta;

  const buildTime = getBuildTime();

  const enableProxy = configEnv.command === 'serve' && !configEnv.isPreview;

  return {
    base: viteEnv.VITE_BASE_URL,
    resolve: {
      alias: {
        '~': fileURLToPath(new URL('./', import.meta.url)),
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
          additionalData: (content, loaderPath) => {
            // 避免循环依赖：不要在以下文件中注入 global.scss
            // 1. global.scss 本身
            // 2. global.scss 导入的文件（避免反向注入造成循环）
            const excludedPaths = [
              '/styles/scss/global.scss',
              '/styles/scss/sxdevops-theme.scss',
              '/styles/scss/layout-theme.scss',
              '/styles/scss/interaction-states.scss',
              '/styles/scss/sidebar.scss',
              '/styles/scss/design-system.scss',
              '/styles/scss/element-plus.scss',
              '/styles/scss/compact-theme.scss',
              '/styles/scss/terminal-workbench.scss'
            ];

            if (excludedPaths.some(path => loaderPath.includes(path))) {
              return content;
            }
            return `@use "@/styles/scss/global.scss" as *; ${content}`;
          }
        }
      }
    },
    plugins: setupVitePlugins(viteEnv, buildTime),
    define: {
      BUILD_TIME: JSON.stringify(buildTime)
    },
    server: {
      host: '0.0.0.0',
      port: 9527,
      open: true,
      proxy: createViteProxy(viteEnv, enableProxy)
    },
    preview: {
      port: 9725
    },
    build: {
      reportCompressedSize: false,
      sourcemap: viteEnv.VITE_SOURCE_MAP === 'Y',
      commonjsOptions: {
        ignoreTryCatch: false
      },
      // 生产环境使用 terser 压缩并移除 console/debugger，其余环境用默认压缩保持构建速度
      ...(configEnv.mode === 'prod'
        ? {
            minify: 'terser' as const,
            terserOptions: {
              compress: {
                drop_console: true,
                drop_debugger: true,
                dead_code: true,
                unused: true
              }
            }
          }
        : {}),
      rolldownOptions: {
        output: {
          // ============================================================================
          // 浏览器缓存策略：通过 contenthash 实现长期缓存
          // 1. [hash] 基于文件内容生成，内容不变 hash 不变
          // 2. 配合服务器设置 Cache-Control: max-age=31536000, immutable
          // 3. 大库独立分包（见 build/config/chunks.ts），业务代码变动不会导致 vendor 缓存失效
          // ============================================================================
          entryFileNames: 'assets/entry/[name]-[hash].js',
          chunkFileNames: 'assets/chunks/[name]-[hash].js',
          assetFileNames,
          advancedChunks: {
            groups: chunkGroups
          }
        }
      }
    }
  };
});
