/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			typography: {
			  quoteless: {
				css: {
				  'blockquote p:first-of-type::before': { content: 'none' },
				  'blockquote p:first-of-type::after': { content: 'none' },
				},
			  },
			},
			fontFamily: {
				'serif': ['Bookerly'],
				'sans': ['Inter'],
			}
		  },
	},
	plugins: [
		require('@tailwindcss/typography'),
	],
};
