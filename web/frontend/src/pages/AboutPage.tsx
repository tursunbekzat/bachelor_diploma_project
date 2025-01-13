import styles from '../styles/AboutPage.module.css';
import { Link } from 'react-router-dom';
import CowboyImage from '../assets/cowboy.png';

const AboutPage = () => {
  return (
    <div className={styles.aboutContainer}>
      <div className={styles.aboutContent}>
        <h1 className={styles.title}>About Bang! Game</h1>
        <p className={styles.description}>
          Bang! Game is a fast-paced, action-packed card game where you and your friends can engage in a Wild West showdown! 🎯
          Choose your roles, strategize your moves, and be the last one standing.
        </p>
        <p className={styles.description}>
          Whether you're the Sheriff, the Outlaw, or a simple Bystander — your fate depends on your decisions! Ready to test your
          luck and skills? 💥
        </p>
        <Link to="/games" className={styles.exploreButton}>
          Explore Games
        </Link>
      </div>

      {/* Animated Image */}
      <div className={styles.imageContainer}>
        <img src={CowboyImage} alt="Cowboy Illustration" className={styles.cowboyImage} />
      </div>
    </div>
  );
};

export default AboutPage;
