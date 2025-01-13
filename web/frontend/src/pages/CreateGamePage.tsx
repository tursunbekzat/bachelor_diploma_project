import { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';
import { AuthContext } from '../context/AuthContext';
import styles from '../styles/CreateGamePage.module.css';
import { Link } from 'react-router-dom';


const CreateGamePage = () => {
  const [game_name, setGameName] = useState('');
  const navigate = useNavigate();
  const authContext = useContext(AuthContext);

  if (!authContext?.isAuthenticated) {
    return <p>Please log in to create a game.</p>;
  }

  const handleCreateGame = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/api/games/new', { game_name });
      alert('Game created successfully!');
      navigate('/games');
    } catch (error) {
      console.error('Error creating game:', error);
      alert('Failed to create game.');
    }
  };

  return (
    <div className={styles.container}>
      <h1>Create a New Game</h1>
      <form onSubmit={handleCreateGame}>
        <input
          type="text"
          placeholder="game_name"
          value={game_name}
          onChange={(e) => setGameName(e.target.value)}
          required
          className={styles.inputField}
        />
        <button type="submit" className={styles.submitButton}>
          Create Game
        </button>
      </form>
      <Link to="/games" className={styles.submitButton}>
        Back
      </Link>
    </div>
  );
};

export default CreateGamePage;
