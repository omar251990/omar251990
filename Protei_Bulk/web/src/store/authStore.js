import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api';

export const useAuthStore = create(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: async (username, password) => {
        try {
          const response = await api.post('/auth/login', { username, password });
          const { access_token, user } = response.data;

          set({
            user,
            token: access_token,
            isAuthenticated: true,
          });

          // Set token in API service
          api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`;

          return { success: true };
        } catch (error) {
          return {
            success: false,
            error: error.response?.data?.detail || 'Login failed',
          };
        }
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        });
        delete api.defaults.headers.common['Authorization'];
      },

      updateUser: (user) => {
        set({ user });
      },
    }),
    {
      name: 'auth-storage',
    }
  )
);
