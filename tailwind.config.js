/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "views/internal/pages/**/*.templ",
    "views/internal/layouts/**/*.templ",
    "views/internal/components/**/*.templ",
    "node_modules/preline/dist/*.js",
  ],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["dark"],
  },
  plugins: [
    require('preline/plugin'),
  ],
}
