let counter = 0;

function uniquePrefix(): string {
  counter++;
  return `e2e-${Date.now()}-${counter}`;
}

export function uniqueUsername(): string {
  return `${uniquePrefix()}-user`;
}

export function uniqueGroupName(): string {
  return `${uniquePrefix()}-group`;
}

export function uniqueNodeName(): string {
  return `${uniquePrefix()}-node`;
}

export function uniqueInboundTag(): string {
  return `${uniquePrefix()}-inbound`;
}

export function uniqueClientTag(): string {
  return `${uniquePrefix()}-client`;
}
