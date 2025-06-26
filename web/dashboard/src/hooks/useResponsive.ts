import { useMediaQuery, useTheme } from '@mui/material';
import { useEffect, useState } from 'react';
import { breakpoints } from '../theme/breakpoints';

export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl';
export type Device = 'mobile' | 'tablet' | 'desktop';

interface ResponsiveState {
  breakpoint: Breakpoint;
  device: Device;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isTouchDevice: boolean;
  width: number;
  height: number;
}

export const useResponsive = (): ResponsiveState => {
  const theme = useTheme();
  const [dimensions, setDimensions] = useState({
    width: window.innerWidth,
    height: window.innerHeight,
  });

  // Media queries
  const isXs = useMediaQuery(`(max-width: ${breakpoints.sm - 1}px)`);
  const isSm = useMediaQuery(`(min-width: ${breakpoints.sm}px) and (max-width: ${breakpoints.md - 1}px)`);
  const isMd = useMediaQuery(`(min-width: ${breakpoints.md}px) and (max-width: ${breakpoints.lg - 1}px)`);
  const isLg = useMediaQuery(`(min-width: ${breakpoints.lg}px) and (max-width: ${breakpoints.xl - 1}px)`);
  const isXl = useMediaQuery(`(min-width: ${breakpoints.xl}px)`);

  // Device queries
  const isMobile = useMediaQuery(`(max-width: ${breakpoints.md - 1}px)`);
  const isTablet = useMediaQuery(`(min-width: ${breakpoints.md}px) and (max-width: ${breakpoints.lg - 1}px)`);
  const isDesktop = useMediaQuery(`(min-width: ${breakpoints.lg}px)`);

  // Touch detection
  const [isTouchDevice, setIsTouchDevice] = useState(false);

  useEffect(() => {
    const checkTouch = () => {
      setIsTouchDevice(
        'ontouchstart' in window ||
        navigator.maxTouchPoints > 0 ||
        // @ts-ignore
        navigator.msMaxTouchPoints > 0
      );
    };

    checkTouch();

    const handleResize = () => {
      setDimensions({
        width: window.innerWidth,
        height: window.innerHeight,
      });
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Determine current breakpoint
  let breakpoint: Breakpoint = 'xl';
  if (isXs) breakpoint = 'xs';
  else if (isSm) breakpoint = 'sm';
  else if (isMd) breakpoint = 'md';
  else if (isLg) breakpoint = 'lg';

  // Determine device type
  let device: Device = 'desktop';
  if (isMobile) device = 'mobile';
  else if (isTablet) device = 'tablet';

  return {
    breakpoint,
    device,
    isMobile,
    isTablet,
    isDesktop,
    isTouchDevice,
    width: dimensions.width,
    height: dimensions.height,
  };
};

// Hook for conditional rendering based on device
export const useDeviceRender = () => {
  const { device } = useResponsive();
  
  return {
    renderMobile: (component: React.ReactNode) => device === 'mobile' ? component : null,
    renderTablet: (component: React.ReactNode) => device === 'tablet' ? component : null,
    renderDesktop: (component: React.ReactNode) => device === 'desktop' ? component : null,
    renderMobileOrTablet: (component: React.ReactNode) => 
      (device === 'mobile' || device === 'tablet') ? component : null,
    renderTabletOrDesktop: (component: React.ReactNode) => 
      (device === 'tablet' || device === 'desktop') ? component : null,
  };
};
