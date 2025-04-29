import React from 'react';
import { Button, ButtonProps } from '@mui/material';
import { motion } from 'framer-motion';
import { styled } from '@mui/material/styles';

// Create a styled div that will wrap our Button
const AnimationWrapper = styled(motion.div)({
  display: 'inline-block'
});

// Styled component for the button with additional styles
const StyledButton = styled(Button)(({ theme }) => ({
  position: 'relative',
  overflow: 'hidden',
  '&::after': {
    content: '""',
    position: 'absolute',
    top: 0,
    left: '-100%',
    width: '100%',
    height: '100%',
    background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent)',
    transition: 'all 0.6s ease',
  },
  '&:hover::after': {
    left: '100%',
  }
}));

interface AnimatedButtonProps extends ButtonProps {
  animate?: boolean;
}

const AnimatedButton: React.FC<AnimatedButtonProps> = ({ 
  children, 
  variant = "contained", 
  color = "primary",
  animate = true,
  ...props 
}) => {
  return (
    <AnimationWrapper
      whileHover={animate ? { scale: 1.05 } : undefined}
      whileTap={animate ? { scale: 0.95 } : undefined}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
    >
      <StyledButton
        variant={variant}
        color={color}
        {...props}
      >
        {children}
      </StyledButton>
    </AnimationWrapper>
  );
};

export default AnimatedButton;