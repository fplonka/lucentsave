import { redirect } from '@sveltejs/kit'
import type { PageServerLoad } from '../$types'
import { isSignedIn } from '../../stores'
import type { Post } from '../[listType]/+page.server'
import { PUBLIC_BACKEND_API_URL } from '$env/static/public'

// export const load: PageServerLoad = async ({ fetch }) => {
// 	const response = await fetch(PUBLIC_BACKEND_API_URL + '/api/getAllUserPosts', {
// 		credentials: 'include'
// 	})
//
// 	if (response.status === 401) {
// 		isSignedIn.set(false)
// 		throw redirect(307, '/signin')
// 	}
//
// 	if (response.ok) {
// 		isSignedIn.set(true)
// 		const allPosts: Post[] = await response.json()
// 		return { posts: allPosts }
// 	}
// 	var empty: Post[] = []
// 	return { posts: empty }
// }
