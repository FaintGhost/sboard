export const API_LIST_MAX_LIMIT = 500;

type PageRequest = {
  limit: number;
  offset: number;
};

export async function listAllByPage<T>(
  fetchPage: (params: PageRequest) => Promise<T[]>,
  pageSize = API_LIST_MAX_LIMIT,
): Promise<T[]> {
  const limit = Math.max(1, Math.min(pageSize, API_LIST_MAX_LIMIT));
  const out: T[] = [];
  let offset = 0;

  while (true) {
    const page = await fetchPage({ limit, offset });
    out.push(...page);
    if (page.length < limit) {
      return out;
    }
    offset += page.length;
  }
}
