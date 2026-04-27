const { API_BASE_URL } = require('./constants')

function request({ url, method = 'GET', data, header = {} }) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...header
      },
      success(res) {
        const { statusCode, data: body } = res

        if (statusCode >= 200 && statusCode < 300) {
          resolve(body)
          return
        }

        reject(new Error(body?.message || `Request failed with status ${statusCode}`))
      },
      fail(error) {
        reject(new Error(error.errMsg || 'Network request failed'))
      }
    })
  })
}

module.exports = {
  request
}
