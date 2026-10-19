/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'monospace'],
      },
      colors: {
        ink: {
          50: '#f7f8fa',
          100: '#eef0f4',
          200: '#dde1ea',
          300: '#c2c8d6',
          400: '#9aa3b8',
          500: '#6b748c',
          600: '#4a526b',
          700: '#353c52',
          800: '#222838',
          900: '#161a26',
          950: '#0c0f18',
        },
        brand: {
          50: '#eefcf6',
          100: '#d6f8e9',
          200: '#aef0d4',
          300: '#79e2ba',
          400: '#41cca0',
          500: '#1bb387',
          600: '#0f9170',
          700: '#0e745c',
          800: '#105c4b',
          900: '#114b3f',
        },
        accent: {
          50: '#eff8ff',
          100: '#dbeefe',
          200: '#bfe2fe',
          300: '#93cffd',
          400: '#60b1fa',
          500: '#3a8ff6',
          600: '#2370eb',
          700: '#1c5bd8',
          800: '#1e49af',
          900: '#1e408a',
        },
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
        },
        success: {
          500: '#1bb387',
          600: '#0f9170',
        },
        error: {
          50: '#fef2f2',
          100: '#fee2e2',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
        },
      },
      boxShadow: {
        card: '0 1px 2px rgba(15, 18, 25, 0.04), 0 4px 12px rgba(15, 18, 25, 0.06)',
        'card-hover': '0 6px 18px rgba(15, 18, 25, 0.10), 0 2px 6px rgba(15, 18, 25, 0.06)',
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0', transform: 'translateY(4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'slide-up': {
          '0%': { opacity: '0', transform: 'translateY(12px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-468px 0' },
          '100%': { backgroundPosition: '468px 0' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.25s ease-out',
        'slide-up': 'slide-up 0.3s ease-out',
        shimmer: 'shimmer 1.4s infinite linear',
      },
    },
  },
  plugins: [],
};
