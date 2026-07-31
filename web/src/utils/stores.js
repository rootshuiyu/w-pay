import http from '../api'

let cache = null
let cacheAt = 0
const TTL = 60_000

/** 加载门店列表（带简单内存缓存，供多页复用） */
export async function loadStoreOptions(force = false) {
  const now = Date.now()
  if (!force && cache && now - cacheAt < TTL) {
    return cache
  }
  const data = await http.get('/admin/store/list', { params: { page: 1, page_size: 500 } })
  const list = data.list || []
  cache = {
    list,
    map: Object.fromEntries(list.map((s) => [s.id, s.store_name])),
    total: data.total || list.length,
  }
  cacheAt = now
  return cache
}

export function storeName(map, id) {
  return map[id] || id
}

export function invalidateStoreCache() {
  cache = null
}
