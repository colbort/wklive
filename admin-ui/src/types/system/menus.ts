export type MenuNode = { id: number; name: string; children?: MenuNode[] }
export type PermItem = { key: string; name: string; group?: string } // key 就是 sys:xxx:yyy