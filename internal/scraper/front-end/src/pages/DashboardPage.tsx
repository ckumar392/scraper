import React, { useState, useEffect } from 'react';
import { 
  Container, Typography, Box,
  Button, CircularProgress
} from '@mui/material';
import { motion } from 'framer-motion';
import { styled } from '@mui/material/styles';
import StatsDashboard from '../components/StatsDashboard';
import AnimatedButton from '../components/AnimatedButton';
import { ReviewStats, Platform } from '../types/reviews';
import { reviewService } from '../services/reviewService';
import { colors } from '../theme/theme';

const HeaderBox = styled(Box)(({ theme }) => ({
  marginBottom: theme.spacing(4),
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center'
}));

const GradientText = styled(Typography)(({ theme }) => ({
  fontWeight: 'bold',
  backgroundImage: `linear-gradient(135deg, ${colors.primary} 0%, ${colors.accent2} 100%)`,
  backgroundClip: 'text',
  WebkitBackgroundClip: 'text',
  color: 'transparent',
}));

const PlatformButton = styled(Button)<{ active: boolean }>(({ active, theme }) => ({
  margin: theme.spacing(0.5),
  borderRadius: '20px',
  transition: 'all 0.3s ease',
  backgroundColor: active ? colors.primary : 'transparent',
  color: active ? '#FFFFFF' : colors.darkGray,
  '&:hover': {
    backgroundColor: active ? colors.primary : 'rgba(0, 114, 198, 0.08)',
    transform: 'translateY(-2px)'
  }
}));

const GlassContainer = styled(Box)(({ theme }) => ({
  padding: theme.spacing(3),
  borderRadius: '16px',
  background: 'rgba(255, 255, 255, 0.8)',
  backdropFilter: 'blur(10px)',
  boxShadow: '0 8px 32px 0 rgba(0, 0, 0, 0.05)',
  marginBottom: theme.spacing(4)
}));

const LoaderContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  height: '400px'
}));

const DashboardPage: React.FC = () => {
  const [stats, setStats] = useState<ReviewStats | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [scrapingInProgress, setScrapingInProgress] = useState<boolean>(false);
  const [selectedPlatforms, setSelectedPlatforms] = useState<Platform[]>([]);
  const [successMessage, setSuccessMessage] = useState<string>('');

  const platforms: Platform[] = [
    'G2', 
    'App Store', 
    'Google Play', 
    'Trustpilot', 
    'Reddit', 
    'Twitter'
  ];

  // Fetch stats on component load
  useEffect(() => {
    const fetchStats = async () => {
      setLoading(true);
      try {
        const data = await reviewService.getStats();
        setStats(data);
      } catch (error) {
        console.error('Error fetching stats:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  // Toggle platform selection for scraping
  const togglePlatform = (platform: Platform) => {
    if (selectedPlatforms.includes(platform)) {
      setSelectedPlatforms(selectedPlatforms.filter(p => p !== platform));
    } else {
      setSelectedPlatforms([...selectedPlatforms, platform]);
    }
  };

  // Start scraping process
  const startScraping = async () => {
    if (selectedPlatforms.length === 0) return;
    
    setScrapingInProgress(true);
    setSuccessMessage('');
    
    try {
      const result = await reviewService.triggerScraping(selectedPlatforms);
      setSuccessMessage(result.message);
      setTimeout(() => {
        setSelectedPlatforms([]);
        setSuccessMessage('');
      }, 5000);
    } catch (error) {
      console.error('Error starting scraping job:', error);
    } finally {
      setScrapingInProgress(false);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <HeaderBox>
          <GradientText variant="h3">
            Dashboard
          </GradientText>
          <Box>
            <AnimatedButton
              variant="contained"
              color="secondary"
              disabled={scrapingInProgress}
              onClick={startScraping}
            >
              {scrapingInProgress ? 'Processing...' : 'Scrape New Reviews'}
            </AnimatedButton>
          </Box>
        </HeaderBox>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1, duration: 0.5 }}
      >
        <GlassContainer>
          <Typography variant="h6" gutterBottom>
            Select Platforms to Scrape
          </Typography>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', my: 1 }}>
            {platforms.map((platform) => (
              <motion.div
                key={platform}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                <PlatformButton
                  variant={selectedPlatforms.includes(platform) ? "contained" : "outlined"}
                  active={selectedPlatforms.includes(platform)}
                  onClick={() => togglePlatform(platform)}
                  disabled={scrapingInProgress}
                >
                  {platform}
                </PlatformButton>
              </motion.div>
            ))}
          </Box>
          {successMessage && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
            >
              <Box 
                sx={{ 
                  mt: 2, 
                  p: 2, 
                  bgcolor: 'rgba(141, 198, 63, 0.1)', 
                  borderRadius: 2,
                  border: `1px solid ${colors.accent1}`
                }}
              >
                <Typography color={colors.accent1}>
                  {successMessage}
                </Typography>
              </Box>
            </motion.div>
          )}
        </GlassContainer>
      </motion.div>

      {loading ? (
        <LoaderContainer>
          <CircularProgress size={60} thickness={4} sx={{ color: colors.accent2 }} />
        </LoaderContainer>
      ) : stats ? (
        <StatsDashboard stats={stats} />
      ) : (
        <Typography variant="h6" color="text.secondary" align="center" sx={{ py: 10 }}>
          No stats available
        </Typography>
      )}
    </Container>
  );
};

export default DashboardPage;