/**
 * Decodes HTML entities in a string.
 * @param text - The text containing HTML entities to decode.
 * @returns The decoded text.
 */
export function decodeHtmlEntities(text: string): string {
  const parser = new DOMParser();
  const doc = parser.parseFromString(text, 'text/html');
  return doc.documentElement.textContent || '';
}