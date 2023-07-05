// $lib/postActions.ts

import type { Post } from '../routes/[listType]/+page.server'

export async function updatePostStatus (
	postId: number,
	read: boolean,
	liked: boolean
): Promise<Post> {
	const response = await fetch('http://localhost:8080/api/updatePostStatus', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ id: postId, read, liked }),
		credentials: 'include'
	})

	if (!response.ok) {
		throw new Error('Failed to update post status')
	}

	return response.json()
}

export function markAsRead (post: Post) {
	if (post.isRead) {
		return updatePostStatus(post.id, false, false) // Can't have unread liked post so we also set liked to false
	} else {
		return updatePostStatus(post.id, true, post.isLiked)
	}
}

export function like (post: Post): Promise<Post> {
	if (post.isRead) {
		return updatePostStatus(post.id, true, !post.isLiked)
	} else {
		if (post.isLiked) {
			return updatePostStatus(post.id, false, false) // Just unlike this post
		} else {
			return updatePostStatus(post.id, true, true) // Like this post and mark as read
		}
	}
}
