import "@testing-library/jest-dom/vitest";
import "@/i18n";

// Configure the generated API client to use an absolute base URL so that
// `new Request(url)` works in the jsdom/Node.js test environment (relative
// URLs like `/api/users` would throw "Invalid URL" without a host).
import { client } from "@/lib/api/gen/client.gen";
client.setConfig({ baseUrl: "http://localhost/api" });

// Make React act() warnings behave correctly in tests.
(globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;

// Radix UI (Select/Dialog) uses Pointer Events APIs that jsdom doesn't fully implement.
// Minimal polyfill to keep unit tests stable.
if (!HTMLElement.prototype.hasPointerCapture) {
  HTMLElement.prototype.setPointerCapture = function () {
    return undefined;
  };
  HTMLElement.prototype.releasePointerCapture = function () {
    return undefined;
  };
  HTMLElement.prototype.hasPointerCapture = function () {
    return false;
  };
}

if (!HTMLElement.prototype.scrollIntoView) {
  HTMLElement.prototype.scrollIntoView = function () {
    return undefined;
  };
}

// sonner (Toaster) relies on matchMedia.
if (!window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
      addListener: () => {},
      removeListener: () => {},
    }) as any;
}
