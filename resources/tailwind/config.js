/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../views/**/*.templ",
  ],
  darkMode: 'class',
  daisyui: {
    themes: ["dracula"],
  },
  plugins: [
    require('@tailwindcss/forms'),
    require("@tailwindcss/typography"),
    require('daisyui')
  ],
}
