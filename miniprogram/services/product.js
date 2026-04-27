const { request } = require('../utils/request')

function getProductList() {
  return request({
    url: '/v1/products',
    method: 'GET'
  })
}

function getProductDetail(productId) {
  return request({
    url: `/v1/products/${productId}`,
    method: 'GET'
  })
}

module.exports = {
  getProductList,
  getProductDetail
}
