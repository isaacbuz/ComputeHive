import React from 'react';
import { Box, BoxProps } from '@mui/material';
import { styled } from '@mui/material/styles';
import { containerWidths, breakpoints } from '../../theme/breakpoints';

interface ResponsiveContainerProps extends BoxProps {
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | false;
  disableGutters?: boolean;
  children: React.ReactNode;
}

const StyledContainer = styled(Box, {
  shouldForwardProp: (prop) => prop !== 'maxWidth' && prop !== 'disableGutters',
})<ResponsiveContainerProps>(({ theme, maxWidth = 'lg', disableGutters = false }) => ({
  width: '100%',
  marginLeft: 'auto',
  marginRight: 'auto',
  paddingLeft: disableGutters ? 0 : theme.spacing(2),
  paddingRight: disableGutters ? 0 : theme.spacing(2),
  
  // Mobile
  [`@media (min-width: ${breakpoints.sm}px)`]: {
    paddingLeft: disableGutters ? 0 : theme.spacing(3),
    paddingRight: disableGutters ? 0 : theme.spacing(3),
  },
  
  // Tablet
  [`@media (min-width: ${breakpoints.md}px)`]: {
    maxWidth: maxWidth && containerWidths.md,
    paddingLeft: disableGutters ? 0 : theme.spacing(4),
    paddingRight: disableGutters ? 0 : theme.spacing(4),
  },
  
  // Desktop
  [`@media (min-width: ${breakpoints.lg}px)`]: {
    maxWidth: maxWidth && (maxWidth === 'md' ? containerWidths.md : containerWidths.lg),
  },
  
  // Desktop large
  [`@media (min-width: ${breakpoints.xl}px)`]: {
    maxWidth: maxWidth && (
      maxWidth === 'md' ? containerWidths.md :
      maxWidth === 'lg' ? containerWidths.lg :
      containerWidths.xl
    ),
  },
}));

export const ResponsiveContainer: React.FC<ResponsiveContainerProps> = ({
  children,
  maxWidth = 'lg',
  disableGutters = false,
  ...props
}) => {
  return (
    <StyledContainer maxWidth={maxWidth} disableGutters={disableGutters} {...props}>
      {children}
    </StyledContainer>
  );
};
