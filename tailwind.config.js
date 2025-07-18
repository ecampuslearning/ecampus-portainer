const plugin = require('tailwindcss/plugin');
const defaultTheme = require('tailwindcss/defaultTheme');
const colors = require('./app/assets/css/colors.json');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{html,tsx}'],
  corePlugins: {
    preflight: false,
  },
  theme: {
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      inherit: 'inherit',
      ...colors,

      'legacy-grey-3': 'var(--grey-3)',
      'legacy-blue-2': 'var(--blue-2)',
      'legacy-blue-9': 'var(--blue-9)',
    },
    extend: {
      fontFamily: {
        sans: ['Inter', ...defaultTheme.fontFamily.sans],
      },
      animation: {
        'spin-slow': 'spin 2s linear infinite',
      },
    },
  },

  plugins: [
    plugin(({ addVariant }) => {
      addVariant('be', '&:is([data-edition="BE"] *)');
      addVariant('th-highcontrast', '&:is([theme="highcontrast"] *)');
      addVariant('th-dark', '&:is([theme="dark"] *)');
    }),
    plugin(function ({ addVariant }) {
      addVariant('progress-filled', ['&::-webkit-progress-value', '&::-moz-progress-bar']);
    }),
    require('tailwindcss-animate'),
  ],
};
