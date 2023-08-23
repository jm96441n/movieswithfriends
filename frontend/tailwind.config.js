/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "node_modules/daisyui/dist/**/*.js",
    "node_modules/react-daisyui/dist/**/*.js",
  ],
  theme: {
    container: {
      padding: "5rem",
    },
    extend: {},
  },
  daisyui: {
    themes: ["retro", "synthwave"],
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
};
