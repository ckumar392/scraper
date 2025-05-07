import React, { useEffect, useState } from 'react';
import { Box, Typography, Chip } from '@mui/material';
import Grid from './GridWrapper';
import { styled } from '@mui/material/styles';
import { motion } from 'framer-motion';
import { ReviewStats, Platform, Department } from '../types/reviews';
import AnimatedCard from './AnimatedCard';
import { colors } from '../theme/theme';

// Styled components
const StatsCard = styled(AnimatedCard)(({ theme }) => ({
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'space-between'
}));

const StatValue = styled(motion.div)(({ theme }) => ({
  fontSize: '2.5rem',
  fontWeight: 'bold',
  backgroundImage: `linear-gradient(135deg, ${colors.primary} 0%, ${colors.accent2} 100%)`,
  backgroundClip: 'text',
  WebkitBackgroundClip: 'text',
  color: 'transparent',
  textShadow: '0px 2px 4px rgba(0, 0, 0, 0.1)'
}));

const TrendChip = styled(Chip)<{ trend: 'up' | 'down' | 'stable' }>(({ trend }) => {
  const trendColors = {
    up: colors.accent1,
    down: '#FF6B6B',
    stable: colors.accent2
  };
  
  return {
    backgroundColor: trendColors[trend],
    color: '#FFFFFF',
    fontWeight: 'bold'
  };
});

// Bar chart component
const BarChart = styled(Box)<{ value: number, max: number, color: string }>(
  ({ value, max, color }) => ({
    position: 'relative',
    height: '12px',
    backgroundColor: 'rgba(0, 0, 0, 0.05)',
    borderRadius: '6px',
    overflow: 'hidden',
    '&::after': {
      content: '""',
      position: 'absolute',
      top: 0,
      left: 0,
      height: '100%',
      width: `${Math.min(100, (value / max) * 100)}%`,
      backgroundColor: color,
      borderRadius: '6px'
    }
  })
);

const BarLabel = styled(Box)({
  display: 'flex',
  justifyContent: 'space-between',
  marginBottom: '4px',
  fontSize: '0.85rem'
});

interface StatsDashboardProps {
  stats: ReviewStats;
}

const StatsDashboard: React.FC<StatsDashboardProps> = ({ stats }) => {
  const [counters, setCounters] = useState({
    total: 0,
    positive: 0,
    neutral: 0,
    negative: 0
  });
  
  // Animated counting effect
  useEffect(() => {
    const duration = 1500;
    const frameRate = 30;
    const frames = duration / (1000 / frameRate);
    let frame = 0;
    
    const timer = setInterval(() => {
      frame++;
      const progress = frame / frames;
      
      setCounters({
        total: Math.floor(progress * stats.totalReviews),
        positive: Math.floor(progress * stats.positiveCount),
        neutral: Math.floor(progress * stats.neutralCount),
        negative: Math.floor(progress * stats.negativeCount)
      });
      
      if (frame === frames) {
        clearInterval(timer);
        setCounters({
          total: stats.totalReviews,
          positive: stats.positiveCount,
          neutral: stats.neutralCount,
          negative: stats.negativeCount
        });
      }
    }, 1000 / frameRate);
    
    return () => clearInterval(timer);
  }, [stats]);
  
  // Get the platforms and departments with the highest counts
  const topPlatform = Object.entries(stats.byPlatform)
    .sort((a, b) => b[1] - a[1])[0];
    
  const topDepartment = Object.entries(stats.byDepartment)
    .sort((a, b) => b[1] - a[1])[0];
    
  // Max values for bar charts
  const maxPlatformCount = Math.max(...Object.values(stats.byPlatform));
  const maxDepartmentCount = Math.max(...Object.values(stats.byDepartment));
  
  // Platform colors
  const platformColors: Record<Platform, string> = {
    'G2': '#FF492C',
    'App Store': '#0D96F6',
    'Google Play': '#01875F',
    'Trustpilot': '#00B67A',
    'Reddit': '#FF4500',
    'Twitter': '#1DA1F2',
    'Other': colors.mediumGray
  };
  
  // Department colors
  const departmentColors: Record<Department, string> = {
    'Product': '#8E44AD',
    'Support': '#16A085',
    'Sales': '#F39C12',
    'Marketing': '#E74C3C',
    'Engineering': '#3498DB',
    'General': colors.mediumGray
  };
  
  // Rating stars component
  const ratingStars = () => {
    const rating = stats.averageRating;
    const fullStars = Math.floor(rating);
    const hasHalfStar = rating % 1 >= 0.5;
    
    return (
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        {[...Array(5)].map((_, i) => (
          <Box 
            key={i}
            component={motion.div}
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.5 + i * 0.1 }}
            sx={{ 
              color: i < fullStars ? colors.accent1 : (i === fullStars && hasHalfStar ? colors.accent2 : colors.mediumGray),
              fontSize: '1.8rem', 
              mr: 0.5 
            }}
          >
            {i < fullStars ? '★' : (i === fullStars && hasHalfStar ? '★' : '☆')}
          </Box>
        ))}
        <Typography variant="h6" sx={{ ml: 1 }}>{rating.toFixed(1)}</Typography>
      </Box>
    );
  };

  return (
    <Grid container component="div" spacing={3}>
      {/* Summary stats */}
      <Grid component="div" item xs={12} md={6}>
        <StatsCard>
          <Typography variant="h5" gutterBottom>
            Reviews Summary
          </Typography>
          
          <Box sx={{ my: 2 }}>
            <Typography variant="subtitle2" color="text.secondary">
              Total Reviews
            </Typography>
            <StatValue
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              {counters.total}
            </StatValue>
          </Box>

          <Grid container spacing={2} sx={{ mb: 2 }}>
            <Grid item xs={4}>
              <Typography variant="body2" color="text.secondary">
                Positive
              </Typography>
              <Typography variant="h6" sx={{ color: colors.accent1, fontWeight: 'bold' }}>
                {counters.positive}
              </Typography>
            </Grid>
            <Grid item xs={4}>
              <Typography variant="body2" color="text.secondary">
                Neutral
              </Typography>
              <Typography variant="h6" sx={{ color: colors.accent2, fontWeight: 'bold' }}>
                {counters.neutral}
              </Typography>
            </Grid>
            <Grid item xs={4}>
              <Typography variant="body2" color="text.secondary">
                Negative
              </Typography>
              <Typography variant="h6" sx={{ color: '#FF6B6B', fontWeight: 'bold' }}>
                {counters.negative}
              </Typography>
            </Grid>
          </Grid>

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="subtitle1">Recent Trend</Typography>
            <motion.div
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.4, type: "spring" }}
            >
              <TrendChip 
                label={stats.recentTrend === 'up' ? '↑ Improving' : stats.recentTrend === 'down' ? '↓ Declining' : '→ Stable'} 
                trend={stats.recentTrend}
              />
            </motion.div>
          </Box>
        </StatsCard>
      </Grid>
      
      {/* Average rating */}
      <Grid item xs={12} md={6}>
        <StatsCard>
          <Typography variant="h5" gutterBottom>
            Average Rating
          </Typography>
          
          <Box sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', height: '100%' }}>
            {ratingStars()}
            
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: '100%' }}
              transition={{ delay: 0.8, duration: 1 }}
              style={{ 
                height: 4, 
                background: `linear-gradient(90deg, ${colors.accent1}, ${colors.accent2})`,
                borderRadius: 2,
                marginTop: 16
              }} 
            />
            
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ mt: 1 }}
            >
              Based on {stats.totalReviews} reviews
            </Typography>
          </Box>
        </StatsCard>
      </Grid>
      
      {/* Platform breakdown */}
      <Grid item xs={12} md={6}>
        <StatsCard>
          <Typography variant="h5" gutterBottom>
            Platform Breakdown
          </Typography>
          
          <Typography variant="subtitle2" sx={{ mb: 2 }}>
            Top Platform: <strong>{topPlatform[0]}</strong> ({topPlatform[1]} reviews)
          </Typography>
          
          <Box sx={{ mb: 2 }}>
            {Object.entries(stats.byPlatform).map(([platform, count], index) => (
              <Box key={platform} sx={{ mb: 2 }}>
                <BarLabel>
                  <Typography variant="body2">{platform}</Typography>
                  <Typography variant="body2">{count}</Typography>
                </BarLabel>
                <motion.div
                  initial={{ scaleX: 0, originX: 0 }}
                  animate={{ scaleX: 1 }}
                  transition={{ delay: 0.3 + index * 0.1, duration: 0.8 }}
                >
                  <BarChart 
                    value={count} 
                    max={maxPlatformCount}
                    color={platformColors[platform as Platform] || colors.primary}
                  />
                </motion.div>
              </Box>
            ))}
          </Box>
        </StatsCard>
      </Grid>
      
      {/* Department breakdown */}
      <Grid item xs={12} md={6}>
        <StatsCard>
          <Typography variant="h5" gutterBottom>
            Department Breakdown
          </Typography>
          
          <Typography variant="subtitle2" sx={{ mb: 2 }}>
            Top Department: <strong>{topDepartment[0]}</strong> ({topDepartment[1]} reviews)
          </Typography>
          
          <Box>
            {Object.entries(stats.byDepartment).map(([dept, count], index) => (
              <Box key={dept} sx={{ mb: 2 }}>
                <BarLabel>
                  <Typography variant="body2">{dept}</Typography>
                  <Typography variant="body2">{count}</Typography>
                </BarLabel>
                <motion.div
                  initial={{ scaleX: 0, originX: 0 }}
                  animate={{ scaleX: 1 }}
                  transition={{ delay: 0.3 + index * 0.1, duration: 0.8 }}
                >
                  <BarChart 
                    value={count} 
                    max={maxDepartmentCount}
                    color={departmentColors[dept as Department] || colors.primary}
                  />
                </motion.div>
              </Box>
            ))}
          </Box>
        </StatsCard>
      </Grid>
    </Grid>
  );
};

export default StatsDashboard;