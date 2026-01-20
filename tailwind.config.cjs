module.exports = {
  content: ["./src/templates/**/*.templ", "./src/templates/**/*_templ.go"],
  theme: { extend: {} },
  plugins: [require("daisyui")],
  daisyui: {
    themes: true, // Enable all 32 themes
    darkTheme: "night", // Default dark theme
  },
};
