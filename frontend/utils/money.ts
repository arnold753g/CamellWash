export const normalizePriceInput = (input: string): string | null => {
  const value = input.trim().replace(',', '.')
  if (!/^\d{1,10}(?:\.\d{1,2})?$/.test(value)) return null
  const [rawInteger = '', rawFraction = ''] = value.split('.')
  const integer = rawInteger.replace(/^0+(?=\d)/, '')
  const fraction = rawFraction.padEnd(2, '0')
  if (integer === '0' && fraction === '00') return null
  return `${integer}.${fraction}`
}
