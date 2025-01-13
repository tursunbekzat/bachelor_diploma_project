import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../api';
import styles from '../styles/GamesPage.module.css';

interface Game {
    id: number;
    game_name: string;
    creator_name: string;
    created_at: string;
}

const GamesPage = () => {
  const [games, setGames] = useState<Game[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGames = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await api.get('/api/games');
        setGames(response.data);
      } catch (error: unknown) {
        console.error('Error fetching games:', error);
        setError('Failed to load games. Please try again later.');
        setGames([]);
      } finally {
        setIsLoading(false);
      }
    };

    fetchGames();
  }, []);

  if (isLoading) {
    return (
      <div className={styles.container}>
        <h1 className={styles.title}>Active Games</h1>
        <p>Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.container}>
        <h1 className={styles.title}>Active Games</h1>
        <p style={{ color: 'red' }}>{error}</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Active Games</h1>

      <Link to="/games/new" className={styles.createButton}>
        Create New Game
      </Link>
      <Link to="/games/join" className={styles.createButton}>
        Join Game
      </Link>

      {games && games.length === 0 ? (
        <p>
          No games available. <Link to="/create-game">Create a new game</Link>
        </p>
      ) : (
        <div className={styles.gameList}>
          {games?.map((game) => (
            <div key={game.id} className={styles.gameCard}>
              <h2>{game.game_name}</h2>
              <p>Creator: {game.creator_name}</p>
              <p>Created At: {new Date(game.created_at).toLocaleDateString()}</p>
              <Link to={`/games/${game.id}`} className={styles.detailsButton}>
                Details
              </Link>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default GamesPage;
