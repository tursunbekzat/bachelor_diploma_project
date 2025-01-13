import { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import api from '../api';
import styles from '../styles/JoinGamePage.module.css';
import { AxiosError } from 'axios';
import { Link } from 'react-router-dom';


const JoinGamePage = () => {
  const [gameID, setGameID] = useState<number | ''>('');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const authContext = useContext(AuthContext);

  // Проверка на авторизацию
  if (!authContext?.isAuthenticated) {
    return <p>Please log in to join a game.</p>;
  }

  const handleJoinGame = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!gameID) {
      setError('Please enter a valid game ID.');
      return;
    }

    try {
      const response = await api.post('/api/games/join', { game_id: gameID });
      alert(response.data.message);
      navigate(`/games/${gameID}`);
    } catch (error) {
      if (error instanceof AxiosError) {
        if (error.response?.status === 403) {
          setError('Creator cannot join their own game.');
        } else if (error.response?.status === 409) {
          setError('You have already joined this game.');
        } else {
          setError('Failed to join the game. Please try again.');
        }
      } else {
        setError('An unexpected error occurred.');
      }
    }
  };

  return (
    <div className={styles.container}>
      <h1>Join a Game</h1>
      <form onSubmit={handleJoinGame}>
        <input
          type="number"
          placeholder="Enter Game ID"
          value={gameID}
          onChange={(e) => setGameID(Number(e.target.value))}
          className={styles.inputField}
          required
        />
        <button type="submit" className={styles.submitButton}>
          Join Game
        </button>
        {error && <p className={styles.errorText}>{error}</p>}
      </form>
      <Link to="/games" className={styles.submitButton}>
        Back
      </Link>
    </div>
  );
};

export default JoinGamePage;
