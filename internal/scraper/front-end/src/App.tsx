import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { theme } from './theme/theme';
import Layout from './components/Layout';
import DashboardPage from './pages/DashboardPage';
import ReviewsPage from './pages/ReviewsPage';

// Placeholder components for routes that haven't been created yet
const AnalyticsPage = () => <div>Analytics Page (Coming Soon)</div>;
const SettingsPage = () => <div>Settings Page (Coming Soon)</div>;

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/reviews" element={<ReviewsPage />} />
            <Route path="/analytics" element={<AnalyticsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </Layout>
      </Router>
    </ThemeProvider>
  );
}

export default App;
