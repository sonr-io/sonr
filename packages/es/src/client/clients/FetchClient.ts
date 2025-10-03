import type { JsonValue } from '@bufbuild/protobuf';

/**
 * A simple and minimal wrapper around the native `fetch` API.
 */
export class FetchClient {
  /**
   * Performs a GET request to the given `endpoint`, and returns the
   * JSON response.
   */
  public static async get<T>(
    endpoint: string,
    searchParams?: Record<string, string> | undefined
  ): Promise<T> {
    const url = new URL(endpoint);
    url.search = new URLSearchParams(searchParams).toString();
    const res = await fetch(url, { method: 'GET' });
    return res.json();
  }

  /**
   * Performs a POST request to the given `endpoint`, and returns the
   * JSON response.
   */
  public static async post<T>(endpoint: string, body: JsonValue): Promise<T> {
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    return res.json();
  }
}
