import { redirect } from '@sveltejs/kit'
import type { PageServerLoad } from './$types'
import { isSignedIn } from '../../stores'
import { filterPosts } from './util'
import { PUBLIC_BACKEND_API_URL } from '$env/static/public'

export interface Post {
	id: number
	url: string
	title: string
	body: string
	isRead: boolean
	isLiked: boolean
}

// export const load: PageServerLoad = async ({ params, fetch }) => {
// 	const response = await fetch(PUBLIC_BACKEND_API_URL + '/api/getAllUserPosts', {
// 		credentials: 'include'
// 	})

// 	if (response.status === 401) {
// 		isSignedIn.set(false)
// 		throw redirect(307, '/signin')
// 	}

// 	if (response.ok) {
// 		isSignedIn.set(true)

// 		const allPosts: Post[] = await response.json()

// 		const listType = params.listType // The route parameter (saved, liked, or read)
// 		if (!['saved', 'liked', 'read'].includes(listType)) {
// 			throw redirect(307, '/')
// 		}

// 		return { posts: filterPosts(allPosts, listType) }
// 	}
// 	var empty: Post[] = []
// 	return { posts: empty }
// }

export const load: PageServerLoad = ({ cookies }) => {
	if (cookies.get('loggedIn') !== 'true') {
		throw redirect(307, '/signin')
	}
}
