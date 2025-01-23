import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  		build: {
  		  // Generate assets to a specific directory
  		  outDir: '../static/css',
  		  // Only process files in src/
  		  rollupOptions: {
  		  		output: {
  		  		  assetFileNames: 'main-dev.css'
  		  		},
  		  		input: './base.css',
  		  }
  		},
  		// Disable Vite's dev server
  		server: false,
  plugins: [
    tailwindcss(),
  ],
})
