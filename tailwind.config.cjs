module.exports = {
  content: ["./src/templates/**/*.templ", "./src/templates/**/*_templ.go"],
  theme: { extend: {} },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        // Custom light theme (clone of DaisyUI light for future customization)
        light: {
          "color-scheme": "light",
          primary: "#570df8",
          secondary: "#f000b8",
          accent: "#37cdbe",
          neutral: "#3d4451",
          "base-100": "#ffffff",
          "base-200": "#f9fafb",
          "base-300": "#d1d5db",
          "base-content": "#1f2937",
          info: "#3abff8",
          success: "#36d399",
          warning: "#fbbd23",
          error: "#f87272",
        },
      },
      {
        // Custom dark theme (clone of DaisyUI dark for future customization)
        dark: {
          "color-scheme": "dark",
          primary: "#661ae6",
          secondary: "#d926aa",
          accent: "#1fb2a5",
          neutral: "#2a323c",
          "base-100": "#1d232a",
          "base-200": "#191e24",
          "base-300": "#15191e",
          "base-content": "#a6adba",
          info: "#3abff8",
          success: "#36d399",
          warning: "#fbbd23",
          error: "#f87272",
        },
      },
    ],
    darkTheme: "dark", // Default for prefers-color-scheme: dark
  },
};
