/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "views/templates/**/*.html",
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
