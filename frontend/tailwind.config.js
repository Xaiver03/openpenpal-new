/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ["class"],
  content: [
    './pages/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './app/**/*.{ts,tsx}',
    './src/**/*.{ts,tsx}',
  ],
  // 生产环境下移除未使用的样式 - Tailwind CSS v3.0+ 使用 content 配置
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // OpenPenPal 主题色彩 - 纸黄色和白色系
        letter: {
          paper: "#fefcf7",      // 温暖的纸黄色背景
          cream: "#fdf6e3",      // 奶油色
          amber: "#f59e0b",      // 琥珀色主色调
          'amber-light': "#fef3c7", // 浅琥珀色
          'amber-dark': "#d97706",  // 深琥珀色
          ink: "#7c2d12",        // 深棕色文字
          'ink-light': "#a3a3a3", // 浅灰色辅助文字
          white: "#ffffff",      // 纯白色
          border: "#f3e8ff"      // 淡紫色边框
        }
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: 0 },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: 0 },
        },
        "envelope-fold": {
          "0%": { transform: "rotateX(0deg)" },
          "100%": { transform: "rotateX(-90deg)" }
        },
        "dove-fly": {
          "0%": { transform: "translateY(0px) scale(1)" },
          "100%": { transform: "translateY(-100px) scale(0.8)" }
        }
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "envelope-fold": "envelope-fold 0.6s ease-in-out",
        "dove-fly": "dove-fly 1s ease-out"
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui"],
        serif: ["Noto Serif SC", "ui-serif", "Georgia"],
        mono: ["ui-monospace", "SFMono-Regular"]
      }
    },
  },
  plugins: [require("tailwindcss-animate")],
}