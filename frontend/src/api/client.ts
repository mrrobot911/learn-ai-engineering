import type { ApiClient } from './types';
import { mockClient } from './mock';

// The one value that selects the implementation. A real client added later (#27-#30)
// replaces the mock here and satisfies the same type, so screens cannot tell which is active.
export const API_MODE = 'mock';

export const api: ApiClient = mockClient;
