export function formatOutput(data: unknown, json: boolean): string {
  if (json) {
    return JSON.stringify(data, null, 2);
  }
  if (typeof data === "string") {
    return data;
  }
  if (Array.isArray(data)) {
    return data.map((item) => (typeof item === "string" ? item : JSON.stringify(item))).join("\n");
  }
  return JSON.stringify(data, null, 2);
}

export function printOutput(data: unknown, json: boolean): void {
  const text = formatOutput(data, json);
  if (text) {
    console.log(text);
  }
}
