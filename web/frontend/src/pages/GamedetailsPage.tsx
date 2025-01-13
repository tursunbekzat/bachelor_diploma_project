import { useEffect, useState, useContext } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../api';
import { AuthContext } from '../context/AuthContext';
import styles from '../styles/GameDetailsPage.module.css';
import { Link } from 'react-router-dom';
import { AxiosError } from 'axios';


interface Player {
  id: number;
  username: string;
  role: string;
  character: string;
}

interface Game {
  id: number;
  game_name: string;
  creator_id: number;
  status: string;
}

const GameDetailsPage = () => {
  const { id } = useParams<{ id: string }>();
  const authContext = useContext(AuthContext);
  const navigate = useNavigate();

  const [game, setGame] = useState<Game | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isJoined, setIsJoined] = useState(false);

  useEffect(() => {
    const fetchGameDetails = async () => {
      try {
        const response = await api.get(`/api/games/${id}`);
        setGame(response.data.game);
        setPlayers(response.data.players);

        // Проверяем, присоединился ли текущий пользователь
        if (response.data.players.some((player: Player) => player.id === authContext?.userID)) {
          setIsJoined(true);
        }
      } catch (error) {
        console.error('Error fetching game details:', error);
        setError('Failed to load game details. Please try again later.');
      }
    };

    fetchGameDetails();
  }, [id, authContext?.userID]);

  const handleJoinGame = async () => {
    try {
      const response = await api.post('/api/games/join', { game_id: Number(id) });
      alert(response.data.message);
      navigate(`/games/${id}`);
    } catch (error) {
      if (error instanceof AxiosError) {
        if (error.response?.status === 401) {
          alert('Unauthorized! Please login to join the game.');
        } else if (error.response?.status === 404) {
          alert('Game not found! Please check the Game ID and try again.');
        } else if (error.response?.status === 409) {
          alert('You have already joined this game.');
        } else if (error.response?.status === 403) {
          alert('You cannot join your own game.');
        } else {
          alert('An error occurred. Please try again later.');
        }
      } else {
        alert('An unexpected error occurred. Please try again.');
      }
      console.error('Error joining game:', error);
    }
  };

  const handleDeleteGame = async () => {
    try {
      await api.delete(`/api/games/${id}/delete`);
      alert('Game deleted successfully!');
      navigate('/games');
    } catch (error) {
      console.error('Error deleting game:', error);
    }
  };

  const handleStartGame = async () => {
    try {
      const response = await api.post(`/api/games/${id}/start`);
      alert(response.data.message);
      alert('Game started successfully! Roles and characters have been assigned.');
      navigate(`/games/${id}/play`);
    } catch (error) {
      if (error instanceof AxiosError) {
        switch (error.response?.status) {
          case 400:
            alert('Not enough players to start the game. At least 4 players are required.');
            break;
          case 401:
            alert('Unauthorized! Please login to start the game.');
            break;
          case 403:
            alert('Only the game creator can start the game.');
            break;
          case 404:
            alert('Game not found! Please check the Game ID and try again.');
            break;
          default:
            alert('An error occurred while starting the game. Please try again.');
        }
      } else {
        alert('An unexpected error occurred. Please try again.');
      }
      console.error('Error starting game:', error);
    }
  };
  

  if (error) {
    return <p>{error}</p>;
  }

  return (
    <div className={styles.container}>
      <h1>Game Details</h1>
      {game && (
        <div>
          <p>Game Name: {game.game_name}</p>
          <p>Status: {game.status}</p>

          <p><strong>User ID: </strong>{authContext?.userID}</p>
          <p><strong>Game Creator ID: </strong>{game.creator_id}</p>
          {authContext?.userID === game.creator_id ? (
            <div>
              <button onClick={handleStartGame} className={styles.joinButton}>
                Start Game
              </button>
              <button onClick={handleDeleteGame} className={styles.joinButton}>
                Delete Game
              </button>
            </div>
          ) : (
            !isJoined && (
              <button onClick={handleJoinGame} className={styles.joinButton}>
                Join Game
              </button>
            )
          )}

          <h2>Players</h2>
          <ul className={styles.playerList}>
            {players.map((player) => (
              <li key={player.id} className={styles.playerCard}>
                {player.username} - {player.role} - {player.character}
              </li>
            ))}
          </ul>
        </div>
      )}
      <Link to="/games" className={styles.joinButton}>
        Back
      </Link>
    </div>
  );
};

export default GameDetailsPage;
