import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  IconButton,
  Typography,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  BottomNavigation,
  BottomNavigationAction,
  Box,
  Badge,
  Avatar,
  Menu,
  MenuItem,
  useTheme,
  alpha,
} from '@mui/material';
import {
  Menu as MenuIcon,
  Dashboard as DashboardIcon,
  Work as JobsIcon,
  Store as MarketplaceIcon,
  Memory as ResourcesIcon,
  Analytics as AnalyticsIcon,
  Settings as SettingsIcon,
  Notifications as NotificationsIcon,
  AccountCircle as AccountIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useResponsive } from '../../hooks/useResponsive';
import { styled } from '@mui/material/styles';

interface NavItem {
  path: string;
  label: string;
  icon: React.ReactElement;
  badge?: number;
}

const navItems: NavItem[] = [
  { path: '/', label: 'Dashboard', icon: <DashboardIcon /> },
  { path: '/jobs', label: 'Jobs', icon: <JobsIcon />, badge: 3 },
  { path: '/marketplace', label: 'Marketplace', icon: <MarketplaceIcon /> },
  { path: '/resources', label: 'Resources', icon: <ResourcesIcon /> },
  { path: '/analytics', label: 'Analytics', icon: <AnalyticsIcon /> },
];

const StyledAppBar = styled(AppBar)(({ theme }) => ({
  backdropFilter: 'blur(10px)',
  backgroundColor: alpha(theme.palette.background.paper, 0.8),
  borderBottom: `1px solid ${theme.palette.divider}`,
}));

const StyledBottomNavigation = styled(BottomNavigation)(({ theme }) => ({
  position: 'fixed',
  bottom: 0,
  left: 0,
  right: 0,
  borderTop: `1px solid ${theme.palette.divider}`,
  backdropFilter: 'blur(10px)',
  backgroundColor: alpha(theme.palette.background.paper, 0.9),
}));

const DrawerContent = styled(Box)(({ theme }) => ({
  width: 280,
  height: '100%',
  backgroundColor: theme.palette.background.paper,
}));

export const ResponsiveNavigation: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const { isMobile, isTablet, isDesktop } = useResponsive();
  
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [notificationAnchor, setNotificationAnchor] = useState<null | HTMLElement>(null);

  const handleNavigation = (path: string) => {
    navigate(path);
    setDrawerOpen(false);
  };

  const handleProfileMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleProfileMenuClose = () => {
    setAnchorEl(null);
  };

  const handleNotificationOpen = (event: React.MouseEvent<HTMLElement>) => {
    setNotificationAnchor(event.currentTarget);
  };

  const handleNotificationClose = () => {
    setNotificationAnchor(null);
  };

  // Desktop Navigation
  const DesktopNav = () => (
    <StyledAppBar position="fixed" elevation={0}>
      <Toolbar>
        <Typography variant="h6" sx={{ flexGrow: 0, mr: 4, fontWeight: 'bold' }}>
          ComputeHive
        </Typography>
        
        <Box sx={{ flexGrow: 1, display: 'flex', gap: 2 }}>
          {navItems.map((item) => (
            <Box
              key={item.path}
              onClick={() => handleNavigation(item.path)}
              sx={{
                px: 2,
                py: 1,
                borderRadius: 1,
                cursor: 'pointer',
                backgroundColor: location.pathname === item.path 
                  ? alpha(theme.palette.primary.main, 0.1)
                  : 'transparent',
                '&:hover': {
                  backgroundColor: alpha(theme.palette.primary.main, 0.05),
                },
                display: 'flex',
                alignItems: 'center',
                gap: 1,
              }}
            >
              <Badge badgeContent={item.badge} color="error">
                {React.cloneElement(item.icon, { fontSize: 'small' })}
              </Badge>
              <Typography variant="body2">{item.label}</Typography>
            </Box>
          ))}
        </Box>

        <Box sx={{ display: 'flex', gap: 1 }}>
          <IconButton onClick={handleNotificationOpen}>
            <Badge badgeContent={5} color="error">
              <NotificationsIcon />
            </Badge>
          </IconButton>
          
          <IconButton onClick={() => handleNavigation('/settings')}>
            <SettingsIcon />
          </IconButton>
          
          <IconButton onClick={handleProfileMenuOpen}>
            <Avatar sx={{ width: 32, height: 32 }}>U</Avatar>
          </IconButton>
        </Box>
      </Toolbar>
    </StyledAppBar>
  );

  // Tablet Navigation (Hamburger + Drawer)
  const TabletNav = () => (
    <>
      <StyledAppBar position="fixed" elevation={0}>
        <Toolbar>
          <IconButton edge="start" onClick={() => setDrawerOpen(true)}>
            <MenuIcon />
          </IconButton>
          
          <Typography variant="h6" sx={{ flexGrow: 1, ml: 2 }}>
            ComputeHive
          </Typography>
          
          <IconButton onClick={handleNotificationOpen}>
            <Badge badgeContent={5} color="error">
              <NotificationsIcon />
            </Badge>
          </IconButton>
          
          <IconButton onClick={handleProfileMenuOpen}>
            <Avatar sx={{ width: 32, height: 32 }}>U</Avatar>
          </IconButton>
        </Toolbar>
      </StyledAppBar>

      <Drawer
        anchor="left"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        PaperProps={{ sx: { backgroundColor: 'transparent' } }}
      >
        <DrawerContent>
          <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
            <Typography variant="h6" fontWeight="bold">
              ComputeHive
            </Typography>
          </Box>
          
          <List>
            {navItems.map((item) => (
              <ListItem
                button
                key={item.path}
                onClick={() => handleNavigation(item.path)}
                selected={location.pathname === item.path}
                sx={{
                  borderRadius: 1,
                  mx: 1,
                  mb: 0.5,
                }}
              >
                <ListItemIcon>
                  <Badge badgeContent={item.badge} color="error">
                    {item.icon}
                  </Badge>
                </ListItemIcon>
                <ListItemText primary={item.label} />
              </ListItem>
            ))}
            
            <ListItem
              button
              onClick={() => handleNavigation('/settings')}
              sx={{ borderRadius: 1, mx: 1, mb: 0.5 }}
            >
              <ListItemIcon>
                <SettingsIcon />
              </ListItemIcon>
              <ListItemText primary="Settings" />
            </ListItem>
          </List>
        </DrawerContent>
      </Drawer>
    </>
  );

  // Mobile Navigation (Bottom Navigation)
  const MobileNav = () => (
    <>
      <StyledAppBar position="fixed" elevation={0}>
        <Toolbar variant="dense">
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            ComputeHive
          </Typography>
          
          <IconButton size="small" onClick={handleNotificationOpen}>
            <Badge badgeContent={5} color="error">
              <NotificationsIcon fontSize="small" />
            </Badge>
          </IconButton>
          
          <IconButton size="small" onClick={handleProfileMenuOpen}>
            <Avatar sx={{ width: 28, height: 28 }}>U</Avatar>
          </IconButton>
        </Toolbar>
      </StyledAppBar>

      <StyledBottomNavigation
        value={location.pathname}
        onChange={(_, newValue) => handleNavigation(newValue)}
        showLabels
      >
        {navItems.slice(0, 4).map((item) => (
          <BottomNavigationAction
            key={item.path}
            label={item.label}
            value={item.path}
            icon={
              <Badge badgeContent={item.badge} color="error">
                {item.icon}
              </Badge>
            }
          />
        ))}
        <BottomNavigationAction
          label="More"
          value="/more"
          icon={<MenuIcon />}
          onClick={(e) => {
            e.preventDefault();
            setDrawerOpen(true);
          }}
        />
      </StyledBottomNavigation>

      {/* Mobile Drawer for additional items */}
      <Drawer
        anchor="bottom"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        PaperProps={{
          sx: {
            borderTopLeftRadius: 16,
            borderTopRightRadius: 16,
            maxHeight: '80vh',
          },
        }}
      >
        <Box sx={{ p: 2 }}>
          <Box
            sx={{
              width: 40,
              height: 4,
              backgroundColor: 'divider',
              borderRadius: 2,
              mx: 'auto',
              mb: 2,
            }}
          />
          
          <List>
            {[...navItems, { path: '/settings', label: 'Settings', icon: <SettingsIcon /> }].map((item) => (
              <ListItem
                button
                key={item.path}
                onClick={() => handleNavigation(item.path)}
                sx={{ borderRadius: 2, mb: 1 }}
              >
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.label} />
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
    </>
  );

  // Profile Menu
  const ProfileMenu = (
    <Menu
      anchorEl={anchorEl}
      open={Boolean(anchorEl)}
      onClose={handleProfileMenuClose}
      transformOrigin={{ horizontal: 'right', vertical: 'top' }}
      anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
    >
      <MenuItem onClick={() => { handleNavigation('/profile'); handleProfileMenuClose(); }}>
        Profile
      </MenuItem>
      <MenuItem onClick={() => { handleNavigation('/account'); handleProfileMenuClose(); }}>
        My Account
      </MenuItem>
      <MenuItem onClick={() => { handleNavigation('/logout'); handleProfileMenuClose(); }}>
        Logout
      </MenuItem>
    </Menu>
  );

  // Notification Menu
  const NotificationMenu = (
    <Menu
      anchorEl={notificationAnchor}
      open={Boolean(notificationAnchor)}
      onClose={handleNotificationClose}
      transformOrigin={{ horizontal: 'right', vertical: 'top' }}
      anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      PaperProps={{
        sx: { width: 320, maxHeight: 400 },
      }}
    >
      <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
        <Typography variant="h6">Notifications</Typography>
      </Box>
      <MenuItem>
        <Typography variant="body2">Job #1234 completed successfully</Typography>
      </MenuItem>
      <MenuItem>
        <Typography variant="body2">New bid on your marketplace offer</Typography>
      </MenuItem>
      <MenuItem>
        <Typography variant="body2">Resource usage alert: GPU at 95%</Typography>
      </MenuItem>
    </Menu>
  );

  return (
    <>
      {isDesktop && <DesktopNav />}
      {isTablet && <TabletNav />}
      {isMobile && <MobileNav />}
      {ProfileMenu}
      {NotificationMenu}
    </>
  );
};
