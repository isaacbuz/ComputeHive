import { useState, useEffect, useCallback } from 'react';
import { breakpoints } from '../theme/design-tokens';

export type Breakpoint = keyof typeof breakpoints;
export type DeviceType = 'mobile' | 'tablet' | 'desktop' | 'wide';

interface ResponsiveState {
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isWide: boolean;
  currentBreakpoint: Breakpoint;
  deviceType: DeviceType;
  width: number;
  height: number;
}

const getBreakpoint = (width: number): Breakpoint => {
  if (width >= parseInt(breakpoints['2xl'])) return '2xl';
  if (width >= parseInt(breakpoints.xl)) return 'xl';
  if (width >= parseInt(breakpoints.lg)) return 'lg';
  if (width >= parseInt(breakpoints.md)) return 'md';
  if (width >= parseInt(breakpoints.sm)) return 'sm';
  return 'xs';
};

const getDeviceType = (width: number): DeviceType => {
  if (width >= parseInt(breakpoints.xl)) return 'wide';
  if (width >= parseInt(breakpoints.lg)) return 'desktop';
  if (width >= parseInt(breakpoints.md)) return 'tablet';
  return 'mobile';
};

export const useResponsive = (): ResponsiveState => {
  const [state, setState] = useState<ResponsiveState>(() => {
    const width = typeof window !== 'undefined' ? window.innerWidth : 1024;
    const height = typeof window !== 'undefined' ? window.innerHeight : 768;
    const currentBreakpoint = getBreakpoint(width);
    const deviceType = getDeviceType(width);

    return {
      isMobile: width < parseInt(breakpoints.md),
      isTablet: width >= parseInt(breakpoints.md) && width < parseInt(breakpoints.lg),
      isDesktop: width >= parseInt(breakpoints.lg) && width < parseInt(breakpoints.xl),
      isWide: width >= parseInt(breakpoints.xl),
      currentBreakpoint,
      deviceType,
      width,
      height,
    };
  });

  const updateState = useCallback(() => {
    const width = window.innerWidth;
    const height = window.innerHeight;
    const currentBreakpoint = getBreakpoint(width);
    const deviceType = getDeviceType(width);

    setState({
      isMobile: width < parseInt(breakpoints.md),
      isTablet: width >= parseInt(breakpoints.md) && width < parseInt(breakpoints.lg),
      isDesktop: width >= parseInt(breakpoints.lg) && width < parseInt(breakpoints.xl),
      isWide: width >= parseInt(breakpoints.xl),
      currentBreakpoint,
      deviceType,
      width,
      height,
    });
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    updateState();
    window.addEventListener('resize', updateState);
    window.addEventListener('orientationchange', updateState);

    return () => {
      window.removeEventListener('resize', updateState);
      window.removeEventListener('orientationchange', updateState);
    };
  }, [updateState]);

  return state;
};

// Utility hook for responsive values
export const useResponsiveValue = <T>(
  mobile: T,
  tablet?: T,
  desktop?: T,
  wide?: T
): T => {
  const { isMobile, isTablet, isDesktop, isWide } = useResponsive();

  if (isWide && wide !== undefined) return wide;
  if (isDesktop && desktop !== undefined) return desktop;
  if (isTablet && tablet !== undefined) return tablet;
  return mobile;
};

// Hook for responsive styles
export const useResponsiveStyles = () => {
  const { isMobile, isTablet, isDesktop, isWide } = useResponsive();

  return {
    // Layout utilities
    container: {
      maxWidth: isWide ? '1536px' : isDesktop ? '1280px' : isTablet ? '1024px' : '100%',
      padding: isMobile ? '1rem' : '2rem',
    },
    
    // Grid utilities
    grid: {
      columns: isMobile ? 1 : isTablet ? 2 : isDesktop ? 3 : 4,
      gap: isMobile ? '1rem' : '1.5rem',
    },
    
    // Navigation utilities
    navigation: {
      type: isMobile ? 'bottom' : 'sidebar',
      collapsed: isTablet,
    },
    
    // Data display utilities
    dataDisplay: {
      tableType: isMobile ? 'cards' : 'table',
      pagination: isMobile ? 'infinite' : 'numbered',
    },
    
    // Touch utilities
    touch: {
      targetSize: '44px',
      spacing: isMobile ? '0.5rem' : '1rem',
    },
  };
};

// Hook for responsive breakpoint matching
export const useBreakpoint = (breakpoint: Breakpoint): boolean => {
  const { currentBreakpoint } = useResponsive();
  const breakpointOrder: Breakpoint[] = ['xs', 'sm', 'md', 'lg', 'xl', '2xl'];
  
  const currentIndex = breakpointOrder.indexOf(currentBreakpoint);
  const targetIndex = breakpointOrder.indexOf(breakpoint);
  
  return currentIndex >= targetIndex;
};

// Hook for responsive visibility
export const useResponsiveVisibility = () => {
  const { isMobile, isTablet, isDesktop, isWide } = useResponsive();

  return {
    showOnMobile: (component: React.ReactNode) => isMobile ? component : null,
    showOnTablet: (component: React.ReactNode) => isTablet ? component : null,
    showOnDesktop: (component: React.ReactNode) => isDesktop ? component : null,
    showOnWide: (component: React.ReactNode) => isWide ? component : null,
    hideOnMobile: (component: React.ReactNode) => !isMobile ? component : null,
    hideOnTablet: (component: React.ReactNode) => !isTablet ? component : null,
    hideOnDesktop: (component: React.ReactNode) => !isDesktop ? component : null,
    hideOnWide: (component: React.ReactNode) => !isWide ? component : null,
  };
};

export default useResponsive; 