import type { Config } from "tailwindcss";

const config: Config = {
    content: [
        "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                // Primary brand colors
                primary: {
                    50: "hsl(var(--primary-50))",
                    100: "hsl(var(--primary-100))",
                    200: "hsl(var(--primary-200))",
                    300: "hsl(var(--primary-300))",
                    400: "hsl(var(--primary-400))",
                    500: "hsl(var(--primary-500))",
                    600: "hsl(var(--primary-600))",
                    700: "hsl(var(--primary-700))",
                    800: "hsl(var(--primary-800))",
                    900: "hsl(var(--primary-900))",
                    950: "hsl(var(--primary-950))",
                },
                // Secondary colors
                secondary: {
                    50: "hsl(var(--secondary-50))",
                    100: "hsl(var(--secondary-100))",
                    200: "hsl(var(--secondary-200))",
                    300: "hsl(var(--secondary-300))",
                    400: "hsl(var(--secondary-400))",
                    500: "hsl(var(--secondary-500))",
                    600: "hsl(var(--secondary-600))",
                    700: "hsl(var(--secondary-700))",
                    800: "hsl(var(--secondary-800))",
                    900: "hsl(var(--secondary-900))",
                    950: "hsl(var(--secondary-950))",
                },
                // Neutral colors
                neutral: {
                    50: "hsl(var(--neutral-50))",
                    100: "hsl(var(--neutral-100))",
                    200: "hsl(var(--neutral-200))",
                    300: "hsl(var(--neutral-300))",
                    400: "hsl(var(--neutral-400))",
                    500: "hsl(var(--neutral-500))",
                    600: "hsl(var(--neutral-600))",
                    700: "hsl(var(--neutral-700))",
                    800: "hsl(var(--neutral-800))",
                    900: "hsl(var(--neutral-900))",
                    950: "hsl(var(--neutral-950))",
                },
                // Success colors
                success: {
                    50: "hsl(var(--success-50))",
                    100: "hsl(var(--success-100))",
                    200: "hsl(var(--success-200))",
                    300: "hsl(var(--success-300))",
                    400: "hsl(var(--success-400))",
                    500: "hsl(var(--success-500))",
                    600: "hsl(var(--success-600))",
                    700: "hsl(var(--success-700))",
                    800: "hsl(var(--success-800))",
                    900: "hsl(var(--success-900))",
                    950: "hsl(var(--success-950))",
                },
                // Warning colors
                warning: {
                    50: "hsl(var(--warning-50))",
                    100: "hsl(var(--warning-100))",
                    200: "hsl(var(--warning-200))",
                    300: "hsl(var(--warning-300))",
                    400: "hsl(var(--warning-400))",
                    500: "hsl(var(--warning-500))",
                    600: "hsl(var(--warning-600))",
                    700: "hsl(var(--warning-700))",
                    800: "hsl(var(--warning-800))",
                    900: "hsl(var(--warning-900))",
                    950: "hsl(var(--warning-950))",
                },
                // Error colors
                error: {
                    50: "hsl(var(--error-50))",
                    100: "hsl(var(--error-100))",
                    200: "hsl(var(--error-200))",
                    300: "hsl(var(--error-300))",
                    400: "hsl(var(--error-400))",
                    500: "hsl(var(--error-500))",
                    600: "hsl(var(--error-600))",
                    700: "hsl(var(--error-700))",
                    800: "hsl(var(--error-800))",
                    900: "hsl(var(--error-900))",
                    950: "hsl(var(--error-950))",
                },
                // Background colors
                background: "hsl(var(--background))",
                foreground: "hsl(var(--foreground))",
                // Card colors
                card: {
                    DEFAULT: "hsl(var(--card))",
                    foreground: "hsl(var(--card-foreground))",
                },
                // Popover colors
                popover: {
                    DEFAULT: "hsl(var(--popover))",
                    foreground: "hsl(var(--popover-foreground))",
                },
                // Muted colors
                muted: {
                    DEFAULT: "hsl(var(--muted))",
                    foreground: "hsl(var(--muted-foreground))",
                },
                // Accent colors
                accent: {
                    DEFAULT: "hsl(var(--accent))",
                    foreground: "hsl(var(--accent-foreground))",
                },
                // Destructive colors
                destructive: {
                    DEFAULT: "hsl(var(--destructive))",
                    foreground: "hsl(var(--destructive-foreground))",
                },
                // Border colors
                border: "hsl(var(--border))",
                // Input colors
                input: "hsl(var(--input))",
                // Ring colors
                ring: "hsl(var(--ring))",
            },
            borderRadius: {
                lg: "var(--radius)",
                md: "calc(var(--radius) - 2px)",
                sm: "calc(var(--radius) - 4px)",
            },
            fontFamily: {
                sans: ["var(--font-geist-sans)", "system-ui", "sans-serif"],
                mono: ["var(--font-geist-mono)", "monospace"],
            },
            keyframes: {
                "accordion-down": {
                    from: { height: "0" },
                    to: { height: "var(--radix-accordion-content-height)" },
                },
                "accordion-up": {
                    from: { height: "var(--radix-accordion-content-height)" },
                    to: { height: "0" },
                },
                "fade-in": {
                    "0%": { opacity: "0" },
                    "100%": { opacity: "1" },
                },
                "fade-out": {
                    "0%": { opacity: "1" },
                    "100%": { opacity: "0" },
                },
                "slide-in-from-top": {
                    "0%": { transform: "translateY(-100%)" },
                    "100%": { transform: "translateY(0)" },
                },
                "slide-in-from-bottom": {
                    "0%": { transform: "translateY(100%)" },
                    "100%": { transform: "translateY(0)" },
                },
                "slide-in-from-left": {
                    "0%": { transform: "translateX(-100%)" },
                    "100%": { transform: "translateX(0)" },
                },
                "slide-in-from-right": {
                    "0%": { transform: "translateX(100%)" },
                    "100%": { transform: "translateX(0)" },
                },
            },
            animation: {
                "accordion-down": "accordion-down 0.2s ease-out",
                "accordion-up": "accordion-up 0.2s ease-out",
                "fade-in": "fade-in 0.2s ease-out",
                "fade-out": "fade-out 0.2s ease-out",
                "slide-in-from-top": "slide-in-from-top 0.2s ease-out",
                "slide-in-from-bottom": "slide-in-from-bottom 0.2s ease-out",
                "slide-in-from-left": "slide-in-from-left 0.2s ease-out",
                "slide-in-from-right": "slide-in-from-right 0.2s ease-out",
            },
        },
    },
    plugins: [],
};

export default config;
