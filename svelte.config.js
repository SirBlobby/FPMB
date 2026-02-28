import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html', // Needed for SPA-like dynamic routing
			precompress: false,
			strict: true
		})
	}
};

export default config;
