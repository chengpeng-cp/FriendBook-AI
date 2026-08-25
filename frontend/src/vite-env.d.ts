/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 后端 API 基地址（空 = 走 Vite 同源代理） */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
