import React, { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import api from '../api';
import styles from '../styles/LoginPage.module.css';

const LoginPage = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const authContext = useContext(AuthContext);
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await api.post('/api/login', { username, password });
      if (authContext) {
        authContext.setAuthTokens(response.data.token);
      }
      alert('Login successful!');
      navigate('/profile');
    } catch (error) {
      console.error('Login error:', error);
      alert('Invalid credentials!');
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.formContainer}>
        <h2 className={styles.title}>Welcome Back!</h2>
        <form onSubmit={handleLogin}>
          <input
            type="text"
            placeholder="Username"
            className={styles.inputField}
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <input
            type="password"
            placeholder="Password"
            className={styles.inputField}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <button type="submit" className={styles.submitButton}>
            Login
          </button>
        </form>
        <p className={styles.linkText}>
          Don't have an account? <a href="/register">Register now</a>
        </p>
      </div>
    </div>
  );
};

export default LoginPage;
