import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue({
      script: {
        defineModel: true,
        propsDestructure: true
      }
    }),
    // 自动导入 API
    AutoImport({
      // 自动导入 Vue 相关函数，如：ref, reactive, toRef 等
      imports: ['vue', 'vue-router', 'pinia'],
      // 自动导入 Element Plus 相关函数
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts',
    }),
    // 自动导入组件
    Components({
      // 自动导入 Element Plus 组件
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@components': resolve(__dirname, 'src/components'),
      '@composables': resolve(__dirname, 'src/composables'),
      '@types': resolve(__dirname, 'src/types'),
      '@utils': resolve(__dirname, 'src/utils'),
      '@assets': resolve(__dirname, 'src/assets')
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        // target: 'http://localhost:8080',
        target: 'http://192.168.252.1:8888',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    // 启用源码映射，便于调试
    sourcemap: true,
    // 代码分割配置
    rollupOptions: {
      output: {
        // 自定义代码分割策略
        manualChunks: (id) => {
          // Vue相关库打包到一起
          if (id.includes('node_modules/vue') || 
              id.includes('node_modules/pinia') || 
              id.includes('node_modules/vue-router')) {
            return 'vue-vendor';
          }
          
          // Element Plus打包到一起
          if (id.includes('node_modules/element-plus') || 
              id.includes('node_modules/@element-plus')) {
            return 'element-plus';
          }
          
          // API相关代码打包到一起
          if (id.includes('/src/api/')) {
            return 'api';
          }
          
          // Composables相关代码打包到一起
          if (id.includes('/src/composables/')) {
            return 'composables';
          }
          
          // 第三方库打包到一起
          if (id.includes('node_modules/') && 
              !id.includes('node_modules/vue') && 
              !id.includes('node_modules/pinia') && 
              !id.includes('node_modules/vue-router') && 
              !id.includes('node_modules/element-plus') && 
              !id.includes('node_modules/@element-plus')) {
            return 'vendor';
          }
        },
        // 自定义chunk文件名格式
        chunkFileNames: 'assets/js/[name]-[hash].js',
        // 入口文件名格式
        entryFileNames: 'assets/js/[name]-[hash].js',
        // 静态资源文件名格式
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]'
      }
    },
    // 设置chunk大小警告阈值
    chunkSizeWarningLimit: 1000
  }
})