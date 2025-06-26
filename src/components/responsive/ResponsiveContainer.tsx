import React from 'react';
import { useResponsiveStyles } from '../../hooks/useResponsive';

interface ResponsiveContainerProps {
  children: React.ReactNode;
  className?: string;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full';
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
  centered?: boolean;
  fluid?: boolean;
  as?: keyof JSX.IntrinsicElements;
}

export const ResponsiveContainer: React.FC<ResponsiveContainerProps> = ({
  children,
  className = '',
  maxWidth = 'lg',
  padding = 'md',
  centered = true,
  fluid = false,
  as: Component = 'div',
}) => {
  const { container } = useResponsiveStyles();

  const getMaxWidth = () => {
    if (fluid) return '100%';
    
    const maxWidthMap = {
      sm: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
      full: '100%',
    };
    
    return maxWidthMap[maxWidth];
  };

  const getPadding = () => {
    const paddingMap = {
      none: '0',
      sm: '0.5rem',
      md: '1rem',
      lg: '1.5rem',
      xl: '2rem',
    };
    
    return paddingMap[padding];
  };

  const containerStyles: React.CSSProperties = {
    maxWidth: getMaxWidth(),
    padding: getPadding(),
    margin: centered ? '0 auto' : '0',
    width: '100%',
    boxSizing: 'border-box',
  };

  return (
    <Component
      className={`responsive-container ${className}`}
      style={containerStyles}
    >
      {children}
    </Component>
  );
};

// Specialized container variants
export const PageContainer: React.FC<Omit<ResponsiveContainerProps, 'maxWidth' | 'padding'>> = (props) => (
  <ResponsiveContainer
    {...props}
    maxWidth="xl"
    padding="lg"
    className={`page-container ${props.className || ''}`}
  />
);

export const ContentContainer: React.FC<Omit<ResponsiveContainerProps, 'maxWidth' | 'padding'>> = (props) => (
  <ResponsiveContainer
    {...props}
    maxWidth="lg"
    padding="md"
    className={`content-container ${props.className || ''}`}
  />
);

export const SectionContainer: React.FC<Omit<ResponsiveContainerProps, 'maxWidth' | 'padding'>> = (props) => (
  <ResponsiveContainer
    {...props}
    maxWidth="2xl"
    padding="xl"
    className={`section-container ${props.className || ''}`}
  />
);

export const CardContainer: React.FC<Omit<ResponsiveContainerProps, 'maxWidth' | 'padding'>> = (props) => (
  <ResponsiveContainer
    {...props}
    maxWidth="md"
    padding="sm"
    className={`card-container ${props.className || ''}`}
  />
);

export default ResponsiveContainer; 