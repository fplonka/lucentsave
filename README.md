# Lucentsave

Lucentsave is a simple website for saving and organizing the things you want to read. It's a lot like [Pocket](https://getpocket.com) or [Instapaper](http://instapaper.com), except **faster**, **cleaner**, and **completely free**.

Features:
- Full body search in saved pages
- Highlighting page text
- Browser extensions for [Firefox](https://addons.mozilla.org/addon/lucentsave/) and [Chrome](https://chrome.google.com/webstore/detail/ecjdaebdopdhiicoeoolkgdichihdlcg)

Try now at [lucentsave.com](https://lucentsave.com).

![output_test](https://github.com/fplonka/lucentsave/assets/92261790/f58c7a2e-2ad7-4ad2-85de-b80b1913807e)

## Features Coming Soon
- Saving and viewing PDF files

## Building

This project was built with [SvelteKit](https://kit.svelte.dev), [Go](https://go.dev) and [Postgres](https://www.postgresql.org).

To clone the repo, run:
```bash
git clone https://github.com/fplonka/lucentsave.git
```

To run the backend:
```bash
cd server/
go run .
```

To run the frontend:
```bash
npm install
npm run dev
```
Or to run a production server:
```bash
npm run build && npm run start
```

Before the project can be run, you will want to create a `.env` file in the projet root directory where you specify where the application is running. If running with `npm run dev`, you would create a `.env` file with the following content:
```conf
PUBLIC_BACKEND_API_URL=http://localhost:8080/api/
PUBLIC_APPLICATION_URL=http://localhost:5173
```
