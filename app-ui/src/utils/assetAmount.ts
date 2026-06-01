export const DEFAULT_ASSET_DECIMAL_PLACES = 8
const CENT_DECIMAL_PLACES = 2
const MAX_ASSET_DECIMAL_PLACES = 18

type DecimalParts = {
  negative: boolean
  integerPart: string
  fractionPart: string
}

export function normalizeAssetDecimalPlaces(value: unknown) {
  const places = Number(value)
  if (!Number.isFinite(places)) return DEFAULT_ASSET_DECIMAL_PLACES
  return Math.min(MAX_ASSET_DECIMAL_PLACES, Math.max(0, Math.trunc(places)))
}

export function normalizeAssetInputDecimalPlaces(value: unknown) {
  return Math.min(CENT_DECIMAL_PLACES, normalizeAssetDecimalPlaces(value))
}

export function formatAssetMinorAmount(value: unknown, decimalPlaces: unknown) {
  const places = normalizeAssetDecimalPlaces(decimalPlaces)
  const parsed = parseDecimalParts(value)
  if (!parsed) return String(value ?? '')

  return formatDecimalText(scaleMinorToDecimal(parsed), places)
}

export function formatAssetDecimalAmount(value: unknown, decimalPlaces: unknown) {
  const places = normalizeAssetDecimalPlaces(decimalPlaces)
  const parsed = parseDecimalParts(value)
  if (!parsed) return String(value ?? '')

  return formatDecimalText(partsToDecimalText(parsed), places)
}

export function parseAssetDecimalToMinorText(value: unknown, decimalPlaces: unknown) {
  const places = normalizeAssetInputDecimalPlaces(decimalPlaces)
  const text = String(value ?? '').trim()
  if (!/^\d+(\.\d*)?$/.test(text)) return ''

  const [integerPart, fractionPart = ''] = text.split('.')
  if (fractionPart.length > places) return ''

  const normalizedInteger = integerPart.replace(/^0+(?=\d)/, '') || '0'
  const minorText =
    `${normalizedInteger}${fractionPart.padEnd(CENT_DECIMAL_PLACES, '0')}`.replace(/^0+(?=\d)/, '') || '0'

  return minorText
}

export function compareDecimalText(left: unknown, right: unknown) {
  const leftParts = parseDecimalParts(left)
  const rightParts = parseDecimalParts(right)
  if (!leftParts || !rightParts) return 0
  if (leftParts.negative !== rightParts.negative) return leftParts.negative ? -1 : 1

  const magnitude = compareMagnitude(leftParts, rightParts)
  return leftParts.negative ? -magnitude : magnitude
}

function parseDecimalParts(value: unknown): DecimalParts | null {
  const text = String(value ?? '').trim()
  if (!text) {
    return { negative: false, integerPart: '0', fractionPart: '' }
  }
  if (!/^[+-]?\d+(\.\d*)?$/.test(text)) return null

  const negative = text.startsWith('-')
  const unsignedText = text.replace(/^[+-]/, '')
  const [rawIntegerPart, rawFractionPart = ''] = unsignedText.split('.')
  const integerPart = rawIntegerPart.replace(/^0+(?=\d)/, '') || '0'
  const fractionPart = rawFractionPart.replace(/0+$/, '')
  const isZero = integerPart === '0' && !fractionPart

  return {
    negative: negative && !isZero,
    integerPart,
    fractionPart,
  }
}

function scaleMinorToDecimal(value: DecimalParts) {
  const decimalPlaces = CENT_DECIMAL_PLACES

  const digits = `${value.integerPart}${value.fractionPart}`.replace(/^0+(?=\d)/, '') || '0'
  const scale = value.fractionPart.length + decimalPlaces
  const paddedDigits = digits.padStart(scale + 1, '0')
  const integerEndIndex = paddedDigits.length - scale
  const integerPart = paddedDigits.slice(0, integerEndIndex).replace(/^0+(?=\d)/, '') || '0'
  const fractionPart = paddedDigits.slice(integerEndIndex).replace(/0+$/, '')
  const sign = value.negative && (integerPart !== '0' || fractionPart) ? '-' : ''

  return fractionPart ? `${sign}${integerPart}.${fractionPart}` : `${sign}${integerPart}`
}

function partsToDecimalText(value: DecimalParts) {
  const sign = value.negative ? '-' : ''
  return value.fractionPart ? `${sign}${value.integerPart}.${value.fractionPart}` : `${sign}${value.integerPart}`
}

function formatDecimalText(value: string, decimalPlaces: number) {
  const parsed = parseDecimalParts(value)
  if (!parsed) return value

  let integerPart = parsed.integerPart
  let fractionPart = parsed.fractionPart

  if (fractionPart.length > decimalPlaces) {
    const keptFraction = fractionPart.slice(0, decimalPlaces)
    const nextDigit = Number(fractionPart[decimalPlaces] || '0')

    if (nextDigit >= 5) {
      const roundedDigits = incrementDigits(`${integerPart}${keptFraction}` || '0')
      if (decimalPlaces > 0) {
        integerPart = roundedDigits.slice(0, -decimalPlaces) || '0'
        fractionPart = roundedDigits.slice(-decimalPlaces).padStart(decimalPlaces, '0')
      } else {
        integerPart = roundedDigits
        fractionPart = ''
      }
    } else {
      fractionPart = keptFraction
    }
  } else {
    fractionPart = fractionPart.padEnd(decimalPlaces, '0')
  }

  const isZero = integerPart === '0' && (!fractionPart || /^0+$/.test(fractionPart))
  const sign = parsed.negative && !isZero ? '-' : ''
  return decimalPlaces > 0 ? `${sign}${integerPart}.${fractionPart}` : `${sign}${integerPart}`
}

function incrementDigits(value: string) {
  const digits = value.split('')
  let carry = 1

  for (let index = digits.length - 1; index >= 0; index -= 1) {
    const nextValue = Number(digits[index]) + carry
    digits[index] = String(nextValue % 10)
    carry = nextValue >= 10 ? 1 : 0
    if (!carry) break
  }

  if (carry) digits.unshift('1')
  return digits.join('')
}

function compareMagnitude(left: DecimalParts, right: DecimalParts) {
  if (left.integerPart.length !== right.integerPart.length) {
    return left.integerPart.length > right.integerPart.length ? 1 : -1
  }
  if (left.integerPart !== right.integerPart) {
    return left.integerPart > right.integerPart ? 1 : -1
  }

  const fractionLength = Math.max(left.fractionPart.length, right.fractionPart.length)
  const leftFraction = left.fractionPart.padEnd(fractionLength, '0')
  const rightFraction = right.fractionPart.padEnd(fractionLength, '0')
  if (leftFraction === rightFraction) return 0
  return leftFraction > rightFraction ? 1 : -1
}
