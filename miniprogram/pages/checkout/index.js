const { getProductDetail } = require('../../services/product')
const { createOrder } = require('../../services/order')
const { DEFAULT_QUANTITY, CHECKOUT_UNAVAILABLE_MESSAGE } = require('../../utils/constants')
const {
  normalizeWechatAddress,
  formatAddressText,
  formatAddressReceiver
} = require('../../utils/address')

Page({
  data: {
    loading: true,
    error: '',
    product: null,
    quantity: DEFAULT_QUANTITY,
    amountText: '0.00',
    address: null,
    addressText: '',
    addressReceiver: '',
    submitting: false
  },

  onLoad(options) {
    this.initializeProduct(options)
  },

  async initializeProduct(options) {
    const incoming = options?.product ? JSON.parse(decodeURIComponent(options.product)) : null

    if (!incoming?.product_id) {
      this.setData({
        loading: false,
        error: CHECKOUT_UNAVAILABLE_MESSAGE
      })
      return
    }

    try {
      const product = await getProductDetail(incoming.product_id)
      this.setData({
        loading: false,
        product,
        amountText: product.display_price,
        error: ''
      })
    } catch (error) {
      this.setData({
        loading: false,
        error: error.message || CHECKOUT_UNAVAILABLE_MESSAGE
      })
    }
  },

  handleChooseAddress() {
    wx.chooseAddress({
      success: (address) => {
        const normalized = normalizeWechatAddress(address)
        this.setData({
          address: normalized,
          addressText: formatAddressText(normalized),
          addressReceiver: formatAddressReceiver(normalized)
        })
      },
      fail: (error) => {
        if (error.errMsg && !error.errMsg.includes('cancel')) {
          wx.showToast({
            title: '地址选择失败',
            icon: 'none'
          })
        }
      }
    })
  },

  async handleSubmitOrder() {
    const { product, quantity, address, submitting } = this.data

    if (submitting) {
      return
    }

    if (!address) {
      wx.showToast({
        title: '请选择收货地址',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })

    try {
      const order = await createOrder({
        productId: product.product_id,
        quantity,
        address
      })

      wx.redirectTo({
        url: `/pages/order-success/index?orderNo=${encodeURIComponent(order.order_no || '')}&productName=${encodeURIComponent(product.product_name)}&amount=${encodeURIComponent(product.display_price)}`
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '订单创建失败',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
    }
  },

  handleGoHome() {
    wx.reLaunch({
      url: '/pages/home/index'
    })
  }
})
