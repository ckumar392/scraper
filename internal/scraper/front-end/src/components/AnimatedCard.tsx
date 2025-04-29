import React from 'react';
import { Card, CardProps, Box } from '@mui/material';
import { motion } from 'framer-motion';
import { styled } from '@mui/material/styles';
import { colors } from '../theme/theme';

// Create a motion wrapper component
const AnimationWrapper = styled(motion.div)({
  display: 'block',
  width: '100%'
});

// Styled component with glass-like effect for futuristic UI
const GlassCard = styled(Card)(({ theme }) => ({
  background: `rgba(255, 255, 255, 0.85)`,
  backdropFilter: 'blur(10px)',
  border: `1px solid rgba(255, 255, 255, 0.18)`,
  boxShadow: `0 8px 32px 0 rgba(0, 114, 198, 0.1)`,
  borderRadius: '16px',
  padding: theme.spacing(3),
  position: 'relative',
  overflow: 'hidden',
  '&::before': {
    content: '""',
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    height: '4px',
    background: `linear-gradient(90deg, ${colors.primary}, ${colors.accent2})`,
  }
}));

interface AnimatedCardProps extends CardProps {
  delay?: number;
}

const AnimatedCard: React.FC<AnimatedCardProps> = ({ 
  children, 
  delay = 0,
  ...props 
}) => {
  return (
    <AnimationWrapper
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ 
        duration: 0.6,
        delay: delay * 0.1,
        ease: [0.22, 1, 0.36, 1]
      }}
      whileHover={{ 
        y: -5,
        boxShadow: `0 12px 40px 0 rgba(0, 114, 198, 0.2)`,
        transition: { duration: 0.3 }
      }}
    >
      <GlassCard {...props}>
        <Box component="div">
          {children}
        </Box>
      </GlassCard>
    </AnimationWrapper>
  );
};

export default AnimatedCard;