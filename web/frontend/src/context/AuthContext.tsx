import { createContext, useState, useEffect, ReactNode } from 'react';
import { jwtDecode } from 'jwt-decode';


// Интерфейс контекста авторизации
interface AuthContextType {
  authTokens: string | null;
  setAuthTokens: (tokens: string | null) => void;
  userID: number | null;
  logout: () => void;
  isAuthenticated: boolean;
}

// Создаем контекст
export const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Провайдер контекста
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [authTokens, setAuthTokens] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);

  // Получаем токен из cookies
  const getTokenFromCookies = () => {
    const cookies = document.cookie.split('; ');
    const tokenCookie = cookies.find((cookie) => cookie.startsWith('token='));
    return tokenCookie ? tokenCookie.split('=')[1] : null;
  };

  // Расшифровка токена
  const decodeToken = (token: string) => {
    try {
      const decodedToken: { user_id: number } = jwtDecode(token);
      return decodedToken.user_id;
    } catch (error) {
      console.error('Error decoding token:', error);
      return null;
    }
  };

  // Проверяем наличие токена в cookies при загрузке приложения
  useEffect(() => {
    const token = getTokenFromCookies();
    if (token) {
      setAuthTokens(token);

      const decodedUserID = decodeToken(token);
      if (decodedUserID) {
        setUserID(decodedUserID);
      }
    }
  }, []);

  // Обновляем токен и cookies при установке новых токенов
  const handleSetAuthTokens = (tokens: string | null) => {
    setAuthTokens(tokens);
    if (tokens) {
      document.cookie = `token=${tokens}; path=/;`;
      const decodedUserID = decodeToken(tokens);
      if (decodedUserID) {
        setUserID(decodedUserID);
      }
    } else {
      document.cookie = 'token=; Max-Age=0; path=/'; // Удаляем cookie
      setUserID(null);
    }
  };

  // Функция выхода
  const logout = () => {
    handleSetAuthTokens(null);
  };

  return (
    <AuthContext.Provider
      value={{
        authTokens,
        setAuthTokens: handleSetAuthTokens,
        userID,
        logout,
        isAuthenticated: !!authTokens,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
