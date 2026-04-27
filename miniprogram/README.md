# OMS WeChat Mini Program

## Open in WeChat DevTools
- Import the `miniprogram` directory as a mini-program project.
- Use `touristappid` for local UI development, then replace it with the real appid before release.

## Current scaffold status
- Task 1 sets up the app shell files (`app.js`, `app.json`, `app.wxss`, `sitemap.json`).
- Task 2 adds shared utility modules under `utils/` for requests, pricing, address formatting, and shared constants.
- No page files have been added yet.

## API base URL setup
- `utils/request.js` reads the API base URL from local storage key `apiBaseUrl` before each request.
- If nothing is configured yet, requests fall back to the placeholder `http://example.com/api`.
- Replace that placeholder, or write the deployed gateway base URL into storage, before testing real API calls.
