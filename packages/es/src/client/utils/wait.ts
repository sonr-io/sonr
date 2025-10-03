/**
 * Synchronously waits for the given number of `milliseconds`.
 */
export async function wait(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}
