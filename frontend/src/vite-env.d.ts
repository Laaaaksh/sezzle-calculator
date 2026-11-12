/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Base URL for API requests. Empty or unset means same origin. See .env.example. */
  readonly VITE_API_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
