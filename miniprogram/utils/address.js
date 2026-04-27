function normalizeWechatAddress(address) {
  if (!address) {
    return null
  }

  return {
    userName: address.userName || '',
    telNumber: address.telNumber || '',
    provinceName: address.provinceName || '',
    cityName: address.cityName || '',
    countyName: address.countyName || '',
    detailInfo: address.detailInfo || ''
  }
}

function formatAddressText(address) {
  if (!address) {
    return ''
  }

  return [
    address.provinceName,
    address.cityName,
    address.countyName,
    address.detailInfo
  ].filter((segment) => typeof segment === 'string' && segment.trim())
    .map((segment) => segment.trim())
    .join('')
}

function formatAddressReceiver(address) {
  if (!address) {
    return ''
  }

  return `${address.userName} ${address.telNumber}`
}

module.exports = {
  normalizeWechatAddress,
  formatAddressText,
  formatAddressReceiver
}
