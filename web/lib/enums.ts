export function getEnumLabel<T extends string | number>(
  enumLabels: Record<T, string>,
  value: T
): string {
  return enumLabels[value] || String(value)
}

export function getEnumOptions<T extends string | number>(
  enumLabels: Record<T, string>
): Array<{ value: T; label: string }> {
  return Object.entries(enumLabels).map(([value, label]) => ({
    value: value as T,
    label: label as string,
  }))
}
