type LocaleTree = Record<string, unknown>

function isPlainObject(value: unknown): value is LocaleTree {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function mergeMissingInto(target: LocaleTree, fallback: LocaleTree): LocaleTree {
  for (const [key, value] of Object.entries(fallback)) {
    const current = target[key]
    if (isPlainObject(current) && isPlainObject(value)) {
      mergeMissingInto(current, value)
    } else if (current === undefined) {
      target[key] = value
    }
  }
  return target
}

export function mergeMissingMessages<T extends LocaleTree>(
  messages: T,
  ...fallbacks: LocaleTree[]
): T {
  for (const fallback of fallbacks) {
    mergeMissingInto(messages, fallback)
  }
  return messages
}
