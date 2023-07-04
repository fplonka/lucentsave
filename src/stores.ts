import { writable } from 'svelte/store'

// When we initialize the store, we call the function to get the current signin status from the cookies
export const isSignedIn = writable(false)
