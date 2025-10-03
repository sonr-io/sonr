/**
 * Extracts and returns the expected account sequence number from an error. If
 * the error is not related to an account sequence mismatch, `null` is returned.
 */
export function extractExpectedAccountSequence(err: Error): bigint | null {
  const matches = err.message.match(
    // This regex is intentionally kept as strict as possible
    /account sequence mismatch, expected (\d+), got (\d+):/
  );
  if (!matches || matches.length < 3) {
    return null;
  }
  return BigInt(matches[1]);
}
