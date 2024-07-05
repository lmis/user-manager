/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [ "./cmd/app/router/render/**/*.{templ,js}"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
  daisyui: {
    themes: ["retro"],
  },
}

