/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			typography: theme => ({
				quoteless: {
					css: {
						'blockquote p:first-of-type::before': { content: 'none' },
						'blockquote p:first-of-type::after': { content: 'none' }
					}
				},
				base: {
					css: {
						lineHeight: 1.5,
						maxWidth: 'none'
					}
				},
				lg: {
					css: {
						lineHeight: 1.5,
						maxWidth: 'none'
					}
				}
			}),
			fontFamily: {
				sans: ['Roboto', 'Inter', 'sans-serif']
			}
		}
	},
	plugins: [require('@tailwindcss/typography')]
}
