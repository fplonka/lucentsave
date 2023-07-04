import { redirect } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

export const prerender = true

export const load = (({ cookies }) => {
	if (cookies.get('loggedIn') == 'true') {
		throw redirect(307, '/saved')
	}
}) satisfies LayoutServerLoad
