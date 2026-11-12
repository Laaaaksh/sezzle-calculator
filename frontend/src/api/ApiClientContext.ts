import { createContext, useContext } from 'react';
import { defaultApiClient, type ApiClient } from './client';

/**
 * Hands the API client to the hooks. It defaults to the real client, so the app needs no
 * provider; tests wrap the tree in <ApiClientContext value={fake}> instead of passing props.
 * Give the provider a stable value (create the client once): the hooks ignore client identity
 * on purpose, so swapping it after mount does not re-send anything.
 */
export const ApiClientContext = createContext<ApiClient>(defaultApiClient);

export function useApiClient(): ApiClient {
  return useContext(ApiClientContext);
}
