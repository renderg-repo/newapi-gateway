/**
 * Parse capabilities JSON string to array
 */
export function parseCapabilities(capStr: string | undefined): string[] {
  if (!capStr) return []
  try {
    const parsed = JSON.parse(capStr)
    if (Array.isArray(parsed)) return parsed as string[]
    return []
  } catch {
    return []
  }
}

/**
 * Serialize capabilities array to JSON string
 */
export function serializeCapabilities(caps: string[]): string {
  if (!caps || caps.length === 0) return ''
  return JSON.stringify(caps)
}
