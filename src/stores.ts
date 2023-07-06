import { PUBLIC_BACKEND_API_URL } from '$env/static/public'
import { writable, type Writable } from 'svelte/store'
import type { Post } from './routes/[listType]/+page.server'

// When we initialize the store, we call the function to get the current signin status from the cookies
export const isSignedIn = writable(false)

export const posts: Writable<Post[]> = writable([])
export const postsLoaded = writable(false)
