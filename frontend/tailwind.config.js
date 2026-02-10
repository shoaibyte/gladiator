/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bronze: {
          50: '#fdf8f6',
          100: '#faefe6',
          200: '#f2d8c1',
          300: '#e6ba8f',
          400: '#cd7f32',
          500: '#b87333',
          600: '#8b5a2b',
          700: '#6b4421',
          800: '#4d3018',
          900: '#332010',
        },
      },
    },
  },
  plugins: [],
}
