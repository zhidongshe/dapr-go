const {
  API_BASE_URL_STORAGE_KEY
} = require('./constants')

function getApiBaseUrl() {
  try {
    const configuredBaseUrl = wx.getStorageSync(API_BASE_URL_STORAGE_KEY)

    if (typeof configuredBaseUrl === 'string' && configuredBaseUrl.trim()) {
      return configuredBaseUrl.trim().replace(/\/$/, '')
    }
  } catch (error) {
    // Ignore storage access errors and validate configuration below.
  }

  throw new Error(
    `Missing API base URL configuration. Set wx storage key \"${API_BASE_URL_STORAGE_KEY}\" to the deployed gateway host before sending requests.`
  )
}

function request({ url, method = 'GET', data, header = {} }) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${getApiBaseUrl()}${url}`,
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
  getApiBaseUrl,
  request
}
