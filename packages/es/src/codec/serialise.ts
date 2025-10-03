import { utf8 } from '@scure/base';
import type { StdSignDoc } from '@sonr.io/es/registry';

/**
 * Escapes <,>,& in string.
 * Golang's json marshaller escapes <,>,& by default.
 * However, because JS doesn't do that by default, to match the sign doc with cosmos-sdk,
 * we should escape <,>,& in string manually.
 * @param str
 */
function escapeHtml(str: string): string {
  return str.replace(/</g, '\\u003c').replace(/>/g, '\\u003e').replace(/&/g, '\\u0026');
}

export function sortObjectByKey<T>(obj: T): T {
  if (typeof obj !== 'object' || obj == null) {
    return obj;
  }
  if (Array.isArray(obj)) {
    return obj.map(sortObjectByKey) as T;
  }
  const sortedKeys = Object.keys(obj).sort();
  const result: Record<string, unknown> = {};
  for (const key of sortedKeys) {
    result[key] = sortObjectByKey((obj as Record<string, unknown>)[key]);
  }
  return result as T;
}

/**
 * Serialises the given sign doc to a `Uint8Array` in a deterministic manner.
 */
export function serialiseSignDoc(doc: StdSignDoc): Uint8Array {
  return utf8.decode(escapeHtml(JSON.stringify(sortObjectByKey(doc))));
}
