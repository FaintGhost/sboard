export function tableTransitionClass(isTransitioning: boolean): string {
  if (isTransitioning) {
    return "opacity-55 transition-[opacity,transform] duration-[var(--motion-fast)] ease-[var(--ease-out)]"
  }
  return "opacity-100 transition-[opacity,transform] duration-[var(--motion-base)] ease-[var(--ease-out)]"
}
