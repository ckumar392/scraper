import React, { useState } from 'react';
import { 
  Box, Drawer, List, ListItem, ListItemButton, ListItemIcon, 
  ListItemText, AppBar, Toolbar, IconButton, Typography,
  useTheme, useMediaQuery, Avatar, Divider
} from '@mui/material';
import { styled } from '@mui/material/styles';
import { motion, AnimatePresence } from 'framer-motion';
import { Link, useLocation } from 'react-router-dom';
import MenuIcon from '@mui/icons-material/Menu';
import DashboardIcon from '@mui/icons-material/Dashboard';
import CommentIcon from '@mui/icons-material/Comment';
import AnalyticsIcon from '@mui/icons-material/Analytics';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import { colors } from '../theme/theme';
import ParticleBackground from './ParticleBackground';

// Constants
const DRAWER_WIDTH = 280;
const CLOSED_DRAWER_WIDTH = 80;

// Styled components
const StyledDrawer = styled(Drawer)(({ theme }) => ({
  width: DRAWER_WIDTH,
  flexShrink: 0,
  '& .MuiDrawer-paper': {
    width: DRAWER_WIDTH,
    boxSizing: 'border-box',
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    backdropFilter: 'blur(10px)',
    borderRight: '1px solid rgba(0, 0, 0, 0.05)',
    boxShadow: '0 4px 30px rgba(0, 0, 0, 0.05)',
    transition: theme.transitions.create(['width', 'transform'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
    overflowX: 'hidden',
  },
}));

const MiniDrawer = styled(Drawer)(({ theme }) => ({
  width: CLOSED_DRAWER_WIDTH,
  flexShrink: 0,
  '& .MuiDrawer-paper': {
    width: CLOSED_DRAWER_WIDTH,
    boxSizing: 'border-box',
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    backdropFilter: 'blur(10px)',
    borderRight: '1px solid rgba(0, 0, 0, 0.05)',
    boxShadow: '0 4px 30px rgba(0, 0, 0, 0.05)',
    transition: theme.transitions.create(['width', 'transform'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
    overflowX: 'hidden',
  },
}));

const StyledAppBar = styled(AppBar)(({ theme }) => ({
  backgroundColor: 'rgba(255, 255, 255, 0.9)',
  backdropFilter: 'blur(10px)',
  boxShadow: '0 4px 30px rgba(0, 0, 0, 0.05)',
  borderBottom: '1px solid rgba(0, 0, 0, 0.05)',
  color: colors.darkGray,
}));

const Logo = styled('div')({
  fontWeight: 'bold',
  fontSize: '1.5rem',
  background: `linear-gradient(135deg, ${colors.primary} 0%, ${colors.accent2} 100%)`,
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  display: 'flex',
  alignItems: 'center',
  '& img': {
    height: '28px',
    marginRight: '12px',
  },
});

const MainContent = styled(Box)(({ theme }) => ({
  flexGrow: 1,
  padding: theme.spacing(3),
  minHeight: '100vh',
  background: `radial-gradient(circle at 90% 10%, rgba(141, 198, 63, 0.03) 0%, rgba(0, 0, 0, 0) 50%),
               radial-gradient(circle at 10% 90%, rgba(0, 114, 198, 0.03) 0%, rgba(0, 0, 0, 0) 50%)`,
  backgroundAttachment: 'fixed',
}));

const NavItemText = styled(ListItemText)({
  '& .MuiTypography-root': {
    fontWeight: 500,
  }
});

const ColorBar = styled('div')<{ active: boolean }>(({ active }) => ({
  position: 'absolute',
  left: 0,
  top: 0,
  bottom: 0,
  width: 4,
  background: active ? `linear-gradient(to bottom, ${colors.primary}, ${colors.accent2})` : 'transparent',
  borderRadius: '0 4px 4px 0',
  transition: 'all 0.3s ease',
}));

// Navigation items
const navItems = [
  { text: 'Dashboard', icon: <DashboardIcon />, path: '/' },
  { text: 'Reviews', icon: <CommentIcon />, path: '/reviews' },
  { text: 'Analytics', icon: <AnalyticsIcon />, path: '/analytics' },
  { text: 'Settings', icon: <SettingsIcon />, path: '/settings' },
];

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [open, setOpen] = useState(true);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const location = useLocation();

  // Automatically close drawer on mobile
  React.useEffect(() => {
    if (isMobile) {
      setOpen(false);
    } else {
      setOpen(true);
    }
  }, [isMobile]);

  const handleDrawerToggle = () => {
    setOpen(!open);
  };

  // Determine if a nav item is active
  const isActive = (path: string) => {
    if (path === '/' && location.pathname === '/') return true;
    if (path !== '/' && location.pathname.startsWith(path)) return true;
    return false;
  };

  const drawer = (
    <>
      <Toolbar sx={{ display: 'flex', justifyContent: open ? 'space-between' : 'center', px: 2 }}>
        <AnimatePresence>
          {open && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
            >
              <Logo>
                <span>Review Scraper</span>
              </Logo>
            </motion.div>
          )}
        </AnimatePresence>
        {!isMobile && (
          <IconButton onClick={handleDrawerToggle} edge="end">
            {open ? <MenuIcon /> : <MenuIcon />}
          </IconButton>
        )}
      </Toolbar>
      <Divider />
      <List sx={{ mt: 2 }}>
        {navItems.map((item) => {
          const active = isActive(item.path);
          return (
            <ListItem key={item.text} disablePadding sx={{ mb: 1 }}>
              <ListItemButton
                component={Link}
                to={item.path}
                sx={{
                  px: open ? 3 : 'auto',
                  py: 1.5,
                  justifyContent: open ? 'flex-start' : 'center',
                  position: 'relative',
                  borderRadius: '0 12px 12px 0',
                  mr: 2,
                  bgcolor: active ? 'rgba(0, 114, 198, 0.04)' : 'transparent',
                  '&:hover': {
                    bgcolor: active ? 'rgba(0, 114, 198, 0.08)' : 'rgba(0, 0, 0, 0.04)'
                  }
                }}
              >
                <ColorBar active={active} />
                <ListItemIcon
                  sx={{
                    color: active ? colors.primary : 'inherit',
                    minWidth: open ? 40 : 'auto',
                    mr: open ? 2 : 'auto',
                    justifyContent: 'center',
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                {open && <NavItemText primary={item.text} />}
                {active && open && (
                  <motion.div
                    layoutId="activeIndicator"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                  >
                    <Box
                      sx={{
                        position: 'absolute',
                        right: 16,
                        width: 8,
                        height: 8,
                        borderRadius: '50%',
                        bgcolor: colors.primary
                      }}
                    />
                  </motion.div>
                )}
              </ListItemButton>
            </ListItem>
          );
        })}
      </List>
      <Box sx={{ flexGrow: 1 }} />
      <List sx={{ mt: 'auto' }}>
        <ListItem disablePadding>
          <ListItemButton
            sx={{
              px: open ? 3 : 'auto',
              py: 1.5,
              justifyContent: open ? 'flex-start' : 'center',
            }}
          >
            <ListItemIcon
              sx={{
                minWidth: open ? 40 : 'auto',
                mr: open ? 2 : 'auto',
                justifyContent: 'center',
              }}
            >
              <LogoutIcon />
            </ListItemIcon>
            {open && <NavItemText primary="Logout" />}
          </ListItemButton>
        </ListItem>
      </List>
      <Box
        sx={{
          p: 2,
          display: 'flex',
          alignItems: 'center',
          gap: 2,
          visibility: open ? 'visible' : 'hidden'
        }}
      >
        <Avatar sx={{ bgcolor: colors.primary }}>U</Avatar>
        {open && (
          <Box>
            <Typography variant="subtitle2" noWrap>
              User Name
            </Typography>
            <Typography variant="caption" noWrap color="text.secondary">
              Administrator
            </Typography>
          </Box>
        )}
      </Box>
    </>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      <ParticleBackground />
      <StyledAppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          {isMobile && (
            <IconButton
              color="inherit"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2 }}
            >
              <MenuIcon />
            </IconButton>
          )}
          <Logo>
            {!open || isMobile ? (
              <>
                <span>Review Scraper</span>
              </>
            ) : (
              <span>&nbsp;</span>
            )}
          </Logo>
        </Toolbar>
      </StyledAppBar>
      
      {isMobile ? (
        <Drawer
          variant="temporary"
          open={open}
          onClose={handleDrawerToggle}
          sx={{
            display: { xs: 'block', md: 'none' },
            '& .MuiDrawer-paper': { width: DRAWER_WIDTH },
          }}
        >
          {drawer}
        </Drawer>
      ) : (
        <>
          {open ? (
            <StyledDrawer variant="permanent" open={open}>
              {drawer}
            </StyledDrawer>
          ) : (
            <MiniDrawer variant="permanent" open={!open}>
              {drawer}
            </MiniDrawer>
          )}
        </>
      )}
      
      <MainContent sx={{ 
        flexGrow: 1, 
        ml: { sm: `${open ? DRAWER_WIDTH : CLOSED_DRAWER_WIDTH}px` },
        width: { sm: `calc(100% - ${open ? DRAWER_WIDTH : CLOSED_DRAWER_WIDTH}px)` },
        transition: theme.transitions.create(['margin', 'width'], {
          easing: theme.transitions.easing.easeOut,
          duration: theme.transitions.duration.enteringScreen,
        }),
        pt: { xs: 8, sm: 10 }
      }}>
        <Box component="main">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5 }}
          >
            {children}
          </motion.div>
        </Box>
      </MainContent>
    </Box>
  );
};

export default Layout;