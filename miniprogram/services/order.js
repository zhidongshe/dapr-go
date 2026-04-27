const { request } = require('../utils/request')

function createOrder({ productId, quantity, address }) {
  return request({
    url: '/v1/orders',
    method: 'POST',
    data: {
      items: [
        {
          product_id: productId,
          quantity
        }
      ],
      address
    }
  })
}

module.exports = {
  createOrder
}
