import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import HomePage from './pages/HomePage';
import ProfilePage from './pages/ProflePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import NavBar from './components/NavBar';
import AboutPage from './pages/AboutPage';
import GamesPage from './pages/GamesPage';
import CreateGamePage from './pages/CreateGamePage';
import GameDetailsPage from './pages/GamedetailsPage';
import JoinGamePage from './pages/JoinGamePage';


function App() {
  return (
    <Router>
      <AuthProvider>
        <div className="app-container">
          <NavBar />
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/profile" element={<ProfilePage />} />
            <Route path="/games" element={<GamesPage />} />
            <Route path="/games/new" element={<CreateGamePage />} />
            <Route path="/games/:id" element={<GameDetailsPage />} />
            <Route path="/games/join" element={<JoinGamePage />} />
          </Routes>
        </div>
      </AuthProvider>
    </Router>
  );
}

export default App;
