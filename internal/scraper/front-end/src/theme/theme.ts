import { createTheme } from '@mui/material/styles';

// Color palette as specified
const colors = {
  primary: '#0072C6',
  white: '#FFFFFF',
  darkGray: '#333333',
  lightGray: '#F5F5F5',
  mediumGray: '#999999',
  accent1: '#8DC63F',
  accent2: '#3399FF',
};

const theme = createTheme({
  palette: {
    primary: {
      main: colors.primary,
    },
    secondary: {
      main: colors.accent1,
    },
    background: {
      default: colors.lightGray,
      paper: colors.white,
    },
    text: {
      primary: colors.darkGray,
      secondary: colors.mediumGray,
    },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontWeight: 700,
    },
    h2: {
      fontWeight: 600,
    },
    h3: {
      fontWeight: 600,
    },
    button: {
      fontWeight: 500,
      textTransform: 'none',
    },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          padding: '10px 24px',
          transition: 'all 0.3s ease',
          '&:hover': {
            transform: 'translateY(-2px)',
            boxShadow: '0 4px 20px rgba(0, 114, 198, 0.2)',
          },
        },
        containedPrimary: {
          background: `linear-gradient(45deg, ${colors.primary} 30%, ${colors.accent2} 90%)`,
        },
        containedSecondary: {
          background: `linear-gradient(45deg, ${colors.accent1} 30%, ${colors.accent2} 90%)`,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.08)',
          transition: 'transform 0.3s ease, box-shadow 0.3s ease',
          '&:hover': {
            transform: 'translateY(-5px)',
            boxShadow: '0 16px 48px rgba(0, 114, 198, 0.12)',
          },
        },
      },
    },
  },
});

// Export both the theme and colors for use throughout the app
export { theme, colors };