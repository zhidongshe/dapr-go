const { request } = require('../utils/request')
const { centsToYuan } = require('../utils/price')

function mapProduct(product) {
  return {
    product_id: product.product_id,
    product_name: product.product_name,
    original_price: product.original_price,
    status: product.status,
    display_price: centsToYuan(product.original_price)
  }
}

function getProductList() {
  return request({
    url: '/v1/products',
    method: 'GET'
  }).then((data) => {
    const list = Array.isArray(data) ? data : Array.isArray(data?.list) ? data.list : []

    return list
      .filter((item) => Number(item.status) === 1)
      .map(mapProduct)
  })
}

function getProductDetail(productId) {
  return request({
    url: `/v1/products/${productId}`,
    method: 'GET'
  }).then(mapProduct)
}

module.exports = {
  getProductList,
  getProductDetail
}
