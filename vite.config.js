import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ command, mode, isSsrBuild, isPreview }) => {
  if (mode === 'dev') {
    return {
  		build: {
		  emptyOutDir: false,
  		  outDir: './static/css',
  		  rollupOptions: {
  		  		output: {
  		  		  assetFileNames: 'main-dev.css'
  		  		},
  		  		input: './resources/css/base.css',
  		  }
  		},
  		server: false,
  		plugins: [
  		  tailwindcss(),
  		],
    }
  } 
  if (mode === 'prod') {
    return {
  		build: {
  		  outDir: './static/css',
  		  rollupOptions: {
  		  		output: {
  		  		  assetFileNames: 'css/main-prod-[hash].css'
  		  		},
  		  		input: './resources/css/base.css',
  		  }
  		},
  		// Disable Vite's dev server
  		server: false,
  		plugins: [
  		  tailwindcss(),
  		],
    }
  } 
})

