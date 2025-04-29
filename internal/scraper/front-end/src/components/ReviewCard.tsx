import React from 'react';
import { 
  Typography, Box, Chip, Rating, 
  Avatar, Stack, Divider 
} from '@mui/material';
import { styled } from '@mui/material/styles';
import { motion } from 'framer-motion';
import { Review } from '../types/reviews';
import AnimatedCard from './AnimatedCard';
import { colors } from '../theme/theme';

const PlatformAvatar = styled(Avatar)<{ platform: string }>(({ platform, theme }) => {
  const platformColors: Record<string, string> = {
    'G2': '#FF492C',
    'App Store': '#0D96F6',
    'Google Play': '#01875F',
    'Trustpilot': '#00B67A',
    'Reddit': '#FF4500',
    'Twitter': '#1DA1F2',
    'Other': colors.mediumGray
  };
  
  return {
    backgroundColor: platformColors[platform] || colors.primary,
    color: '#FFFFFF',
    width: 40,
    height: 40,
    fontWeight: 'bold',
    fontSize: '1rem'
  };
});

const SentimentIndicator = styled(Box)<{ sentiment: string }>(({ sentiment }) => {
  const sentimentColors: Record<string, string> = {
    'Positive': colors.accent1,
    'Neutral': colors.accent2,
    'Negative': '#FF6B6B'
  };
  
  return {
    width: 8,
    height: 8,
    borderRadius: '50%',
    backgroundColor: sentimentColors[sentiment] || colors.mediumGray,
    marginRight: 8,
    boxShadow: `0 0 8px ${sentimentColors[sentiment] || colors.mediumGray}`
  };
});

const HighlightChip = styled(Chip)(({ theme }) => ({
  background: `linear-gradient(135deg, ${colors.accent2} 0%, ${colors.primary} 100%)`,
  color: 'white',
  fontWeight: 500,
  '& .MuiChip-label': {
    textShadow: '0 1px 2px rgba(0,0,0,0.1)'
  }
}));

const TagChip = styled(Chip)(({ theme }) => ({
  backgroundColor: 'rgba(0, 114, 198, 0.08)',
  borderColor: 'rgba(0, 114, 198, 0.2)',
  '&:hover': {
    backgroundColor: 'rgba(0, 114, 198, 0.12)',
  }
}));

interface ReviewCardProps {
  review: Review;
  index: number;
  onClick?: (review: Review) => void;
}

const ReviewCard: React.FC<ReviewCardProps> = ({ review, index, onClick }) => {
  const getPlatformInitials = (platform: string): string => {
    if (platform === 'App Store') return 'AS';
    if (platform === 'Google Play') return 'GP';
    return platform.charAt(0);
  };

  return (
    <AnimatedCard 
      delay={index} 
      onClick={() => onClick && onClick(review)}
      sx={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <PlatformAvatar platform={review.platform}>
            {getPlatformInitials(review.platform)}
          </PlatformAvatar>
          <Box sx={{ ml: 2 }}>
            <Typography variant="subtitle1" fontWeight="bold">
              {review.author || 'Anonymous'}
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <SentimentIndicator sentiment={review.sentiment} />
              <Typography variant="body2" color="text.secondary">
                {review.platform} • {new Date(review.date).toLocaleDateString()}
              </Typography>
            </Box>
          </Box>
        </Box>
        {review.rating && (
          <Rating value={review.rating} readOnly precision={0.5} />
        )}
      </Box>
      
      <Typography variant="h6" gutterBottom>
        {review.title}
      </Typography>
      
      <Typography variant="body1" paragraph>
        {review.content}
      </Typography>

      {review.highlights && review.highlights.length > 0 && (
        <>
          <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
            Highlights
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mb: 2 }}>
            {review.highlights.map((highlight, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 + idx * 0.1 }}
              >
                <HighlightChip 
                  label={highlight}
                  size="small"
                />
              </motion.div>
            ))}
          </Stack>
        </>
      )}
      
      {review.tags && review.tags.length > 0 && (
        <>
          <Divider sx={{ my: 2 }} />
          <Stack direction="row" spacing={1} flexWrap="wrap">
            {review.tags.map((tag, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.5 + idx * 0.05 }}
              >
                <TagChip 
                  label={`#${tag}`}
                  variant="outlined"
                  size="small"
                />
              </motion.div>
            ))}
          </Stack>
        </>
      )}

      <motion.div
        style={{ 
          position: 'absolute', 
          top: 12, 
          right: 12, 
          width: 12, 
          height: 12, 
          borderRadius: '50%',
          backgroundColor: review.isProcessed ? colors.accent1 : colors.mediumGray
        }}
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 0.6, type: "spring" }}
      />
    </AnimatedCard>
  );
};

export default ReviewCard;