import { writable} from 'svelte/store';

// When we initialize the store, we call the function to get the current login status from the cookies
export const isLoggedIn = writable(false);