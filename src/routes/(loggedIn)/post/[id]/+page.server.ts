import { redirect } from '@sveltejs/kit'
import type { Post } from '$lib/types'
import { PUBLIC_BACKEND_API_URL } from '$env/static/public'
import type { PageServerLoad } from './$types'

// ???
export const ssr = false

export const load: PageServerLoad = async ({ params, fetch }) => {
	const response = await fetch(PUBLIC_BACKEND_API_URL + `getPost?id=${params.id}`, {
		credentials: 'include'
	})
	if (response.ok) {
		const post: Post = await response.json()
		return { post, id: params.id }
	}

	throw redirect(307, '/saved')
}
