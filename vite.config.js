import { defineConfig } from 'vite'
import tailwindcss from "@tailwindcss/vite"

export default defineConfig(() => {
	return {
		build: {
			emptyOutDir: false,
			outDir: "./assets/css",
			rollupOptions: {
				output: {
					assetFileNames: "tw.css",
				},
				input: "./base.css",
			}
		},
		server: false,
		plugins: [
			tailwindcss(),
		],
	}
})

