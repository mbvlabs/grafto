export default {
	content: [
		"../views/**/*.templ"
	],
	blocklist: [],
	darkMode: 'class',
	plugins: [
		require('@tailwindcss/forms'),
		require("@tailwindcss/typography")
	],
}
