import type { AppInitialState } from './app';

export default function access(initialState?: AppInitialState) {
  return {
    canViewAdmin: Boolean(initialState?.currentUser),
    isLoggedIn: Boolean(initialState?.currentUser)
  };
}
