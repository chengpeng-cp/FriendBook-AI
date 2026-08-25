/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      /* 品牌色板：深墨底 + 温度光谱（FriendBook 核心视觉语言） */
      colors: {
        ink: {
          950: '#0B0E14', // 页面底色（深墨）
          900: '#10141D', // 区块底色
          800: '#161B26', // 卡片底色
          700: '#1F2532', // 边框/悬浮
          600: '#2A3242', // 强边框
        },
        fog: {
          900: '#E8ECF3', // 主文字
          700: '#B4BDCB', // 次文字
          500: '#8A94A6', // 弱文字
          300: '#5A6478', // 占位符
        },
        /* 温度光谱：夯→热→温→凉→拉 */
        temp: {
          hot: '#FF5C39',
          warm: '#FFA940',
          mild: '#F6C453',
          cool: '#63B3ED',
          cold: '#8A94A3',
        },
        brand: {
          400: '#FFB25C',
          500: '#FF8A4C',
          600: '#F0653A',
        },
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'Segoe UI',
          'Roboto',
          'sans-serif',
        ],
      },
      boxShadow: {
        card: '0 8px 30px rgba(0,0,0,.35)',
        glow: '0 0 24px rgba(255,138,76,.25)',
      },
      animation: {
        'pulse-soft': 'pulseSoft 2.4s ease-in-out infinite',
        shimmer: 'shimmer 1.8s linear infinite',
      },
      keyframes: {
        pulseSoft: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '.45' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-400px 0' },
          '100%': { backgroundPosition: '400px 0' },
        },
      },
    },
  },
  plugins: [],
}
