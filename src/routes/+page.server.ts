import { redirect } from '@sveltejs/kit'

// export const ssr = false

export const load = ({ cookies }) => {
	if (cookies.get('loggedIn') == 'true') {
		throw redirect(307, '/saved')
	} else {
		throw redirect(307, '/signin')
	}
}
