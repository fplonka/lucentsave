import { redirect } from '@sveltejs/kit'
import type { LayoutServerLoad } from './$types'

// Is there some better way to do this? 
export const load: LayoutServerLoad = ({ cookies }) => {
 	if (cookies.get('loggedIn') !== 'true') {
 		throw redirect(307, '/signin')
 	}
}