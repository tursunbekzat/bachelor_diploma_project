import { Link } from 'react-router-dom';
import styles from '../styles/HomePage.module.css';

const HomePage = () => {
  return (
    <div className={styles['home-container']}>
      {/* Header Section */}
      <div className="text-center">
        <h1 className={styles['header-text']}>Добро пожаловать в Bang! Game</h1>
        <p className={styles['sub-text']}>
          Выберите игру и наслаждайтесь динамичным игровым процессом с друзьями!
        </p>

        {/* Button to view games */}
        <Link to="/games" className={styles['button-link']}>
          Просмотреть игры
        </Link>
      </div>
    </div>
  );
};

export default HomePage;
