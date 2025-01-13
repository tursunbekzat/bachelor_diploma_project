import { useContext, useEffect, useState } from 'react';
import { AuthContext } from '../context/AuthContext';
import api from '../api';
import styles from '../styles/ProfilePage.module.css';

interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
}

const ProfilePage = () => {
  const authContext = useContext(AuthContext);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        if (!authContext?.authTokens) {
          console.error('No auth tokens available.');
          return;
        }

        const response = await api.get('/api/user', {
          headers: {
            Authorization: `Bearer ${authContext.authTokens}`,
          },
        });
        setUser(response.data);
      } catch (error) {
        console.error('Error fetching user data:', error);
      }
    };

    fetchUser();
  }, [authContext?.authTokens]);

  if (!user) {
    return <p>Loading...</p>;
  }

  const formattedDate = new Date(user.created_at).toLocaleDateString();

  return (
    <div className={styles.profileContainer}>
      <div className={styles.profileCard}>
        <h2 className={styles.welcomeText}>Welcome, {user.username}!</h2>
        <p className={styles.info}>
          <strong>ID:</strong> {user.id}
        </p>
        <p className={styles.info}>
          <strong>Email:</strong> {user.email}
        </p>
        <p className={styles.info}>
          <strong>Account Created At:</strong> {formattedDate}
        </p>
      </div>
    </div>
  );
};

export default ProfilePage;
