import "@testing-library/jest-dom/vitest"
import "@/i18n"

// Make React act() warnings behave correctly in tests.
;(globalThis as any).IS_REACT_ACT_ENVIRONMENT = true

// Radix UI (Select/Dialog) uses Pointer Events APIs that jsdom doesn't fully implement.
// Minimal polyfill to keep unit tests stable.
if (!HTMLElement.prototype.hasPointerCapture) {
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  HTMLElement.prototype.setPointerCapture = function () {}
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  HTMLElement.prototype.releasePointerCapture = function () {}
  HTMLElement.prototype.hasPointerCapture = function () {
    return false
  }
}

if (!HTMLElement.prototype.scrollIntoView) {
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  HTMLElement.prototype.scrollIntoView = function () {}
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
    }) as any
}
