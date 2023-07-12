import type { Post } from '../[listType]/+page.server'

export const filterPosts = (posts: Post[], path: string) => {
	switch (path) {
		case 'saved':
			return posts.filter(post => !post.isRead)
		case 'liked':
			return posts.filter(post => post.isLiked)
		case 'read':
			return posts.filter(post => post.isRead)
		default:
			return posts
	}
}
