Page({
  data: {
    orderNo: '',
    productName: '',
    amount: '0.00'
  },

  onLoad(options) {
    this.setData({
      orderNo: decodeURIComponent(options.orderNo || ''),
      productName: decodeURIComponent(options.productName || ''),
      amount: decodeURIComponent(options.amount || '0.00')
    })
  },

  handleBackHome() {
    wx.reLaunch({
      url: '/pages/home/index'
    })
  }
})
