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
import { useResponsive, useResponsiveStyles, useResponsiveVisibility } from '../../hooks/useResponsive';
import { styled } from '@mui/material/styles';

interface NavigationItem {
  id: string;
  label: string;
  icon?: React.ReactNode;
  href?: string;
  onClick?: () => void;
  badge?: number;
  disabled?: boolean;
}

interface ResponsiveNavigationProps {
  items: NavigationItem[];
  activeItem?: string;
  onItemClick?: (item: NavigationItem) => void;
  className?: string;
  showLabels?: boolean;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
}

const navItems: NavigationItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <DashboardIcon /> },
  { id: 'jobs', label: 'Jobs', icon: <JobsIcon />, badge: 3 },
  { id: 'marketplace', label: 'Marketplace', icon: <MarketplaceIcon /> },
  { id: 'resources', label: 'Resources', icon: <ResourcesIcon /> },
  { id: 'analytics', label: 'Analytics', icon: <AnalyticsIcon /> },
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

export const ResponsiveNavigation: React.FC<ResponsiveNavigationProps> = ({
  items,
  activeItem,
  onItemClick,
  className = '',
  showLabels = true,
  collapsed = false,
  onToggleCollapse,
}) => {
  const { navigation } = useResponsiveStyles();
  const { showOnMobile, showOnDesktop } = useResponsiveVisibility();
  const [isOpen, setIsOpen] = useState(false);

  const handleItemClick = (item: NavigationItem) => {
    if (item.disabled) return;
    
    if (onItemClick) {
      onItemClick(item);
    }
    
    // Close mobile menu after item click
    if (navigation.type === 'bottom') {
      setIsOpen(false);
    }
  };

  const handleToggle = () => {
    setIsOpen(!isOpen);
    if (onToggleCollapse) {
      onToggleCollapse();
    }
  };

  // Desktop Sidebar Navigation
  const SidebarNavigation = () => (
    <nav className={`sidebar-navigation ${collapsed ? 'collapsed' : ''} ${className}`}>
      <div className="sidebar-header">
        {!collapsed && <h2 className="sidebar-title">ComputeHive</h2>}
        {onToggleCollapse && (
          <button
            className="sidebar-toggle"
            onClick={handleToggle}
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? '→' : '←'}
          </button>
        )}
      </div>
      
      <ul className="sidebar-menu">
        {items.map((item) => (
          <li key={item.id} className="sidebar-item">
            <button
              className={`sidebar-link ${activeItem === item.id ? 'active' : ''} ${item.disabled ? 'disabled' : ''}`}
              onClick={() => handleItemClick(item)}
              disabled={item.disabled}
            >
              {item.icon && <span className="sidebar-icon">{item.icon}</span>}
              {(!collapsed || !showLabels) && (
                <span className="sidebar-label">{item.label}</span>
              )}
              {item.badge && item.badge > 0 && (
                <span className="sidebar-badge">{item.badge}</span>
              )}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );

  // Mobile Bottom Navigation
  const BottomNavigation = () => (
    <nav className={`bottom-navigation ${className}`}>
      <ul className="bottom-menu">
        {items.map((item) => (
          <li key={item.id} className="bottom-item">
            <button
              className={`bottom-link ${activeItem === item.id ? 'active' : ''} ${item.disabled ? 'disabled' : ''}`}
              onClick={() => handleItemClick(item)}
              disabled={item.disabled}
            >
              {item.icon && <span className="bottom-icon">{item.icon}</span>}
              <span className="bottom-label">{item.label}</span>
              {item.badge && item.badge > 0 && (
                <span className="bottom-badge">{item.badge}</span>
              )}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );

  // Mobile Drawer Navigation
  const DrawerNavigation = () => (
    <>
      <button
        className="drawer-toggle"
        onClick={handleToggle}
        aria-label="Toggle navigation menu"
      >
        <span className="hamburger-icon">☰</span>
      </button>
      
      {isOpen && (
        <div className="drawer-overlay" onClick={() => setIsOpen(false)}>
          <nav className="drawer-navigation" onClick={(e) => e.stopPropagation()}>
            <div className="drawer-header">
              <h2 className="drawer-title">ComputeHive</h2>
              <button
                className="drawer-close"
                onClick={() => setIsOpen(false)}
                aria-label="Close navigation"
              >
                ✕
              </button>
            </div>
            
            <ul className="drawer-menu">
              {items.map((item) => (
                <li key={item.id} className="drawer-item">
                  <button
                    className={`drawer-link ${activeItem === item.id ? 'active' : ''} ${item.disabled ? 'disabled' : ''}`}
                    onClick={() => handleItemClick(item)}
                    disabled={item.disabled}
                  >
                    {item.icon && <span className="drawer-icon">{item.icon}</span>}
                    <span className="drawer-label">{item.label}</span>
                    {item.badge && item.badge > 0 && (
                      <span className="drawer-badge">{item.badge}</span>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          </nav>
        </div>
      )}
    </>
  );

  return (
    <>
      {showOnDesktop(<SidebarNavigation />)}
      {showOnMobile(
        navigation.type === 'bottom' ? <BottomNavigation /> : <DrawerNavigation />
      )}
    </>
  );
};

// Navigation wrapper component
export const NavigationWrapper: React.FC<{
  children: React.ReactNode;
  navigation: React.ReactNode;
}> = ({ children, navigation }) => {
  const { navigation: navStyles } = useResponsiveStyles();

  return (
    <div className="navigation-wrapper">
      {navStyles.type === 'sidebar' && (
        <div className="sidebar-layout">
          {navigation}
          <main className="main-content">{children}</main>
        </div>
      )}
      
      {navStyles.type === 'bottom' && (
        <div className="bottom-layout">
          <main className="main-content">{children}</main>
          {navigation}
        </div>
      )}
    </div>
  );
};

export default ResponsiveNavigation; 