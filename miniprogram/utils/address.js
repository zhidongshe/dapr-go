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
  ].join('')
}

module.exports = {
  normalizeWechatAddress,
  formatAddressText
}
