function centsToYuan(cents) {
  return (Number(cents || 0) / 100).toFixed(2)
}

function formatPrice(cents) {
  return `¥${centsToYuan(cents)}`
}

module.exports = {
  centsToYuan,
  formatPrice
}
