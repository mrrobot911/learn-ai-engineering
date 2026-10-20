import { useContext } from 'react';
import { StoreContext } from './context';

export function useStore(): ReturnType<typeof import('./context').StoreContext extends React.Context<infer T> ? NonNullable<T> : never> {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error('useStore must be used within StoreProvider');
  return ctx;
}
