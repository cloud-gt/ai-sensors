import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: '../dashboard/dist',
			assets: '../dashboard/dist',
			fallback: 'index.html',
			precompress: false,
			strict: false
		}),
		paths: {
			base: '/dashboard'
		}
	}
};

export default config;
