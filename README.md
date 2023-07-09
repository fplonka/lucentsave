# Lucentsave

Lucentsavel is a simple website for saving and organizing the things you want to read. It's a lot like [Pocket](https://getpocket.com) or [Instapaper](http://instapaper.com), except **faster**, **cleaner**, and **completely free**.

Try it now at [lucentsave.com](https://lucentsave.com).

<!-- ![image](https://github.com/fplonka/lucentsave/assets/92261790/d4a898cc-e4ad-4ed8-aba2-21e1977c4453) -->
![image](https://github.com/fplonka/lucentsave/assets/92261790/fa877afd-1d96-4804-81ed-525cd2089d94)
![image](https://github.com/fplonka/lucentsave/assets/92261790/c8a87e5d-565d-487d-9b9b-cc8766f048aa)


## Features Coming Soon
- Browser plugins to quickly save the current page
- Saving highlights in text
- Saving and viewing PDF files

## About this project

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
