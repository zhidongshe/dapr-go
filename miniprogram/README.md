# OMS WeChat Mini Program

## Open in WeChat DevTools
- Import the `miniprogram` directory as a mini-program project.
- Use `touristappid` for local UI development, then replace it with the real appid before release.

## Backend base URL
- Update `miniprogram/utils/constants.js` to point at the local `api-gateway` address.
- For phone/device debugging, use a LAN-reachable host instead of `localhost`.
