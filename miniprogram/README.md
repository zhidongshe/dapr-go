# OMS WeChat Mini Program

## Open in WeChat DevTools
- Import the `miniprogram` directory as a mini-program project.
- Use `touristappid` for local UI development, then replace it with the real appid before release.

## Backend base URL setup
- Update `miniprogram/utils/constants.js` to point at the local `api-gateway` address.
- For phone/device debugging, use a LAN-reachable host instead of `localhost`.

## Manual verification
1. Start the backend gateway and related services.
2. Open the mini-program in WeChat DevTools.
3. Confirm the home page loads on-sale products only.
4. Tap `立即购买` and confirm checkout refreshes product detail.
5. Select an address with `wx.chooseAddress`.
6. Submit the order and confirm the success page shows order number, product name, and amount.
7. Temporarily make the product off-sale and confirm checkout blocks order creation.
