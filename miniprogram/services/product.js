const { request } = require('../utils/request')

function normalizeProduct(product) {
  if (!product || typeof product !== 'object') {
    return null
  }

  return {
    product_id: product.product_id,
    product_name: product.product_name,
    original_price: product.original_price,
    status: product.status
  }
}

function normalizeProductList(data) {
  const list = Array.isArray(data) ? data : Array.isArray(data?.list) ? data.list : []

  return list
    .map(normalizeProduct)
    .filter(Boolean)
}

function getProductList() {
  return request({
    url: '/v1/products',
    method: 'GET'
  }).then(normalizeProductList)
}

function getProductDetail(productId) {
  return request({
    url: `/v1/products/${productId}`,
    method: 'GET'
  }).then(normalizeProduct)
}

module.exports = {
  getProductList,
  getProductDetail,
  normalizeProduct,
  normalizeProductList
}
