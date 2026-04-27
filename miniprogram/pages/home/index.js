const { getProductList } = require('../../services/product')

Page({
  data: {
    loading: true,
    error: '',
    products: []
  },

  onLoad() {
    this.loadProducts()
  },

  onPullDownRefresh() {
    this.loadProducts().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadProducts() {
    this.setData({
      loading: true,
      error: ''
    })

    try {
      const products = await getProductList()
      this.setData({ products })
    } catch (error) {
      this.setData({
        error: error.message || '商品加载失败，请稍后重试'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  handleBuyNow(event) {
    const product = event.currentTarget.dataset.product

    wx.navigateTo({
      url: `/pages/checkout/index?product=${encodeURIComponent(JSON.stringify(product))}`
    })
  }
})
