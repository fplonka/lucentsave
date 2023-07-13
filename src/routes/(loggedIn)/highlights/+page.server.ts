import { redirect } from '@sveltejs/kit'
import { PUBLIC_BACKEND_API_URL } from '$env/static/public'
import type { PageServerLoad } from './$types'

// ???
export const ssr = false

export const load: PageServerLoad = async ({ fetch }) => {
	const response = await fetch(PUBLIC_BACKEND_API_URL + 'getAllUserHighlights', {
		credentials: 'include'
	})
	if (response.ok) {
		const highlights: { id: string; postId: number; text: string; title: string }[] =
			await response.json()
		return { highlights }
	}

	throw redirect(307, '/saved')
}
