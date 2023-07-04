// src/routes/posts/+page.ts
import { redirect } from '@sveltejs/kit'
import type { PageLoad } from './$types'
import { isSignedIn } from '../../stores'
import { filterPosts } from './util'

export interface Post {
	id: number
	url: string
	title: string
	body: string
	isRead: boolean
	isLiked: boolean
}

export const load: PageLoad = async ({ params, fetch }) => {
	const response = await fetch('http://localhost:8080/api/getAllUserPosts', {
		credentials: 'include'
	})

	if (response.status === 401) {
		isSignedIn.set(false)
		throw redirect(307, '/signin')
	}

	if (response.ok) {
		isSignedIn.set(true)

		const allPosts: Post[] = await response.json()

		const listType = params.listType // The route parameter (saved, liked, or read)
		if (!['saved', 'liked', 'read'].includes(listType)) {
			throw redirect(307, '/')
		}

		return { posts: filterPosts(allPosts, listType) }
	}
	var empty: Post[] = []
	return { posts: empty }
}
