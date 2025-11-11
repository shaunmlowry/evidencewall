/**
 * Decodes HTML entities in a string.
 * @param text - The text containing HTML entities to decode.
 * @returns The decoded text.
 */
export function decodeHtmlEntities(text: string): string {
  const textarea = document.createElement('textarea');
  textarea.innerHTML = text;
  return textarea.value;
}