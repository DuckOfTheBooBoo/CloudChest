/** @type {import('tailwindcss').Config} */
export default {
  important: true,
  prefix: "tw-",
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,vue}"
  ],
  theme: {
    extend: {
      backgroundImage: {
        'fade-out': 'linear-gradient(to right, rgba(255, 255, 255, 0), rgba(255, 255, 255, 1))',
      },
    },
  },
  plugins: [],
}

